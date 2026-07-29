# Halora Land — Architecture & Rework Reference

> **Purpose:** A complete map of the current system for the planned rework:
> 1. **Split FE / BE** — backend rewritten in **Go**, frontend stays on Next.js (API routes removed).
> 2. **Throw away Prisma** — move to **raw PostgreSQL** (stdlib `database/sql` + `pgx`, or sqlx/sqlc).
>
> Each section ends with a **Porting to Go / raw-PG** callout. The full consolidated DDL is embedded in §5 and is runnable as-is against a fresh Postgres instance.

---

## 1. Overview

**App:** Halora Land — Indonesian construction cost estimation (RAB = *Rencana Anggaran Biaya*) using PUPR ministry AHSP (*Analisa Harga Satuan Pekerjaan*) standards. Users build a project, add work items (pekerjaan) priced from the AHSP library or manually, and the app rolls up subtotals → overhead → profit → PPN → total.

**Current stack:**
- Next.js 16.2 (App Router) + React 19 + TypeScript
- Prisma ORM 7.9 + `@prisma/adapter-pg` + `pg`
- PostgreSQL (Supabase-hosted in prod, local in dev)
- **Supabase Auth** (live) — sessions in SSR cookies
- Tailwind v4 + shadcn/ui, Three.js (3D roof preview), Chart.js (S-curve)
- Zustand (client state), React Context (`ProjectContext`)
- Zod (validation), bcryptjs (legacy/seed-only — see §7), jose (dead — see §7)
- Sentry (errors), PostHog + Vercel Analytics

**Repo shape:**
```
src/
  app/
    (auth)/         login, register, reset-password pages
    (dashboard)/    31 feature pages (many still on mock data — see §10)
    admin/          admin UI (no middleware protection — see §7)
    api/            35 route handlers (the entire backend today)
    auth/callback   Supabase email-verification / recovery callback
  components/        layout, pekerjaan, proyek, shared, ui (shadcn)
  contexts/          ProjectContext (current project selection)
  hooks/             use-mobile, usePekerjaan
  lib/               16 modules: auth, session, supabase, supabase-auth,
                    prisma, ahsp-parser, snapshot, audit, audit-log,
                    api-utils, rate-limit, schemas, validate, constants, utils, posthog
  stores/            zustand: use-project-store, use-sidebar-store, use-ui-store
  types/index.ts     shared TS interfaces (User, Proyek, Pekerjaan, ...)
  proxy.ts           DEAD — old JWT middleware (see §7)
middleware.ts        live — protects /dashboard, redirects authed users off /login
prisma/
  schema.prisma      source of truth (ahead of the SQL dumps — see §5)
  seed.ts            seeds 3 users, 2 projects, 3 master harga, 2 pekerjaan
  migrations/        one delta SQL (AHSP support)
public/data/ahsp-2026.xlsx   PUPR AHSP 2026 catalog (2.1 MB, the source of truth)
supabase-migration-clean.sql pre-AHSP baseline DDL (Prisma-generated)
```

> **Porting to Go / raw-PG:** The backend is entirely `src/app/api/**` — 35 files, no shared service layer. Each route hand-rolls auth + validation + Prisma calls. A Go rewrite should extract these into a clean `handler → service → repository` layering; the per-route duplication is a smells list, not a spec. Keep `src/types/index.ts` as the contract between FE and BE (regenerate Go structs from it, or share via OpenAPI).

---

## 2. End-to-end Flow

### 2.1 Request lifecycle (today)

```
Browser
  └─ fetch('/api/...')   same-origin, cookies included automatically
     └─ middleware.ts    checks Supabase cookie for /dashboard/* and /login only
        └─ route handler src/app/api/<feature>/route.ts
           ├─ getCurrentUser()  ─→ src/lib/session.ts ─→ getCurrentSupabaseUser()
           │   └─ supabase.auth.getUser()  (server-side JWT verify)
           │   └─ prisma.user.findFirst({ OR:[{supabaseAuthId},{email}] })
           │       (auto-links supabaseAuthId if null)
           ├─ zod validate body  (src/lib/schemas.ts)
           ├─ inline access check:  ADMIN || owner || team member
           ├─ prisma.<model>.<op>(...)
           ├─ optional: logAudit() / createAuditLog()  (non-blocking)
           └─ return JSON
```

There is **no central `requireAuth`/`requireRole` helper** — every route repeats the 401 check and the ownership/team check inline. The access pattern is uniform:
```ts
const hasAccess =
  session.role === 'ADMIN' ||
  proyek.userId === session.userId ||
  proyek.timProyek.some(t => t.userId === session.userId)
```
Edit access adds `t.role !== 'viewer'`; delete (pekerjaan) requires `t.role === 'owner'`.

### 2.2 Auth flow (Supabase, live)

1. **Register** — `POST /api/auth/register` → `supabase.auth.signUp()` (Supabase sends verification email) → create local `users` row with `password:''`, `role:'USER'`, `supabaseAuthId:<supabase user id>`.
2. **Verify** — email link → `/auth/callback?code=...` → `exchangeCodeForSession` → redirect `/login?verified=true`.
3. **Login** — `POST /api/auth/login` → `supabase.auth.signInWithPassword()` → find-or-create local user → **manually set cookie** `sb-<projectRef>-auth-token` (httpOnly, secure, sameSite=strict, maxAge=expires_in).
4. **Session read** — `getCurrentSupabaseUser()` builds a server Supabase client from `next/headers` cookies → `supabase.auth.getUser()` (verifies JWT server-side) → Prisma lookup by `supabaseAuthId` or `email`.
5. **Logout** — `POST /api/auth/logout` → `supabase.auth.signOut()` (clears cookie via Supabase client — **potential bug: deletion attributes don't match the manual set, see §7**).
6. **Password reset** — no API endpoint triggers `resetPasswordForEmail`; the "Lupa password?" link is `href="#"`. Recovery only via Supabase Studio or a future endpoint. The reset page `POST /api/auth/update-password` calls `supabase.auth.updateUser({password})`.
7. **Demo login** — login page hardcodes `demo@haloraland.id` / `password123`. The demo Supabase user must be provisioned separately (seed only creates the local row).

### 2.3 The core flow: add an AHSP item to a RAB

This is the most important business flow to preserve in Go.

```
User picks a MasterAnalisa (AHSP catalog item) in the UI
  └─ POST /api/proyek/[id]/pekerjaan/from-ahsp
     body: { masterAnalisaId, volume, applyBreakdown? }
     ├─ auth + edit-access check (t.role !== 'viewer')
     ├─ load MasterAnalisa with rincianAnalisa.komponen
     ├─ Pekerjaan.hargaSatuan ← MasterAnalisa.hargaSatuan   (top-down, from Excel col 8)
     ├─ Pekerjaan.totalBiaya  ← hargaSatuan × volume
     ├─ for each rincianAnalisa:
     │     DetailAnalisa ← COPY { nama, satuan, koef, hargaSatuan (frozen),
     │                            totalBiaya = rincian.jumlahHarga × volume,
     │                            masterHargaId (often null), masterAnalisaId,
     │                            snapshotAt=now, sourceKode=kodeReferensi }
     └─ return created Pekerjaan (with DetailAnalisa)
```

**Key point:** the breakdown is *copied* (frozen) into `DetailAnalisa`. Later price changes in `MasterHarga`/`MasterAnalisa` do **not** retroactively change this Pekerjaan. Drift is detectable via the nullable `masterHargaId` FK (see §3.1).

### 2.4 Recalculate / drift-detection flow

- `GET /api/pekerjaan/[id]/validate-snapshot` — compares each `DetailAnalisa.hargaSatuan` (frozen) against current `MasterHarga.harga` (live) via `masterHargaId`. Returns `{ isValid, changes[] }` with per-component diff + percent change. Gracefully falls back to frozen values when `masterHargaId` is null or the master was deleted (`onDelete: SetNull`).
- `POST /api/pekerjaan/[id]/recalculate` — single item: deletes old `DetailAnalisa`, re-snapshots from current master prices, updates `Pekerjaan.totalBiaya`. Writes an `audit_log` row (`action:'recalculate'`, oldValue/newValue JSON).
- `POST /api/proyek/[id]/recalculate-all` — bulk: all `metodeHitung='ahsp'` pekerjaan in one transaction, one `audit_log` per item (`action:'bulk_recalculate'`).

> **Porting to Go / raw-PG:** The snapshot copy is a `BEGIN; INSERT pekerjaan; INSERT detail_analisa (...) SELECT ... FROM rincian_analisa JOIN master_harga ...; COMMIT;` block. The drift check is a single `SELECT ... LEFT JOIN master_harga ON detail_analisa.master_harga_id = master_harga.id WHERE detail_analisa.harga_satuan IS DISTINCT FROM master_harga.harga`. The "manual cookie set" in login becomes irrelevant — Go issues its own session (JWT in a cookie, or hands the Supabase access_token to the FE). Decide early: keep Supabase Auth (Go verifies the Supabase JWT) or move auth fully into Go (bcrypt + self-issued JWT). **Recommendation:** keep Supabase Auth for the rewrite — Go just verifies the Supabase JWT with `jose`/`golang-jwt` using Supabase's JWKS. This eliminates the `users.password` column entirely.

---

## 3. Key Architectural Decisions

### 3.1 Snapshot pricing (THE central decision)

`DetailAnalisa` stores **frozen copies** of every value at the moment a Pekerjaan is created:
```prisma
model DetailAnalisa {
  masterHargaId   Int?      // nullable FK — kept only for drift detection
  masterAnalisaId Int?      // nullable FK — kept for lineage
  nama            String    // copied
  satuan          String    // copied
  koef            Decimal   // copied
  hargaSatuan     Decimal   // FROZEN — source of truth for this Pekerjaan
  totalBiaya      Decimal   // FROZEN
  tipe            TipeKomponen
  snapshotAt      DateTime
  sourceKode      String?
}
```
`masterHargaId` / `masterAnalisaId` are `onDelete: SetNull` — even if a master item is deleted, the snapshot survives intact. The FKs exist *only* for audit/drift linkage.

**Rationale:** a RAB is a contractual budget document. Once a contractor prices a job, the quoted price must not silently shift when the master material catalog is updated. Yet you still want to *detect* drift so you can re-quote.

> **Porting to Go / raw-PG:** Keep this exactly. The nullable-FK-with-snapshot pattern is pure SQL — no ORM needed. Use `ON DELETE SET NULL` on the FKs. For the drift query, `detail_analisa.harga_satuan IS DISTINCT FROM master_harga.harga` is the canonical check. **Do not** be tempted to "normalize" by dropping the snapshot columns and JOINing live — you will lose budget immutability.

### 3.2 Two-tier breakdown

Two parallel breakdown tables with different lifecycles:

| Table | Lives on | Role | Mutability |
|---|---|---|---|
| `rincian_analisa` | `MasterAnalisa` (catalog) | reusable template breakdown of an AHSP library item | editable; prices evolve |
| `detail_analisa` | `Pekerjaan` (project instance) | per-Pekerjaan frozen snapshot | immutable once created |

`from-ahsp` copies rows from `rincian_analisa` → `detail_analisa`. The `rincian_analisa` rows imported from Excel have `komponen_id = NULL` and store prices inline (snapshot-style); user-created rincian rows may link to `master_harga` via `komponen_id`.

> **Porting to Go / raw-PG:** Two tables, same shape. The import path writes `rincian_analisa` with `komponen_id = NULL` and inline `nama/satuan/harga_satuan/jumlah_harga/kode_referensi/urutan`. The `from-ahsp` path is `INSERT INTO detail_analisa SELECT ... FROM rincian_analisa WHERE master_analisa_id = $1` (scale `total_biaya` by volume in the SELECT).

### 3.3 Auth: Supabase live, `jose`/`bcrypt` dead

- **Live:** `src/lib/supabase-auth.ts` — all auth routes use Supabase. Local `users.password` is set to `''` on register and on the login migration path; the column is vestigial.
- **Dead:** `src/lib/auth.ts` — full `jose` JWT + `bcryptjs` implementation, but only referenced by `src/proxy.ts` (not registered as middleware — Next.js uses `middleware.ts`) and unit tests. The `auth-token` cookie it reads is **never set anywhere**. `JWT_SECRET` is still required at module load — a latent footgun if anything imports it transitively.
- **Seed artifact:** `prisma/seed.ts` still stores bcrypt hashes for the 3 seeded users, but login never calls `verifyPassword` — Supabase holds the real credentials.

> **Porting to Go / raw-PG:** Delete `src/lib/auth.ts`, `src/proxy.ts`, the `bcryptjs` dep, and the `password` column (or keep it nullable and unused). In Go, verify the Supabase JWT with `golang-jwt` using Supabase's JWKS URL (`https://<project>.supabase.co/auth/v1/.well-known/jwks.json`). The Go user lookup is `SELECT id, email, role, ... FROM users WHERE supabase_auth_id = $1 OR email = $2` — same as today, with the auto-link UPDATE in the same transaction.

### 3.4 Excel as AHSP source of truth

`public/data/ahsp-2026.xlsx` (2.1 MB, 1548 work items, 41 detail sheets) is the canonical PUPR AHSP 2026 catalog. `src/lib/ahsp-parser.ts` parses it with `xlsx` (SheetJS):

- **Sheet → kategori** mapping is hardcoded (`SHEET_TO_KATEGORI`, ahsp-parser.ts:30-50). First 2 sheets skipped; unknown → `'custom'`.
- **Work item detection:** column 2 = kode (e.g. `2.2.1.1.1`), column 3 = nama, column 8 = hargaSatuan. Valid kode has ≥4 numeric dot-separated levels.
- **Breakdown extraction:** scans up to 100 rows after a work item; section markers in column 2: `A`→upah, `B`→material, `C`→alat, `D/E/F`→end. Column indices for the breakdown row have defensive fallbacks (`row[4] || row[5]`, etc.) because the Excel format isn't perfectly uniform.
- `biayaUmum` hardcoded `0.10` for every parsed item.
- `extractSatuan` regex-matches common units (`m|m2|m3|kg|ton|unit|buah|ls|liter|OH|hari|jam`) from column 4 or the nama string.

Import route: `POST /api/admin/ahsp/import` (ADMIN-only) — reads the xlsx, transactionally inserts `MasterAnalisa` + `RincianAnalisa` rows (`isSystem=true`, `isGlobal=true`, `komponenId=null`, inline snapshot fields). Supports `forceReimport` (delete + reimport per sheet). Skips already-imported kodes.

> **Porting to Go / raw-PG:** Use `xuri/excelize` for parsing — same column-index logic, port `SHEET_TO_KATEGORI` and the breakdown-section markers verbatim. The import becomes a Go batch INSERT inside a transaction. Consider making the import a CLI command (`./halora-be import-ahsp --sheet Beton --force`) instead of an HTTP endpoint, since the xlsx is a server-side file anyway. Keep the `is_system` / `ahsp_kode` / `ahsp_sheet` columns for idempotency checks.

### 3.5 Hardcoded 10% overhead / 11% PPN; `biayaUmum` stored but unwired

The RAB rollup formula (used in `rab/page.tsx`, `rekap/page.tsx`, `utils.ts:calculateRAB`):
```
subtotal = Σ Pekerjaan.totalBiaya
overhead = subtotal × 0.10         // hardcoded 10%
profit   = (subtotal + overhead) × (margin/100)
ppn      = (subtotal + overhead + profit) × 0.11   // hardcoded 11%
total    = subtotal + overhead + profit + ppn
```
`MasterAnalisa.biayaUmum` (DECIMAL(5,4), default 0.10) is stored per-item and returned by the search API, **but the rollup does NOT read it** — overhead is applied as a flat 10% at the subtotal level. The `rekap/route.ts` even has `overhead = 0 // TODO: make configurable`. `biayaUmum` is infrastructure for future per-item overhead.

> **Porting to Go / raw-PG:** Keep the rollup in a Go function with configurable rates (don't hardcode 0.10/0.11 — read from a `settings` table or env). Decide whether `biayaUmum` stays per-item (future) or moves to project-level. The current frontend duplicates the formula in 3 places — centralize it in the BE and have the FE just display BE-computed totals.

### 3.6 Non-blocking audit logging (two modules, one table)

Two parallel audit modules write to the same `audit_log` table:

- **`src/lib/audit.ts`** (rich) — `createAuditLog({ action, entityType, entityId, oldValue, newValue, ... })` stores full before/after JSON in `oldValue`/`newValue` JSONB columns. Has query helpers (`getAuditTrail`, `getProjectAuditLogs`, `getAuditLogsByAction`, `getUserAuditLogs`).
- **`src/lib/audit-log.ts`** (lightweight, typed) — `logAudit({ action, resource, resourceId, metadata })` serializes `metadata` into the `description` string; leaves `oldValue`/`newValue` null. **Non-blocking by design** — errors are swallowed (`console.error`) so audit never breaks user flows. Typed enums: `AuditAction = CREATE|UPDATE|DELETE|LOGIN|LOGOUT|REGISTER|EXPORT`, `AuditResource = USER|PROYEK|PEKERJAAN|...`.

**Only 4 routes actually write audit logs:** auth/login (`LOGIN`), auth/register (`REGISTER`), pekerjaan/recalculate (`recalculate`), proyek/recalculate-all (`bulk_recalculate`). Most CRUD does **not** audit.

> **Porting to Go / raw-PG:** Collapse into one Go package `internal/audit` with one `Log(ctx, action, entityType, entityID, old, new, ...)` function. Run the INSERT in a goroutine with a buffered channel + background worker (true non-blocking, survives request cancellation). Use `JSONB` columns and `json.Marshal` for old/new. Wire it into **all** mutations, not just recalculate — the current coverage is a gap, not a feature.

### 3.7 No rounding — DB `DECIMAL` owns precision

There is **no `Math.round`/`toFixed` anywhere** in the calculation paths. Precision is delegated to:
- `DECIMAL(15,2)` — prices, totals (rupiah + sen)
- `DECIMAL(10,6)` — coefficients (e.g. `0.012500 zak/m3`)
- `DECIMAL(5,4)` — percentages (`biayaUmum`)
- `DECIMAL(5,2)` — `Rekap.margin`

Display-only rounding via `Intl.NumberFormat("id-ID", { currency:"IDR", maximumFractionDigits:0 })` in `utils.ts:formatCurrency`. `parseCurrency` strips non-digits.

> **Porting to Go / raw-PG:** Use `github.com/shopspring/decimal` (or `pgtype.Numeric` from `pgx`) for all money/coef fields — **do not use `float64`**. Scan DECIMAL columns into `decimal.Decimal`. The rollup function must operate on decimals end-to-end. Format with `message/printer` or a simple `%.0f` for IDR display.

### 3.8 In-memory rate limiting

`src/lib/rate-limit.ts` — fixed-window, `Map`-based, process-local. Cleanup every 5 min. Applied only to auth endpoints:

| Route | Key | Limit | Window |
|---|---|---|---|
| POST `/api/auth/register` | IP | 5 | 15 min |
| POST `/api/auth/login` | IP | 10 | 15 min |
| POST `/api/auth/resend-verification` | email | 3 | 15 min |

Returns `429` with `Retry-After`, `X-RateLimit-*` headers. The file's own comment notes this won't work behind multiple serverless instances.

> **Porting to Go / raw-PG:** Use Redis (`redis_rate` or `go-redis` rate limiter) or Postgres-backed sliding window. Apply to all auth + write endpoints, not just the three above. `getClientIp` reads `X-Forwarded-For` first hop — keep that logic for prod behind a proxy.

### 3.9 Role system gaps

- **Roles:** `ADMIN`, `OWNER`, `USER`, `DEMO` (Prisma enum). `OWNER` is **never assigned and never checked** — reserved/placeholder. `DEMO` is seeded but `isDemo`/`DEMO` is **surfaced in API responses but never enforced** (no "demo can't delete" logic).
- **Enforcement is inline** — no central helper; every route repeats the 401 + ownership/team check.
- **`/admin/*` pages have NO middleware or layout-level protection** — only API routes enforce ADMIN. The admin users page even renders from mock data.
- **No role-based redirects** in middleware (which only checks session presence for `/dashboard/*`).

> **Porting to Go / raw-PG:** Build a real middleware chain in Go: `AuthMiddleware` (verify JWT → load user) → `RoleMiddleware(roles...)` → handler. Delete the `OWNER` role or actually wire it (project owner ≠ `OWNER` role — `tim_proyek.role='owner'` is the real project-owner concept; `User.role` is global). Decide whether `DEMO` is a real read-only role or remove it. Protect admin **routes** at the middleware layer, not just handlers.

---

## 4. Data Models

14 tables, 8 enums. All PKs are `SERIAL` (autoincrement int). Timestamps are `TIMESTAMP(3)` (ms precision). See §5 for full DDL.

### 4.1 Enum types
```sql
Role               := 'ADMIN' | 'OWNER' | 'USER' | 'DEMO'
TipeProyek         := 'gedung' | 'infra'
RoleTimProyek      := 'owner' | 'editor' | 'viewer'
KategoriPekerjaan  := 'persiapan'|'pondasi'|'beton'|'kanopi'|'baja'|'tangga'|
                      'atap'|'dinding'|'plesteran'|'acian'|'keramik'|'paving'|
                      'pengecatan'|'pintu'|'interior'|'toilet'|'mep'|'custom'
MetodeHitung       := 'ahsp'|'manual'|'harga_borong'|'harga_manual'|'harga_custom'
TipeKomponen       := 'material' | 'upah' | 'alat'
StatusInvoice      := 'draft' | 'sent' | 'paid'
StatusFeedback     := 'open' | 'in_progress' | 'resolved' | 'closed'
```

### 4.2 Tables (fields, types, constraints)

**`users`** — app-level profile; identity lives in Supabase.
- `id` SERIAL PK, `namaLengkap` TEXT, `email` TEXT UNIQUE, `password` TEXT (vestigial, `''`), `role` Role default 'USER', `accountType` TEXT default 'free', `isDemo` BOOL default false, `supabaseAuthId` TEXT UNIQUE, `createdAt`, `updatedAt`.
- Relations out: `proyek`, `timProyek`, `masterAnalisa`, `masterHarga`, `feedback`, `feedbackReply`, `auditLogs`.

**`proyek`** — a construction project owned by a user.
- `id` PK, `userId` INT → users(CASCADE), `namaProyek` TEXT, `lokasi` TEXT?, `tipe` TipeProyek default 'gedung', `nilaiKontrak` DECIMAL(15,2)?, `timeline` TEXT?, timestamps.
- Relations: `timProyek`, `pekerjaan`, `rekap`, `invoice`, `logistik`, `realisasi`, `auditLogs`.

**`tim_proyek`** — project team membership (M:N user↔proyek).
- `id` PK, `proyekId` → proyek(CASCADE), `userId` → users(CASCADE), `role` RoleTimProyek, timestamps.
- UNIQUE(`proyekId`,`userId`). Indexes on both FKs.

**`pekerjaan`** — a work item in a project (a line in the RAB).
- `id` PK, `proyekId` → proyek(CASCADE), `kategori` KategoriPekerjaan, `uraianPekerjaan` TEXT, `volume` DECIMAL(15,4), `satuan` TEXT, `hargaSatuan` DECIMAL(15,2), `totalBiaya` DECIMAL(15,2), `metodeHitung` MetodeHitung, `levelPekerjaan` TEXT?, `tipePekerjaan` TEXT?, timestamps.
- Indexes on `proyekId`, `kategori`.
- `totalBiaya = hargaSatuan × volume` (when from-ahsp: `hargaSatuan` ← `MasterAnalisa.hargaSatuan`).

**`detail_analisa`** — frozen breakdown of a pekerjaan (snapshot, §3.1).
- `id` PK, `pekerjaanId` → pekerjaan(CASCADE), `masterHargaId` INT? → master_harga(SET NULL), `masterAnalisaId` INT? → master_analisa(SET NULL), `nama` TEXT, `satuan` TEXT, `koef` DECIMAL(10,6), `hargaSatuan` DECIMAL(15,2), `totalBiaya` DECIMAL(15,2), `tipe` TipeKomponen, `snapshotAt` TIMESTAMP(3), `sourceKode` TEXT?.
- Indexes on `pekerjaanId`, `masterHargaId`, `masterAnalisaId`.

**`master_analisa`** — AHSP catalog item; self-referential hierarchy (5 levels: 0=root, 1=divisi, 2=kelompok, 3=sub, 4=item).
- `id` PK, `kode` TEXT, `nama` TEXT, `level` INT, `parentId` INT? → master_analisa(CASCADE) [self-rel "MasterAnalisaHierarchy"], `satuan` TEXT?, `hargaSatuan` DECIMAL(15,2)?, `kategori` TEXT?, `isGlobal` BOOL default false, `userId` INT? → users(SET NULL), `isSystem` BOOL default false, `ahspKode` TEXT?, `ahspSheet` TEXT?, `biayaUmum` DECIMAL(5,4) default 0.10, timestamps.
- UNIQUE(`kode`,`userId`). Indexes: `userId`, `parentId`, `(kategori,isSystem)`, `ahspKode`, `ahspSheet`, trgm GIN on `nama` and `ahspKode`.

**`rincian_analisa`** — template breakdown of a master_analisa item (§3.2).
- `id` PK, `masterAnalisaId` → master_analisa(CASCADE), `komponenId` INT? → master_harga(CASCADE) [nullable for snapshot-only], `koef` DECIMAL(10,6), `tipe` TipeKomponen, snapshot fields: `nama` TEXT?, `satuan` TEXT?, `hargaSatuan` DECIMAL(15,2)?, `jumlahHarga` DECIMAL(15,2)?, `kodeReferensi` TEXT?, `urutan` INT default 0, timestamps.
- Indexes on `masterAnalisaId`, `komponenId`.

**`master_harga`** — live price catalog (materials, labor, equipment).
- `id` PK, `nama` TEXT, `satuan` TEXT, `harga` DECIMAL(15,2), `kategori` TipeKomponen, `isGlobal` BOOL default false, `userId` INT? → users(SET NULL), `kodeAHSP` TEXT?, `isSystem` BOOL default false, timestamps.
- UNIQUE(`nama`,`userId`,`kategori`). Indexes: `userId`, `kodeAHSP`.

**`rekap`** — RAB summary settings per project (margin etc.).
- `id` PK, `proyekId` → proyek(CASCADE), `kategori` TEXT, `uraian` TEXT, `urutan` INT, `margin` DECIMAL(5,2)?, timestamps.
- Index on `proyekId`. The `rekap/route.ts` uses a row with `kategori='settings'` to store the project's `margin`.

**`invoice`** — project invoice.
- `id` PK, `proyekId` → proyek(CASCADE), `nomor` TEXT UNIQUE (auto-gen `INV-{proyekId}-{NNNN}`), `tanggal` TIMESTAMP(3), `total` DECIMAL(15,2), `status` StatusInvoice default 'draft', timestamps.
- Index on `proyekId`.

**`logistik`** — logistics/material tracking.
- `id` PK, `proyekId` → proyek(CASCADE), `namaMaterial` TEXT, `satuan` TEXT, `volume` DECIMAL(15,4), `hargaSatuan` DECIMAL(15,2), `totalBiaya` DECIMAL(15,2) (= volume×hargaSatuan), `tanggal` TIMESTAMP(3)?, `keterangan` TEXT?, timestamps.

**`realisasi`** — actual financial realization (expenditure tracking).
- `id` PK, `proyekId` → proyek(CASCADE), `tanggal` TIMESTAMP(3), `kategori` TEXT, `jumlah` DECIMAL(15,2), `keterangan` TEXT?, timestamps.

**`feedback`** — user feedback tickets.
- `id` PK, `userId` → users(CASCADE), `subject` TEXT, `message` TEXT, `status` StatusFeedback default 'open', timestamps. Relations: `replies`.

**`feedback_reply`** — replies on feedback (admin or user).
- `id` PK, `feedbackId` → feedback(CASCADE), `userId` → users(CASCADE), `message` TEXT, `isAdmin` BOOL default false, timestamps.

**`news`** — admin-managed news/announcements.
- `id` PK, `title` TEXT, `content` TEXT, `isActive` BOOL default true, timestamps.

**`audit_log`** — audit trail (§3.6).
- `id` PK, `proyekId` INT? → proyek(CASCADE), `pekerjaanId` INT? → pekerjaan(CASCADE), `userId` → users(CASCADE), `action` TEXT, `entityType` TEXT, `entityId` INT?, `oldValue` JSONB?, `newValue` JSONB?, `description` TEXT?, `ipAddress` TEXT?, `userAgent` TEXT?, `createdAt`.
- Indexes: `proyekId`, `userId`, `action`, `createdAt`.

### 4.3 ERD (text)

```
users 1───∞ proyek          (userId CASCADE)
users 1───∞ tim_proyek      (userId CASCADE)  ┐
proyek 1──∞ tim_proyek      (proyekId CASCADE)┘ UNIQUE(proyekId,userId)
proyek 1──∞ pekerjaan       (proyekId CASCADE)
pekerjaan 1──∞ detail_analisa (pekerjaanId CASCADE)
detail_analisa ∞──1 master_harga  (masterHargaId SET NULL)
detail_analisa ∞──1 master_analisa(masterAnalisaId SET NULL)
master_analisa 1──∞ master_analisa (parentId CASCADE, self-rel "MasterAnalisaHierarchy")
master_analisa 1──∞ rincian_analisa (masterAnalisaId CASCADE)
rincian_analisa ∞──1 master_harga   (komponenId CASCADE, nullable)
users 1───∞ master_analisa  (userId SET NULL)
users 1───∞ master_harga    (userId SET NULL)
proyek 1──∞ rekap, invoice, logistik, realisasi  (all CASCADE)
users 1───∞ feedback        (userId CASCADE)
feedback 1──∞ feedback_reply (feedbackId CASCADE)
users 1───∞ feedback_reply  (userId CASCADE)
proyek 1──∞ audit_log       (proyekId CASCADE)
pekerjaan 1──∞ audit_log    (pekerjaanId CASCADE)
users 1───∞ audit_log       (userId CASCADE)
```

> **Porting to Go / raw-PG:** Generate Go structs with `sqlc` (recommended — type-safe SQL-first) or write them by hand matching `src/types/index.ts`. Keep `decimal.Decimal` for all DECIMAL columns. JSONB columns (`oldValue`/`newValue`) → `json.RawMessage` or `[]byte` in Go. Enums → Go string types with constants. The self-referential `master_analisa` is the only tricky relation — model with a `ParentID *int32` and load children with a recursive CTE or a flat query + in-memory tree build.

---

## 5. Database Schema & Migrations

### 5.1 The three artifacts (and the inconsistency)

1. **`prisma/schema.prisma`** — the source of truth. Includes all AHSP columns. This is what `prisma generate` and `prisma db push` use.
2. **`supabase-migration-clean.sql`** (420 lines) — Prisma-generated DDL, the **pre-AHSP baseline**. Does NOT have `master_analisa.harga_satuan/kategori/ahsp_kode/ahsp_sheet/biaya_umum/is_system`, `master_harga.kode_ahsp/is_system`, or the `rincian_analisa` snapshot fields; has `rincian_analisa.komponenId` as NOT NULL.
3. **`prisma/migrations/20260725_add_ahsp_support.sql`** (42 lines) — the **delta** that brings the DB up to `schema.prisma`. All statements `IF NOT EXISTS` for idempotency. Adds the AHSP columns, makes `komponen_id` nullable, adds snapshot fields to `rincian_analisa`, creates `pg_trgm` extension + GIN indexes.

**The true current schema = clean baseline + delta.** `schema.prisma` matches that combination. There's also a `schema.prisma.backup` (ignore).

> **Porting to Go / raw-PG:** Throw away all three files and use the **consolidated DDL below** as your starting `schema.sql`. From then on, manage migrations with `golang-migrate` or `goose` — plain SQL files, no ORM. The consolidated DDL is already runnable on a fresh Postgres (≥14 for `pg_trgm`).

### 5.2 Consolidated raw-SQL DDL (runnable)

```sql
-- =====================================================================
-- Halora Land — consolidated schema (clean baseline + AHSP delta merged)
-- Target: PostgreSQL ≥ 14.  Runnable on a fresh DB.
-- =====================================================================

-- Required extension for trigram fuzzy search on master_analisa
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------
CREATE TYPE "Role" AS ENUM ('ADMIN', 'OWNER', 'USER', 'DEMO');
CREATE TYPE "TipeProyek" AS ENUM ('gedung', 'infra');
CREATE TYPE "RoleTimProyek" AS ENUM ('owner', 'editor', 'viewer');
CREATE TYPE "KategoriPekerjaan" AS ENUM (
  'persiapan','pondasi','beton','kanopi','baja','tangga','atap',
  'dinding','plesteran','acian','keramik','paving','pengecatan',
  'pintu','interior','toilet','mep','custom'
);
CREATE TYPE "MetodeHitung" AS ENUM (
  'ahsp','manual','harga_borong','harga_manual','harga_custom'
);
CREATE TYPE "TipeKomponen" AS ENUM ('material', 'upah', 'alat');
CREATE TYPE "StatusInvoice" AS ENUM ('draft', 'sent', 'paid');
CREATE TYPE "StatusFeedback" AS ENUM ('open', 'in_progress', 'resolved', 'closed');

-- ---------------------------------------------------------------------
-- Tables
-- ---------------------------------------------------------------------
CREATE TABLE "users" (
    "id"              SERIAL NOT NULL,
    "namaLengkap"     TEXT    NOT NULL,
    "email"           TEXT    NOT NULL,
    "password"        TEXT    NOT NULL DEFAULT '',   -- vestigial; Supabase owns creds
    "role"            "Role"  NOT NULL DEFAULT 'USER',
    "accountType"     TEXT    NOT NULL DEFAULT 'free',
    "isDemo"          BOOLEAN NOT NULL DEFAULT false,
    "supabaseAuthId"  TEXT,
    "createdAt"       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "proyek" (
    "id"           SERIAL NOT NULL,
    "userId"       INTEGER NOT NULL,
    "namaProyek"   TEXT    NOT NULL,
    "lokasi"       TEXT,
    "tipe"         "TipeProyek" NOT NULL DEFAULT 'gedung',
    "nilaiKontrak" DECIMAL(15,2),
    "timeline"     TEXT,
    "createdAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "proyek_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "tim_proyek" (
    "id"        SERIAL NOT NULL,
    "proyekId"  INTEGER NOT NULL,
    "userId"    INTEGER NOT NULL,
    "role"      "RoleTimProyek" NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "tim_proyek_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "pekerjaan" (
    "id"              SERIAL NOT NULL,
    "proyekId"        INTEGER NOT NULL,
    "kategori"        "KategoriPekerjaan" NOT NULL,
    "uraianPekerjaan" TEXT    NOT NULL,
    "volume"          DECIMAL(15,4) NOT NULL,
    "satuan"          TEXT    NOT NULL,
    "hargaSatuan"     DECIMAL(15,2) NOT NULL,
    "totalBiaya"      DECIMAL(15,2) NOT NULL,
    "metodeHitung"    "MetodeHitung" NOT NULL,
    "levelPekerjaan"  TEXT,
    "tipePekerjaan"   TEXT,
    "createdAt"       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "pekerjaan_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "detail_analisa" (
    "id"              SERIAL NOT NULL,
    "pekerjaanId"     INTEGER NOT NULL,
    "masterHargaId"   INTEGER,          -- nullable: drift-detection linkage only
    "masterAnalisaId" INTEGER,          -- nullable: lineage only
    "nama"            TEXT    NOT NULL, -- frozen
    "satuan"          TEXT    NOT NULL, -- frozen
    "koef"            DECIMAL(10,6) NOT NULL,
    "hargaSatuan"     DECIMAL(15,2) NOT NULL,  -- FROZEN — source of truth
    "totalBiaya"      DECIMAL(15,2) NOT NULL,  -- FROZEN
    "tipe"            "TipeKomponen" NOT NULL,
    "snapshotAt"      TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "sourceKode"      TEXT,
    CONSTRAINT "detail_analisa_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "master_analisa" (
    "id"          SERIAL NOT NULL,
    "kode"        TEXT    NOT NULL,
    "nama"        TEXT    NOT NULL,
    "level"       INTEGER NOT NULL,     -- 0=root,1=divisi,2=kelompok,3=sub,4=item
    "parentId"    INTEGER,              -- self-FK (cascade)
    "satuan"      TEXT,
    "hargaSatuan" DECIMAL(15,2),        -- unit price (from Excel col 8)
    "kategori"    TEXT,
    "isGlobal"    BOOLEAN NOT NULL DEFAULT false,
    "userId"      INTEGER,              -- owner; null for system/global
    "isSystem"    BOOLEAN NOT NULL DEFAULT false,  -- true = from AHSP 2026
    "ahspKode"    TEXT,                 -- original AHSP code e.g. 2.2.1.1.1
    "ahspSheet"   TEXT,                 -- source Excel sheet e.g. Beton
    "biayaUmum"   DECIMAL(5,4) NOT NULL DEFAULT 0.10,  -- overhead ratio (0.10 = 10%)
    "createdAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "master_analisa_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "rincian_analisa" (
    "id"             SERIAL NOT NULL,
    "masterAnalisaId" INTEGER NOT NULL,
    "komponenId"     INTEGER,           -- nullable: link to master_harga OR snapshot-only
    "koef"           DECIMAL(10,6) NOT NULL,
    "tipe"           "TipeKomponen" NOT NULL,
    -- Snapshot fields (for AHSP breakdown; used when komponenId is null)
    "nama"           TEXT,
    "satuan"         TEXT,
    "hargaSatuan"    DECIMAL(15,2),
    "jumlahHarga"    DECIMAL(15,2),
    "kodeReferensi"  TEXT,              -- L.01, M.123, E.456
    "urutan"         INTEGER NOT NULL DEFAULT 0,
    "createdAt"      TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"      TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "rincian_analisa_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "master_harga" (
    "id"        SERIAL NOT NULL,
    "nama"      TEXT    NOT NULL,
    "satuan"    TEXT    NOT NULL,
    "harga"     DECIMAL(15,2) NOT NULL,
    "kategori"  "TipeKomponen" NOT NULL,
    "isGlobal"  BOOLEAN NOT NULL DEFAULT false,
    "userId"    INTEGER,
    "kodeAHSP"  TEXT,
    "isSystem"  BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "master_harga_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "rekap" (
    "id"        SERIAL NOT NULL,
    "proyekId"  INTEGER NOT NULL,
    "kategori"  TEXT    NOT NULL,       -- 'settings' row holds project margin
    "uraian"    TEXT    NOT NULL,
    "urutan"    INTEGER NOT NULL,
    "margin"    DECIMAL(5,2),
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "rekap_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "invoice" (
    "id"        SERIAL NOT NULL,
    "proyekId"  INTEGER NOT NULL,
    "nomor"     TEXT    NOT NULL,       -- auto: INV-{proyekId}-{NNNN}
    "tanggal"   TIMESTAMP(3) NOT NULL,
    "total"     DECIMAL(15,2) NOT NULL,
    "status"    "StatusInvoice" NOT NULL DEFAULT 'draft',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "invoice_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "logistik" (
    "id"           SERIAL NOT NULL,
    "proyekId"     INTEGER NOT NULL,
    "namaMaterial" TEXT    NOT NULL,
    "satuan"       TEXT    NOT NULL,
    "volume"       DECIMAL(15,4) NOT NULL,
    "hargaSatuan"  DECIMAL(15,2) NOT NULL,
    "totalBiaya"   DECIMAL(15,2) NOT NULL,  -- = volume × hargaSatuan
    "tanggal"      TIMESTAMP(3),
    "keterangan"   TEXT,
    "createdAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"    TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "logistik_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "realisasi" (
    "id"         SERIAL NOT NULL,
    "proyekId"   INTEGER NOT NULL,
    "tanggal"    TIMESTAMP(3) NOT NULL,
    "kategori"   TEXT    NOT NULL,
    "jumlah"     DECIMAL(15,2) NOT NULL,
    "keterangan" TEXT,
    "createdAt"  TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"  TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "realisasi_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "feedback" (
    "id"        SERIAL NOT NULL,
    "userId"    INTEGER NOT NULL,
    "subject"   TEXT    NOT NULL,
    "message"   TEXT    NOT NULL,
    "status"    "StatusFeedback" NOT NULL DEFAULT 'open',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "feedback_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "feedback_reply" (
    "id"         SERIAL NOT NULL,
    "feedbackId" INTEGER NOT NULL,
    "userId"     INTEGER NOT NULL,
    "message"    TEXT    NOT NULL,
    "isAdmin"    BOOLEAN NOT NULL DEFAULT false,
    "createdAt"  TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"  TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "feedback_reply_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "news" (
    "id"        SERIAL NOT NULL,
    "title"     TEXT    NOT NULL,
    "content"   TEXT    NOT NULL,
    "isActive"  BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "news_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "audit_log" (
    "id"          SERIAL NOT NULL,
    "proyekId"    INTEGER,
    "pekerjaanId" INTEGER,
    "userId"      INTEGER NOT NULL,
    "action"      TEXT    NOT NULL,     -- 'CREATE'|'UPDATE'|'DELETE'|'LOGIN'|'recalculate'|...
    "entityType"  TEXT    NOT NULL,     -- 'pekerjaan'|'USER'|...
    "entityId"    INTEGER,
    "oldValue"    JSONB,                -- before-state (rich audit module)
    "newValue"    JSONB,                -- after-state
    "description" TEXT,                 -- JSON.stringify(metadata) (lightweight module)
    "ipAddress"   TEXT,
    "userAgent"   TEXT,
    "createdAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "audit_log_pkey" PRIMARY KEY ("id")
);

-- ---------------------------------------------------------------------
-- Indexes (unique)
-- ---------------------------------------------------------------------
CREATE UNIQUE INDEX "users_email_key"                     ON "users"("email");
CREATE UNIQUE INDEX "users_supabaseAuthId_key"            ON "users"("supabaseAuthId");
CREATE UNIQUE INDEX "tim_proyek_proyekId_userId_key"      ON "tim_proyek"("proyekId","userId");
CREATE UNIQUE INDEX "master_analisa_kode_userId_key"      ON "master_analisa"("kode","userId");
CREATE UNIQUE INDEX "master_harga_nama_userId_kategori_key" ON "master_harga"("nama","userId","kategori");
CREATE UNIQUE INDEX "invoice_nomor_key"                   ON "invoice"("nomor");

-- ---------------------------------------------------------------------
-- Indexes (non-unique)
-- ---------------------------------------------------------------------
CREATE INDEX "tim_proyek_userId_idx"              ON "tim_proyek"("userId");
CREATE INDEX "tim_proyek_proyekId_idx"            ON "tim_proyek"("proyekId");
CREATE INDEX "pekerjaan_proyekId_idx"             ON "pekerjaan"("proyekId");
CREATE INDEX "pekerjaan_kategori_idx"             ON "pekerjaan"("kategori");
CREATE INDEX "detail_analisa_pekerjaanId_idx"     ON "detail_analisa"("pekerjaanId");
CREATE INDEX "detail_analisa_masterHargaId_idx"   ON "detail_analisa"("masterHargaId");
CREATE INDEX "detail_analisa_masterAnalisaId_idx" ON "detail_analisa"("masterAnalisaId");
CREATE INDEX "master_analisa_userId_idx"          ON "master_analisa"("userId");
CREATE INDEX "master_analisa_parentId_idx"        ON "master_analisa"("parentId");
CREATE INDEX "rincian_analisa_masterAnalisaId_idx" ON "rincian_analisa"("masterAnalisaId");
CREATE INDEX "rincian_analisa_komponenId_idx"     ON "rincian_analisa"("komponenId");
CREATE INDEX "master_harga_userId_idx"            ON "master_harga"("userId");
CREATE INDEX "rekap_proyekId_idx"                 ON "rekap"("proyekId");
CREATE INDEX "invoice_proyekId_idx"               ON "invoice"("proyekId");
CREATE INDEX "logistik_proyekId_idx"              ON "logistik"("proyekId");
CREATE INDEX "realisasi_proyekId_idx"             ON "realisasi"("proyekId");
CREATE INDEX "feedback_userId_idx"                ON "feedback"("userId");
CREATE INDEX "feedback_reply_feedbackId_idx"     ON "feedback_reply"("feedbackId");
CREATE INDEX "feedback_reply_userId_idx"          ON "feedback_reply"("userId");
CREATE INDEX "audit_log_proyekId_idx"             ON "audit_log"("proyekId");
CREATE INDEX "audit_log_userId_idx"               ON "audit_log"("userId");
CREATE INDEX "audit_log_action_idx"               ON "audit_log"("action");
CREATE INDEX "audit_log_createdAt_idx"            ON "audit_log"("createdAt");

-- AHSP-specific indexes (from the delta migration)
CREATE INDEX "idx_master_analisa_kategori_is_system" ON "master_analisa"("kategori","is_system");
CREATE INDEX "idx_master_analisa_ahsp_kode"          ON "master_analisa"("ahspKode");
CREATE INDEX "idx_master_analisa_ahsp_sheet"         ON "master_analisa"("ahspSheet");
CREATE INDEX "idx_master_harga_kode_ahsp"            ON "master_harga"("kodeAHSP");
-- Trigram fuzzy search (pg_trgm). NOTE: current app uses ILIKE, not trigram ops —
-- these are forward-infra. Keep them; switch the search queries to use % and similarity().
CREATE INDEX "idx_master_analisa_nama_trgm"          ON "master_analisa" USING gin ("nama" gin_trgm_ops);
CREATE INDEX "idx_master_analisa_ahsp_kode_trgm"     ON "master_analisa" USING gin ("ahspKode" gin_trgm_ops);

-- ---------------------------------------------------------------------
-- Foreign keys
-- ---------------------------------------------------------------------
ALTER TABLE "proyek"
  ADD CONSTRAINT "proyek_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "tim_proyek"
  ADD CONSTRAINT "tim_proyek_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "tim_proyek"
  ADD CONSTRAINT "tim_proyek_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "pekerjaan"
  ADD CONSTRAINT "pekerjaan_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "detail_analisa"
  ADD CONSTRAINT "detail_analisa_pekerjaanId_fkey"
  FOREIGN KEY ("pekerjaanId") REFERENCES "pekerjaan"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- SET NULL: snapshot survives even if the master is deleted (§3.1)
ALTER TABLE "detail_analisa"
  ADD CONSTRAINT "detail_analisa_masterHargaId_fkey"
  FOREIGN KEY ("masterHargaId") REFERENCES "master_harga"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "detail_analisa"
  ADD CONSTRAINT "detail_analisa_masterAnalisaId_fkey"
  FOREIGN KEY ("masterAnalisaId") REFERENCES "master_analisa"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "master_analisa"
  ADD CONSTRAINT "master_analisa_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- Self-referential hierarchy
ALTER TABLE "master_analisa"
  ADD CONSTRAINT "master_analisa_parentId_fkey"
  FOREIGN KEY ("parentId") REFERENCES "master_analisa"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "rincian_analisa"
  ADD CONSTRAINT "rincian_analisa_masterAnalisaId_fkey"
  FOREIGN KEY ("masterAnalisaId") REFERENCES "master_analisa"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "rincian_analisa"
  ADD CONSTRAINT "rincian_analisa_komponenId_fkey"
  FOREIGN KEY ("komponenId") REFERENCES "master_harga"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "master_harga"
  ADD CONSTRAINT "master_harga_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "rekap"
  ADD CONSTRAINT "rekap_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "invoice"
  ADD CONSTRAINT "invoice_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "logistik"
  ADD CONSTRAINT "logistik_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "realisasi"
  ADD CONSTRAINT "realisasi_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "feedback"
  ADD CONSTRAINT "feedback_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "feedback_reply"
  ADD CONSTRAINT "feedback_reply_feedbackId_fkey"
  FOREIGN KEY ("feedbackId") REFERENCES "feedback"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "feedback_reply"
  ADD CONSTRAINT "feedback_reply_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "audit_log"
  ADD CONSTRAINT "audit_log_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "audit_log"
  ADD CONSTRAINT "audit_log_proyekId_fkey"
  FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "audit_log"
  ADD CONSTRAINT "audit_log_pekerjaanId_fkey"
  FOREIGN KEY ("pekerjaanId") REFERENCES "pekerjaan"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- ---------------------------------------------------------------------
-- Column comments (documentation)
-- ---------------------------------------------------------------------
COMMENT ON COLUMN "master_analisa"."isSystem"  IS 'true = from AHSP 2026, false = user custom';
COMMENT ON COLUMN "master_analisa"."ahspKode"  IS 'Original AHSP code e.g. 2.2.1.1.1';
COMMENT ON COLUMN "master_analisa"."ahspSheet" IS 'Source Excel sheet name e.g. Beton';
COMMENT ON COLUMN "master_analisa"."biayaUmum" IS 'Overhead percentage e.g. 0.10 = 10%';
COMMENT ON COLUMN "users"."password"           IS 'Vestigial — Supabase owns credentials. Safe to drop in rework.';

-- =====================================================================
-- End of schema
-- =====================================================================
```

### 5.3 Seed data (`prisma/seed.ts`)

The seed wipes all tables then creates:
- **3 users** (all `password:'password123'` bcrypt-hashed, but irrelevant — Supabase holds creds):
  - `admin@haloraland.id` — ADMIN, accountType 'pro'
  - `budi@example.com` — USER, 'free'
  - `demo@haloraland.id` — DEMO, 'free', `isDemo:true`
- **2 projects** for Budi: "Pembangunan Rumah 2 Lantai" (Jakarta Selatan, gedung, 850M, 6mo), "Renovasi Kantor" (Tangerang, gedung, 450M, 3mo).
- **3 master_harga** (global): Semen Portland (zak, 75000, material), Pasir Pasang (m3, 350000, material), Tukang Batu (hari, 150000, upah).
- **2 pekerjaan** on project 1: Galian Tanah Pondasi (25.5 m3, 125000, manual), Pasang Bata Merah (120 m2, 185000, ahsp).
- **3 detail_analisa** for pekerjaan 2 (snapshot: semen koef 1.5, pasir 0.5, tukang 1.0 — all linked to their master_harga).
- **1 news** item.

**The seed does NOT populate AHSP data.** The AHSP library (`master_analisa` + `rincian_analisa`) is imported on-demand via `POST /api/admin/ahsp/import` per Excel sheet.

> **Porting to Go / raw-PG:** Replace `seed.ts` with a Go CLI: `./halora-be seed` (users + sample project) and `./halora-be import-ahsp --file ./data/ahsp-2026.xlsx --sheet Beton [--force]`. Use `database/sql` + transactions. The demo Supabase user must be provisioned via Supabase Admin API or manually in the Supabase dashboard before seed runs.

---

## 6. API Surface (all 35 routes)

**Methods:** GET×21, POST×16, PUT×11, DELETE×8, PATCH×0.

### 6.1 Auth (`src/app/api/auth/`)

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| POST | `/api/auth/register` | public (RL: 5/15min/IP) | `{email, password, namaLengkap}` | Supabase signUp + create local user (role USER). Audit `REGISTER`. |
| POST | `/api/auth/login` | public (RL: 10/15min/IP) | `{email, password}` | Supabase signIn, find-or-create local user, set cookie. Audit `LOGIN`. |
| POST | `/api/auth/logout` | (cookie) | — | `supabase.auth.signOut()`. |
| GET | `/api/auth/me` | user | — | Current user profile. |
| PUT | `/api/auth/me` | user | `{namaLengkap?, email?}` | Update profile (email uniqueness check). |
| POST | `/api/auth/update-password` | user | `{password}` (≥8, upper/lower/digit) | `supabase.auth.updateUser({password})`. |
| POST | `/api/auth/resend-verification` | public (RL: 3/15min/email) | `{email}` | `supabase.auth.resend({type:'signup', email})`. |

### 6.2 Proyek (`src/app/api/proyek/`)

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| GET | `/api/proyek` | user | `?tipe=gedung\|infra&search=` | List projects (non-admin: owned + team). |
| POST | `/api/proyek` | user | `{namaProyek, lokasi?, tipe?, nilaiKontrak?, timeline?}` | Create project. |
| GET | `/api/proyek/[id]` | user+access | — | Single project + user + timProyek + last 10 pekerjaan + counts. |
| PUT | `/api/proyek/[id]` | edit | `createProyekSchema.partial()` | Update project. |
| DELETE | `/api/proyek/[id]` | owner\|ADMIN | — | Delete project. |
| GET | `/api/proyek/[id]/rekap` | user+access | — | RAB rekap: group pekerjaan by kategori, subtotals, grand total, overhead, profit, PPN 11%, totalAkhir. |
| PUT | `/api/proyek/[id]/rekap` | editor | `{margin}` | Upsert rekap settings row margin. |
| POST | `/api/proyek/[id]/recalculate-all` | edit | — | Bulk re-snapshot all AHSP pekerjaan (1 txn, 1 audit_log per item, `bulk_recalculate`). |
| GET | `/api/proyek/[id]/realisasi` | user+access | `?startDate=&endDate=` | List realisasi + totals + monthly trend. |
| POST | `/api/proyek/[id]/realisasi` | editor | `{tanggal, kategori, jumlah, keterangan?}` | Create realisasi. |
| GET/PUT/DELETE | `/api/proyek/[id]/realisasi/[realisasiId]` | user\|editor | — | CRUD single realisasi. |
| GET/PUT | `/api/proyek/[id]/progress` | user\|editor | PUT: `{pekerjaanId, progress, notes?}` | Per-kategori progress + summary (progress currently mock). PUT is a stub (TODO persist). |
| POST | `/api/proyek/[id]/pekerjaan/from-ahsp` | edit | `{masterAnalisaId, volume, applyBreakdown?=true}` | Create Pekerjaan from AHSP catalog item + copy rincian→detail_analisa (§2.3). |
| GET | `/api/proyek/[id]/logistik` | user+access | — | List logistik + total biaya. |
| POST | `/api/proyek/[id]/logistik` | editor | `{namaMaterial, satuan, volume, hargaSatuan, tanggal?, keterangan?}` | Create logistik (computes totalBiaya). |
| GET/PUT/DELETE | `/api/proyek/[id]/logistik/[logistikId]` | user\|editor | — | CRUD single logistik. |
| GET | `/api/proyek/[id]/kurva-s` | user+access | — | S-curve (planned vs actual). **Returns dummy data** (TODO compute from realisasi). |
| GET | `/api/proyek/[id]/invoice` | user+access | — | List invoices. |
| POST | `/api/proyek/[id]/invoice` | editor | `{tanggal, total, status?='draft'}` | Create invoice (auto `nomor=INV-{proyekId}-{NNNN}`). |
| GET/PUT/DELETE | `/api/proyek/[id]/invoice/[invoiceId]` | user\|editor | — | CRUD single invoice. |

### 6.3 Pekerjaan (`src/app/api/pekerjaan/`)

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| GET | `/api/pekerjaan` | user+access | `?proyekId=*&kategori=&search=` | List pekerjaan (includes detailAnalisa.masterHarga). |
| POST | `/api/pekerjaan` | edit | `createPekerjaanSchema` (proyekId, kategori, uraianPekerjaan, volume, satuan, metodeHitung, tipePekerjaan?, masterAnalisaId?, hargaSatuan?, detailAnalisa?[]) | Create pekerjaan; for `ahsp` calls `snapshotAHSP`, for `manual` uses provided values. |
| GET | `/api/pekerjaan/[id]` | user+access | — | Single pekerjaan + detailAnalisa. |
| PUT | `/api/pekerjaan/[id]` | edit | `createPekerjaanSchema.partial()` | Update (recomputes `totalBiaya = volume × hargaSatuan`). |
| DELETE | `/api/pekerjaan/[id]` | owner\|ADMIN | — | Delete (requires `tim_proyek.role='owner'`). |
| POST | `/api/pekerjaan/[id]/analisa` | edit | `{masterHargaId?, nama, satuan, koef, hargaSatuan, tipe}` | Add DetailAnalisa component (manual; computes totalBiaya). |
| GET | `/api/pekerjaan/[id]/analisa` | user+access | — | List detail analisa. |
| POST | `/api/pekerjaan/[id]/recalculate` | edit | — | Re-snapshot single AHSP pekerjaan; audit `recalculate`. |
| GET | `/api/pekerjaan/[id]/validate-snapshot` | user+access | — | Drift check: `{isValid, changes[]}` with per-component diffs. |

### 6.6 Master Analisa (`src/app/api/master-analisa/`)

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| GET | `/api/master-analisa` | user | `?level=0-4&parentId='null'\|<id>&search=&isGlobal=true\|false` | List tree (global + user-owned). |
| POST | `/api/master-analisa` | user (global: ADMIN) | `createMasterAnalisaSchema` (kode, nama, level, parentId?, satuan?, isGlobal?) | Create node (global only by ADMIN; checks kode uniqueness). |
| GET | `/api/master-analisa/[id]` | user | — | Single + parent + children + rincian. |
| PUT | `/api/master-analisa/[id]` | owner\|ADMIN(if global) | `createMasterAnalisaSchema.partial()` | Update (kode uniqueness check). |
| DELETE | `/api/master-analisa/[id]` | owner\|ADMIN(if global) | — | Delete if no children (409 if children). |
| GET | `/api/master-analisa/[id]/rincian` | user | — | List rincian + komponen + totalHargaSatuan. |
| POST | `/api/master-analisa/[id]/rincian` | owner\|ADMIN(if global) | `{komponenId, koef, tipe}` | Add komponen (dup check). |
| DELETE | `/api/master-analisa/[id]/rincian` | owner\|ADMIN(if global) | `{rincianId}` | Remove rincian. |
| GET | `/api/master-analisa/search` | user | `?q=&kategori=&limit=20` | Search system AHSP items (`isSystem:true`) by nama/ahspKode ILIKE. |

### 6.5 Master Harga (`src/app/api/master-harga/`)

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| GET | `/api/master-harga` | user | `?kategori=material\|upah\|alat&search=&isGlobal=` | List (global + user-owned). |
| POST | `/api/master-harga` | user (global: ADMIN) | `{nama, satuan, harga, kategori, isGlobal?}` | Create (non-admin: isGlobal silently false). |
| GET | `/api/master-harga/[id]` | user | — | Single (global visible; user-owned by owner). |
| PUT | `/api/master-harga/[id]` | owner\|ADMIN(if global) | `createMasterHargaSchema.partial()` | Update. |
| DELETE | `/api/master-harga/[id]` | owner\|ADMIN(if global) | — | Delete. |

### 6.6 Admin (`src/app/api/admin/`)

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| POST | `/api/admin/ahsp/import` | **ADMIN** | `{sheetName, kategori?, forceReimport?}` | Import AHSP sheet from `public/data/ahsp-2026.xlsx` → master_analisa + rincian_analisa (txn). |
| GET | `/api/admin/ahsp/import` | **ADMIN** | — | Import status per sheet (imported? count). |

### 6.6 Dashboard, Audit, Feedback, Monitoring

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| GET | `/api/dashboard/stats` | user | — | totalProyek, proyekAktif, totalRAB (Σ hargaSatuan×volume), totalPekerjaan, 5 recent projects + per-project RAB, last 10 audit logs. |
| GET | `/api/audit-log` | user | `?proyekId=&entityType=&action=&userId=&limit=50(max200)` | List audit logs (non-admin scoped to own/team projects). |
| GET | `/api/feedback` | user | — | List current user's feedback + replies. |
| POST | `/api/feedback` | user | `{message(10-2000), rating?1-5, category?bug\|feature\|question\|other='other'}` | Create feedback (stores category in `subject`). |
| GET | `/api/monitoring` | user+access | `?proyekId=` (required) | Per-kategori progress (hardcoded 0 — TODO). |

**Access legend:** `user` = any authenticated; `user+access` = user + project ownership/team check; `edit` = non-viewer team member or owner or ADMIN; `owner|ADMIN` = project owner or ADMIN; `editor` = `tim_proyek.role='editor'`.

> **Porting to Go / raw-PG:** Map each route 1:1 to a Go handler. Use `chi` or `net/http` with a middleware chain (`Auth → Access → Handler`). The `from-ahsp` and `recalculate-all` routes are the only ones needing transactions — the rest are simple CRUD. The query-heavy endpoints (rekap, dashboard/stats, audit-log list) benefit from raw SQL views or CTEs — much cleaner in Go+sqlc than Prisma's nested includes. **Drop the dummy endpoints** (kurva-s, monitoring, progress PUT) or actually implement them in Go from the start.

---

## 7. Auth & Session Model (deep dive)

### 7.1 Cookie mechanics (live)

- **Cookie name:** `sb-<projectRef>-auth-token` where `projectRef` = `NEXT_PUBLIC_SUPABASE_URL` split on `//` and `.` (e.g. `https://abcdefgh.supabase.co` → `abcdefgh`).
- **Set on login** (manually, `login/route.ts:106-127`):
  - `value`: `JSON.stringify({ access_token, refresh_token, expires_at, expires_in, token_type, user })`
  - `path:'/'`, `httpOnly:true`, `secure:true`, `sameSite:'strict'`, `maxAge:session.expires_in`
- **Read by middleware** (`middleware.ts:27-44`): `JSON.parse(cookie)` → check `expires_at × 1000 > Date.now()`.
- **Read by Supabase SSR client** (`supabase-auth.ts:8-28`): wired to `next/headers` cookies; `supabase.auth.getUser()` verifies the JWT server-side.
- **Cleared on logout** via `supabase.auth.signOut()` → Supabase client's `remove` handler (`supabase-auth.ts:22-24`): `cookieStore.set({ name, value:'', ...options })`.

### 7.2 Known issues to fix in the rewrite

1. **Logout cookie-clear bug** — the manual set uses `secure:true, sameSite:'strict', path:'/'` but the Supabase client's `remove` doesn't re-specify these attributes. To reliably overwrite a cookie, deletion must match the original attributes (especially `path` and `sameSite`). Logout may leave the cookie in place.
2. **`secure:true` blocks local dev** over plain HTTP (no cookie set without HTTPS). Needs a `NODE_ENV`-conditional flag.
3. **`/admin/*` pages unprotected** — no middleware, no layout check. Admin users page reads mock data. Only API routes enforce ADMIN.
4. **`OWNER` role unused**, `isDemo`/`DEMO` surfaced but never enforced.
5. **No central `requireAuth`/`requireRole`** — every route hand-rolls checks (high duplication, easy to miss).
6. **`src/lib/auth.ts` + `src/proxy.ts` dead** — `JWT_SECRET` still required at module load (latent footgun). `auth-token` cookie never set.
7. **In-memory rate limiting** — won't work multi-instance.
8. **No forgot-password endpoint** — "Lupa password?" link is `href="#"`. `sendPasswordResetEmail` (`supabase-auth.ts:189`) is never called.
9. **`users.password` vestigial** — `''` on register/login; seed stores bcrypt hashes that are never checked.

> **Porting to Go / raw-PG:** In Go, issue your own JWT (or pass the Supabase access_token through). Set cookies with `http.SetCookie` with consistent attributes (`Path='/', HttpOnly=true, Secure=<prod>, SameSite=http.SameSiteStrictMode, MaxAge=<ttl>`). For logout, set the same cookie with `MaxAge=-1` and `Value=''` — this reliably clears it. Build a single `AuthMiddleware` that verifies the JWT, loads the user from `users`, and puts `{userID, role, ...}` in `context.Context`. Build a `RequireRole(roles...)` wrapper. Drop the `password` column. Use Redis for rate limiting. Actually implement forgot-password (`supabase.auth.ResetPasswordForEmail` or your own email flow).

---

## 8. Business Logic to Port

### 8.1 AHSP parser (`src/lib/ahsp-parser.ts`, 213 lines)

Port to Go with `github.com/xuri/excelize/v2`. Key logic to preserve:
- `SHEET_TO_KATEGORI` hardcoded map (sheet name → kategori slug); first 2 sheets skipped; unknown → `'custom'`.
- Work item detection: col 2 = kode (≥4 numeric dot-levels), col 3 = nama, col 8 = hargaSatuan.
- Breakdown extraction: scan up to 100 rows; col 2 section markers `A`→upah, `B`→material, `C`→alat, `D/E/F`→end. Column indices with fallbacks (`row[4]||row[5]`, etc.) for non-uniform sheets.
- `biayaUmum` hardcoded `0.10`.
- `extractSatuan` regex: `m|m2|m3|kg|ton|unit|buah|ls|liter|OH|hari|jam` from col 4 or embedded in nama.

### 8.2 Snapshot engine (`src/lib/snapshot.ts`, ~250 lines)

Three functions:
- `snapshotAHSP(masterAnalisaId, volume)` — loads rincian + komponen, freezes `{nama, satuan, koef, hargaSatuan (from master_harga.harga), totalBiaya=koef×harga, tipe, snapshotAt, sourceKode}`. Returns `hargaSatuan = Σ totalBiaya`, `totalBiaya = hargaSatuan × volume`.
- `validateSnapshot(pekerjaanId)` — for each DetailAnalisa with non-null `masterHargaId`, compares `detail.hargaSatuan` vs `master_harga.harga`. Returns `{isValid, changes[]}`.
- `compareSnapshot(pekerjaanId)` — recomputes hypothetical cost from live master prices (falls back to frozen `detail.totalBiaya` when master deleted). Returns per-component + aggregate diff.

> **Go port:** All three are single SQL queries or small query+process functions. `validateSnapshot` is `SELECT ... LEFT JOIN master_harga ... WHERE detail.harga_satuan IS DISTINCT FROM master_harga.harga`. The fallback in `compareSnapshot` is `COALESCE(master_harga.harga * detail.koef, detail.total_biaya)`.

### 8.3 RAB rollup formula

```go
subtotal := sum(pekerjaan.totalBiaya)
overhead := subtotal × 0.10                    // hardcoded — make configurable
profit   := (subtotal + overhead) × (margin/100)
ppn      := (subtotal + overhead + profit) × 0.11   // 11% VAT — make configurable
total    := subtotal + overhead + profit + ppn
```
Currently duplicated in `utils.ts:calculateRAB`, `rab/page.tsx:154-157`, `rekap/page.tsx:70-73`. The `rekap/route.ts` API version has `overhead=0, profit=0` (TODO). **Centralize in the Go BE**; have the FE just display BE-computed totals.

### 8.4 Audit logging

One Go package `internal/audit`:
```go
func Log(ctx context.Context, q *sql.Tx, p AuditParams) // action, entityType, entityID, old, new, proyekID, pekerjaanID, userID, ip, ua
```
Run via a buffered channel + background worker (true non-blocking). Use `JSONB` for old/new. Wire into **all** mutations (current coverage is only LOGIN/REGISTER/recalculate).

### 8.5 Utility functions (`src/lib/utils.ts`)

- `formatCurrency` — `Intl.NumberFormat("id-ID", {currency:"IDR", maximumFractionDigits:0})`. Go: `fmt.Sprintf("%.0f", float64(amount))` or `message/printer`.
- `formatDate`/`formatDateShort` — Indonesian locale. Go: `time.Format` with custom layout or `github.com/lestrrat-go/strftime`.
- `calculateAHS(volume, components)` — sum `koef×hargaSatuan` per tipe. Go: trivial loop with `decimal.Decimal`.
- `parseCurrency(s)` — strip non-digits. Go: `strings.Map` over digits.

> **Porting note:** Move all formatting to the FE (it's display logic, locale-specific). The BE should return raw `decimal.Decimal` / strings; the FE formats. Keep `calculateAHS` + `calculateRAB` in Go (BE is the source of truth for totals).

---

## 9. Frontend Coupling Points (for the FE/BE split)

The FE currently couples to the BE in these ways that must change:

### 9.1 Same-origin fetch
All calls are `fetch('/api/...')` (relative URLs, same-origin). Cookies ride along automatically. **After split:** FE must use an absolute `NEXT_PUBLIC_API_URL` + send credentials (`credentials:'include'` or an `Authorization: Bearer <token>` header). CORS must be configured on the Go BE (`Access-Control-Allow-Origin`, `Allow-Credentials`).

### 9.2 SSR + cookies
`ProjectContext.tsx` and the dashboard use SSR `fetch` to `/api/proyek` etc., relying on `next/headers` cookies being forwarded. **After split:** SSR fetch must explicitly forward the browser's cookies to the Go API (`headers.Cookie`), or the FE moves to client-side-only data fetching (simpler, loses SSR SEO — fine for a dashboard app).

### 9.3 `middleware.ts`
Only protects `/dashboard/*` and redirects authed users off `/login`. **After split:** Keep FE middleware for route guarding (redirect to `/login` if no session cookie), but the real auth check moves to the Go BE. The FE middleware can check a cookie existence (Supabase or Go-issued JWT) without verifying.

### 9.4 `src/proxy.ts` (DEAD)
Delete it. It's the old JWT middleware, not registered.

### 9.5 State management
- `ProjectContext` — current project ID (localStorage + `fetch('/api/proyek')`). **After split:** change fetch URL to Go API.
- `useProjectStore` (zustand) — `activeProject`, `appMode`. Pure client state, no BE coupling. Keep.
- `useSidebarStore`, `useUIStore` (zustand) — UI state. Keep.

### 9.6 `src/types/index.ts`
The shared TS contract (226 lines: `User`, `Proyek`, `Pekerjaan`, `DetailAnalisa`, `MasterAnalisa`, `RincianAnalisa`, `MasterHarga`, `Rekap`, `Invoice`, `Logistik`, `Realisasi`, `Feedback`, `FeedbackReply`, `News`, `CalculationResult`, `RABResult`). **Keep as the FE's API contract.** Generate Go structs from it (or share via OpenAPI spec generated from the Go handlers).

### 9.7 Mock data (`src/mock/`)
Many dashboard pages still read from mock data (see §10). The split is a good time to wire them to the real Go API.

### 9.8 What to DELETE from the FE during the split
- `src/app/api/**` (all 35 route files) — moved to Go.
- `src/lib/prisma.ts`, `src/lib/audit.ts`, `src/lib/audit-log.ts`, `src/lib/ahsp-parser.ts`, `src/lib/snapshot.ts`, `src/lib/api-utils.ts`, `src/lib/validate.ts`, `src/lib/rate-limit.ts`, `src/lib/schemas.ts`, `src/lib/auth.ts` (dead), `src/lib/proxy.ts` (dead).
- `prisma/` directory, `prisma.config.ts`, `@prisma/*` + `prisma` + `pg` + `@prisma/adapter-pg` deps.
- `postinstall` script (`prisma generate`).
- `supabase-migration*.sql` (replaced by Go-managed migrations).

### 9.9 What to KEEP in the FE
- `src/app/(auth)/**`, `src/app/(dashboard)/**`, `src/app/admin/**` (pages), `src/app/auth/callback` (if keeping Supabase Auth — the callback can still hit Supabase directly then redirect to the Go API to exchange).
- `src/components/**`, `src/contexts/**`, `src/stores/**`, `src/hooks/**`, `src/types/**`.
- `src/lib/utils.ts` (formatting), `src/lib/constants.ts`, `src/lib/posthog.tsx`, `src/lib/supabase.ts` (browser client, if keeping Supabase Auth).
- `middleware.ts` (simplified — just route guarding).
- `public/data/ahsp-2026.xlsx` — **move to the Go BE** (it's the BE's import source now).

> **Porting to Go / raw-PG:** Recommended FE→BE contract: Go serves `/api/v1/*` with JSON. FE uses a single `apiClient` (`src/lib/api.ts`) that prefixes `NEXT_PUBLIC_API_URL` and attaches credentials. Generate the OpenAPI spec from Go handlers (`swaggo` or hand-write) and feed it to the FE for type gen (`openapi-typescript`). This replaces `src/types/index.ts` as the source of truth — but only after the BE is stable.

---

## 10. Known Gaps & Inconsistencies

A checklist of things to fix (or consciously carry forward) in the rewrite.

| # | Issue | Where | Fix in rewrite? |
|---|---|---|---|
| 1 | 21/31 dashboard pages still use mock data (only `proyek` is fully wired) | `FRONTEND_API_MAPPING.md` | Yes — wire all to Go API |
| 2 | `kurva-s` API returns dummy planned/actual arrays | `proyek/[id]/kurva-s/route.ts` | Yes — compute from realisasi |
| 3 | `monitoring` API returns hardcoded progress=0 | `monitoring/route.ts`, `progress/route.ts` | Yes — implement progress tracking |
| 4 | `progress` PUT is a stub (TODO persist) | `proyek/[id]/progress/route.ts` | Yes |
| 5 | `rekap/route.ts` API has `overhead=0, profit=0` (TODO) while FE hardcodes 10% | `proyek/[id]/rekap/route.ts` | Yes — centralize in BE |
| 6 | `biayaUmum` stored per-item but not wired into rollup | §3.5 | Decide: per-item or project-level |
| 7 | `OWNER` role defined but never assigned/checked | §3.9 | Drop or wire it |
| 8 | `isDemo`/`DEMO` surfaced but never enforced | §3.9 | Make read-only or drop |
| 9 | `/admin/*` pages have no middleware/layout protection | §7.2 | Protect at Go middleware |
| 10 | No central `requireAuth`/`requireRole` helper | §3.9 | Go middleware chain |
| 11 | `src/lib/auth.ts` + `src/proxy.ts` dead; `JWT_SECRET` still required at load | §3.3 | Delete |
| 12 | Logout cookie-clear may not work (attribute mismatch) | §7.2 | Fix in Go cookie handling |
| 13 | `secure:true` on cookie blocks local dev HTTP | §7.2 | Env-conditional |
| 14 | In-memory rate limiting (not multi-instance) | §3.8 | Acceptable for small single-instance deploy (≈15 users); revisit if multi-instance |
| 15 | No forgot-password endpoint; "Lupa password?" is `href="#"` | §7.2 | Implement |
| 16 | `users.password` vestigial (Supabase owns creds) | §3.3 | Drop column |
| 17 | Two parallel audit modules, one table, inconsistent fields | §3.6 | One Go package |
| 18 | Audit logging only on 4 routes (most CRUD not audited) | §3.6 | Wire all mutations |
| 19 | `Pekerjaan.totalBiaya` (from MasterAnalisa.hargaSatuan) may not reconcile with Σ DetailAnalisa.totalBiaya | §2.3, RAB_IMPLEMENTATION.md TODO | Decide canonical source |
| 20 | `master_analisa.level` capped at 4 in zod but 5-part codes compute to 5 | §4.2, import route | Reconcile |
| 21 | Search uses ILIKE; `pg_trgm` extension + GIN indexes exist but unused | §5.2 | Use trigram in Go |
| 22 | PPN hardcoded 11% (Indonesia raised to 12% in 2025) | §8.3 | Make configurable |
| 23 | `RAB_IMPLEMENTATION.md:157` notes schema push was "blocked by rtk prisma issue" — migration may not be applied to live DB | | Verify with `prisma db pull` before migrating data |

---

## 11. Rework Checklist (suggested order)

1. **Stand up Go BE skeleton** — `chi`/`net/http`, `pgx` pool, `sqlc` config, run the consolidated DDL (§5.2) against a fresh Postgres.
2. **Auth** — verify Supabase JWT via JWKS; `AuthMiddleware` + `RequireRole`; cookie set/clear with consistent attributes; user lookup with auto-link. Drop `users.password`.
3. **Port CRUD** — proyek, pekerjaan (+detail_analisa), master_harga, master_analisa (+rincian), invoice, logistik, realisasi, feedback, news. Each with the inline access check → middleware.
4. **Port business logic** — `from-ahsp` (snapshot copy), `recalculate`/`recalculate-all` (txn), `validate-snapshot` (drift), `rekap` (rollup with configurable rates), `dashboard/stats`.
5. **AHSP import** — Go CLI (`excelize`) + optional HTTP endpoint; idempotent via `ahsp_kode` + `is_system`.
6. **Audit** — one `internal/audit` package, background worker, wire into all mutations.
7. **Rate limiting** — in-memory fixed-window (single-instance, small user base); apply to auth + writes.
8. **Split FE** — introduce `NEXT_PUBLIC_API_URL` + `apiClient`; delete `src/app/api/**` and the BE-coupled `lib` files; simplify `middleware.ts` to route-guarding only.
9. **Migrate data** — `pg_dump` the existing Supabase DB, restore into the new Postgres, verify row counts. Run `import-ahsp` if `master_analisa` is empty.
10. **Fix the gaps** from §10 (kurva-s, monitoring, progress, forgot-password, etc.).

---

*End of document.*
