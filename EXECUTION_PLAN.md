# HitungBangun V3 — Full Feature Execution Plan

**Auth Strategy:** Supabase Auth (email verify gratis, password reset, OAuth bonus)
**Rate Limiting:** Memory-based
**Deploy:** Vercel
**Monitoring:** Sentry (errors) + PostHog (analytics) + Vercel Analytics (pageviews)
**News:** Admin-only CRUD

**Timeline:** 10-12 days
**Total Tasks:** 51 across 6 phases

---

## Phase 0 — Pre-Flight Fixes (Day 1 AM ~4 jam)

| # | Task | Files | Est |
|---|------|-------|-----|
| 0.1 | Add indexes + updatedAt to schema.prisma | prisma/schema.prisma | 45m |
| 0.2 | Unique constraint MasterHarga + datasource URL | prisma/schema.prisma | 15m |
| 0.3 | Validation helpers | src/lib/validate.ts (NEW) | 30m |
| 0.4 | Postinstall script | package.json | 5m |
| 0.5 | Rotate JWT_SECRET | .env | 10m |
| 0.6 | Verify .env gitignored | .gitignore | 5m |

## Phase 1 — Supabase Auth + DB (Day 1 PM ~4 jam)

| # | Task | Files | Est |
|---|------|-------|-----|
| 1.1 | Install Supabase packages | package.json | 10m |
| 1.2 | Create Supabase project | Supabase dashboard (manual) | 10m |
| 1.3 | Get URL + anon key, update .env | .env | 5m |
| 1.4 | Create Supabase client | src/lib/supabase.ts (NEW) | 30m |
| 1.5 | Enable email verification | Supabase dashboard (manual) | 5m |
| 1.6 | Migrate login page | src/app/(auth)/login/page.tsx | 45m |
| 1.7 | Migrate register page | src/app/(auth)/register/page.tsx | 45m |
| 1.8 | Update getCurrentUser helper | src/lib/auth.ts | 30m |
| 1.9 | Push Prisma schema to Supabase | terminal | 10m |
| 1.10 | Test auth flow E2E | browser | 30m |

## Phase 2 — Frontend Integration (Day 2-5 ~3 hari)

| # | Task | Files | Est |
|---|------|-------|-----|
| 2.1 | API client | src/lib/api.ts (NEW) | 1h |
| 2.2 | useApiQuery hook | src/hooks/use-api-query.ts (NEW) | 30m |
| 2.3 | useApiMutation hook | src/hooks/use-api-mutation.ts (NEW) | 30m |
| 2.4 | Integrate master-harga (pilot) | master-harga/page.tsx | 2h |
| 2.5 | Integrate master-analisa | master-analisa/page.tsx | 2h |
| 2.6 | Refactor PekerjaanPage | components/pekerjaan/PekerjaanPage.tsx | 4h |
| 2.7 | Test 1 kategori | pekerjaan/persiapan/page.tsx | 1h |
| 2.8 | Roll out 18 kategori | 18 page files | 1h |
| 2.9 | Invoice integration | invoice/page.tsx | 2h |
| 2.10 | Logistik integration | logistik/page.tsx | 2h |
| 2.11 | Realisasi integration | realisasi/page.tsx | 2h |
| 2.12 | Rekap integration | rekap/page.tsx | 2h |

## Phase 3 — Missing APIs (Day 6-8 ~3 hari)

| # | Task | Endpoint | Est |
|---|------|----------|-----|
| 3.1 | Dashboard Stats | GET /api/dashboard/stats | 4h |
| 3.2 | Kurva-S | GET /api/proyek/[id]/kurva-s | 6h |
| 3.3 | Monitoring | GET /api/proyek/[id]/monitoring | 4h |
| 3.4 | Feedback | 5 endpoints | 4h |
| 3.5 | TimProyek | 4 endpoints | 3h |
| 3.6 | News Admin | 4 endpoints | 2h |

## Phase 4 — Integrate Remaining (Day 9-10 ~2 hari)

| # | Task | Files | Est |
|---|------|-------|-----|
| 4.1 | Dashboard page | dashboard/page.tsx | 3h |
| 4.2 | Kurva-S page | kurva-s/page.tsx | 2h |
| 4.3 | Monitoring page | monitoring/page.tsx | 2h |
| 4.4 | Feedback page | feedback/page.tsx | 3h |
| 4.5 | Admin news UI | admin/news/page.tsx (NEW) | 2h |

## Phase 5 — Security + Monitoring (Day 11 ~1 hari)

| # | Task | Est |
|---|------|-----|
| 5.1 | Rate limiting auth endpoints | 2h |
| 5.2 | Security headers next.config.ts | 30m |
| 5.3 | Setup Sentry | 30m |
| 5.4 | Setup PostHog | 20m |
| 5.5 | Enable Vercel Analytics | 5m |
| 5.6 | Add audit logging | 2h |

## Phase 6 — Deploy (Day 12 ~1 hari)

| # | Task | Est |
|---|------|-----|
| 6.1 | Pre-deploy checklist (build, lint) | 1h |
| 6.2 | Deploy to Vercel | 1h |
| 6.3 | Production smoke test | 1h |
| 6.4 | Performance check | 30m |
| 6.5 | Documentation | 30m |

---

## Success Criteria

- [ ] All 31 pages load from real API
- [ ] All CRUD operations work
- [ ] Auth complete (register → verify → login → logout)
- [ ] Calculations accurate
- [ ] Rate limiting + security headers active
- [ ] Sentry error tracking active
- [ ] Deployed to Vercel, accessible publicly
- [ ] No console errors
