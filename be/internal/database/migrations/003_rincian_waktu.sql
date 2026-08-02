-- waktu (jam kerja = 1/koefisien) untuk rincian_analisa, khusus baris "Tukang batu".
-- Idempotent: safe to re-run.
ALTER TABLE "rincian_analisa" ADD COLUMN IF NOT EXISTS "waktu" DECIMAL(12,4);
