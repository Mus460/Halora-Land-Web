-- Estimasi waktu per item pekerjaan: tautan ke master analisa sumber + koefisien
-- waktu (jam per satuan, diambil dari SUM(waktu) rincian "Tukang batu").
-- Idempotent: safe to re-run.
ALTER TABLE "pekerjaan" ADD COLUMN IF NOT EXISTS "masterAnalisaId" INTEGER REFERENCES "master_analisa"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "pekerjaan" ADD COLUMN IF NOT EXISTS "waktu" DECIMAL(12,4);
