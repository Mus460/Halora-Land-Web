-- Soft deletion: user-facing entities get a deletedAt tombstone instead of
-- being physically removed. detail_analisa (recalc snapshots) and AHSP import
-- replacement rows intentionally remain hard-deleted.
ALTER TABLE "proyek" ADD COLUMN IF NOT EXISTS "deletedAt" TIMESTAMP(3);
ALTER TABLE "pekerjaan" ADD COLUMN IF NOT EXISTS "deletedAt" TIMESTAMP(3);
ALTER TABLE "master_harga" ADD COLUMN IF NOT EXISTS "deletedAt" TIMESTAMP(3);
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "deletedAt" TIMESTAMP(3);
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "deletedAt" TIMESTAMP(3);
ALTER TABLE "news" ADD COLUMN IF NOT EXISTS "deletedAt" TIMESTAMP(3);

-- Unique indexes must ignore soft-deleted rows so re-creating the same
-- nama/kode after a delete is allowed again. The full-index versions were
-- removed from 001_schema.sql; this migration is the sole owner of the
-- partial indexes and swaps them in on upgrade from the old 001.
DROP INDEX IF EXISTS "master_harga_nama_userId_kategori_key";
CREATE UNIQUE INDEX IF NOT EXISTS "master_harga_nama_userId_kategori_active_key"
    ON "master_harga"("nama", "userId", "kategori") WHERE "deletedAt" IS NULL;

DROP INDEX IF EXISTS "master_analisa_kode_userId_key";
CREATE UNIQUE INDEX IF NOT EXISTS "master_analisa_kode_userId_active_key"
    ON "master_analisa"("kode", "userId") WHERE "deletedAt" IS NULL;
