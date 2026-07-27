-- Add AHSP support fields to master_analisa
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "harga_satuan" DECIMAL(15,2);
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "kategori" TEXT;
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "ahsp_kode" TEXT;
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "ahsp_sheet" TEXT;
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "biaya_umum" DECIMAL(5,4) DEFAULT 0.10;
ALTER TABLE "master_analisa" ADD COLUMN IF NOT EXISTS "is_system" BOOLEAN DEFAULT false;

-- Add indexes for AHSP fields
CREATE INDEX IF NOT EXISTS "idx_master_analisa_kategori_is_system" ON "master_analisa"("kategori", "is_system");
CREATE INDEX IF NOT EXISTS "idx_master_analisa_ahsp_kode" ON "master_analisa"("ahsp_kode");
CREATE INDEX IF NOT EXISTS "idx_master_analisa_ahsp_sheet" ON "master_analisa"("ahsp_sheet");

-- Add AHSP support to master_harga
ALTER TABLE "master_harga" ADD COLUMN IF NOT EXISTS "kode_ahsp" TEXT;
ALTER TABLE "master_harga" ADD COLUMN IF NOT EXISTS "is_system" BOOLEAN DEFAULT false;

CREATE INDEX IF NOT EXISTS "idx_master_harga_kode_ahsp" ON "master_harga"("kode_ahsp");

-- Add snapshot fields to rincian_analisa (for AHSP breakdown storage)
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "nama" TEXT;
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "satuan" TEXT;
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "harga_satuan" DECIMAL(15,2);
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "jumlah_harga" DECIMAL(15,2);
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "kode_referensi" TEXT;
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "urutan" INTEGER DEFAULT 0;

-- Make komponenId nullable (for snapshot-only records)
ALTER TABLE "rincian_analisa" ALTER COLUMN "komponen_id" DROP NOT NULL;

-- Enable PostgreSQL full-text search extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create trigram indexes for fuzzy search
CREATE INDEX IF NOT EXISTS "idx_master_analisa_nama_trgm" ON "master_analisa" USING gin ("nama" gin_trgm_ops);
CREATE INDEX IF NOT EXISTS "idx_master_analisa_ahsp_kode_trgm" ON "master_analisa" USING gin ("ahsp_kode" gin_trgm_ops);

-- Add comment for documentation
COMMENT ON COLUMN "master_analisa"."is_system" IS 'true = from AHSP 2026, false = user custom';
COMMENT ON COLUMN "master_analisa"."ahsp_kode" IS 'Original AHSP code e.g. 2.2.1.1.1';
COMMENT ON COLUMN "master_analisa"."ahsp_sheet" IS 'Source Excel sheet name e.g. Beton';
COMMENT ON COLUMN "master_analisa"."biaya_umum" IS 'Overhead percentage e.g. 0.10 = 10%';
