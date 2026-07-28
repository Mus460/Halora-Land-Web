-- =====================================================================
-- Halora Land — consolidated schema (clean baseline + AHSP delta merged)
-- Target: PostgreSQL >= 14.  Runnable on a fresh DB.
-- Per ARCHITECTURE.md §5.2 + §3.3 (users.password dropped in rework).
-- =====================================================================

BEGIN;

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
    "masterHargaId"   INTEGER,
    "masterAnalisaId" INTEGER,
    "nama"            TEXT    NOT NULL,
    "satuan"          TEXT    NOT NULL,
    "koef"            DECIMAL(10,6) NOT NULL,
    "hargaSatuan"     DECIMAL(15,2) NOT NULL,
    "totalBiaya"      DECIMAL(15,2) NOT NULL,
    "tipe"            "TipeKomponen" NOT NULL,
    "snapshotAt"      TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "sourceKode"      TEXT,
    CONSTRAINT "detail_analisa_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "master_analisa" (
    "id"          SERIAL NOT NULL,
    "kode"        TEXT    NOT NULL,
    "nama"        TEXT    NOT NULL,
    "level"       INTEGER NOT NULL,
    "parentId"    INTEGER,
    "satuan"      TEXT,
    "hargaSatuan" DECIMAL(15,2),
    "kategori"    TEXT,
    "isGlobal"    BOOLEAN NOT NULL DEFAULT false,
    "userId"      INTEGER,
    "isSystem"    BOOLEAN NOT NULL DEFAULT false,
    "ahspKode"    TEXT,
    "ahspSheet"   TEXT,
    "biayaUmum"   DECIMAL(5,4) NOT NULL DEFAULT 0.10,
    "createdAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "master_analisa_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "rincian_analisa" (
    "id"             SERIAL NOT NULL,
    "masterAnalisaId" INTEGER NOT NULL,
    "komponenId"     INTEGER,
    "koef"           DECIMAL(10,6) NOT NULL,
    "tipe"           "TipeKomponen" NOT NULL,
    "nama"           TEXT,
    "satuan"         TEXT,
    "hargaSatuan"    DECIMAL(15,2),
    "jumlahHarga"    DECIMAL(15,2),
    "kodeReferensi"  TEXT,
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
    "kategori"  TEXT    NOT NULL,
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
    "nomor"     TEXT    NOT NULL,
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
    "totalBiaya"   DECIMAL(15,2) NOT NULL,
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
    "action"      TEXT    NOT NULL,
    "entityType"  TEXT    NOT NULL,
    "entityId"    INTEGER,
    "oldValue"    JSONB,
    "newValue"    JSONB,
    "description" TEXT,
    "ipAddress"   TEXT,
    "userAgent"   TEXT,
    "createdAt"   TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "audit_log_pkey" PRIMARY KEY ("id")
);

-- ---------------------------------------------------------------------
-- Indexes (unique)
-- ---------------------------------------------------------------------
CREATE UNIQUE INDEX "users_email_key"                       ON "users"("email");
CREATE UNIQUE INDEX "users_supabaseAuthId_key"              ON "users"("supabaseAuthId");
CREATE UNIQUE INDEX "tim_proyek_proyekId_userId_key"        ON "tim_proyek"("proyekId","userId");
CREATE UNIQUE INDEX "master_analisa_kode_userId_key"        ON "master_analisa"("kode","userId");
CREATE UNIQUE INDEX "master_harga_nama_userId_kategori_key" ON "master_harga"("nama","userId","kategori");
CREATE UNIQUE INDEX "invoice_nomor_key"                     ON "invoice"("nomor");

-- ---------------------------------------------------------------------
-- Indexes (non-unique)
-- ---------------------------------------------------------------------
CREATE INDEX "tim_proyek_userId_idx"                ON "tim_proyek"("userId");
CREATE INDEX "tim_proyek_proyekId_idx"              ON "tim_proyek"("proyekId");
CREATE INDEX "pekerjaan_proyekId_idx"               ON "pekerjaan"("proyekId");
CREATE INDEX "pekerjaan_kategori_idx"               ON "pekerjaan"("kategori");
CREATE INDEX "detail_analisa_pekerjaanId_idx"       ON "detail_analisa"("pekerjaanId");
CREATE INDEX "detail_analisa_masterHargaId_idx"     ON "detail_analisa"("masterHargaId");
CREATE INDEX "detail_analisa_masterAnalisaId_idx"   ON "detail_analisa"("masterAnalisaId");
CREATE INDEX "master_analisa_userId_idx"            ON "master_analisa"("userId");
CREATE INDEX "master_analisa_parentId_idx"          ON "master_analisa"("parentId");
CREATE INDEX "rincian_analisa_masterAnalisaId_idx"  ON "rincian_analisa"("masterAnalisaId");
CREATE INDEX "rincian_analisa_komponenId_idx"       ON "rincian_analisa"("komponenId");
CREATE INDEX "master_harga_userId_idx"              ON "master_harga"("userId");
CREATE INDEX "rekap_proyekId_idx"                   ON "rekap"("proyekId");
CREATE INDEX "invoice_proyekId_idx"                 ON "invoice"("proyekId");
CREATE INDEX "logistik_proyekId_idx"                ON "logistik"("proyekId");
CREATE INDEX "realisasi_proyekId_idx"               ON "realisasi"("proyekId");
CREATE INDEX "feedback_userId_idx"                  ON "feedback"("userId");
CREATE INDEX "feedback_reply_feedbackId_idx"        ON "feedback_reply"("feedbackId");
CREATE INDEX "feedback_reply_userId_idx"            ON "feedback_reply"("userId");
CREATE INDEX "audit_log_proyekId_idx"               ON "audit_log"("proyekId");
CREATE INDEX "audit_log_userId_idx"                 ON "audit_log"("userId");
CREATE INDEX "audit_log_action_idx"                 ON "audit_log"("action");
CREATE INDEX "audit_log_createdAt_idx"              ON "audit_log"("createdAt");

CREATE INDEX "idx_master_analisa_kategori_is_system" ON "master_analisa"("kategori","isSystem");
CREATE INDEX "idx_master_analisa_ahsp_kode"          ON "master_analisa"("ahspKode");
CREATE INDEX "idx_master_analisa_ahsp_sheet"         ON "master_analisa"("ahspSheet");
CREATE INDEX "idx_master_harga_kode_ahsp"            ON "master_harga"("kodeAHSP");
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

ALTER TABLE "detail_analisa"
  ADD CONSTRAINT "detail_analisa_masterHargaId_fkey"
  FOREIGN KEY ("masterHargaId") REFERENCES "master_harga"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "detail_analisa"
  ADD CONSTRAINT "detail_analisa_masterAnalisaId_fkey"
  FOREIGN KEY ("masterAnalisaId") REFERENCES "master_analisa"("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "master_analisa"
  ADD CONSTRAINT "master_analisa_userId_fkey"
  FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

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

COMMENT ON COLUMN "master_analisa"."isSystem"  IS 'true = from AHSP 2026, false = user custom';
COMMENT ON COLUMN "master_analisa"."ahspKode"  IS 'Original AHSP code e.g. 2.2.1.1.1';
COMMENT ON COLUMN "master_analisa"."ahspSheet" IS 'Source Excel sheet name e.g. Beton';
COMMENT ON COLUMN "master_analisa"."biayaUmum" IS 'Overhead percentage e.g. 0.10 = 10%';

COMMIT;
