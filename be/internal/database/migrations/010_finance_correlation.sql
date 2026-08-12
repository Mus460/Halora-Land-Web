-- Finance correlation: realisasi entries can be income (pemasukan) or
-- expense (pengeluaran), carry an approval lifecycle, and trace back to
-- their source (logistik purchase or invoice).
ALTER TABLE "realisasi" ADD COLUMN IF NOT EXISTS "jenis" TEXT NOT NULL DEFAULT 'pengeluaran';
ALTER TABLE "realisasi" ADD COLUMN IF NOT EXISTS "status" TEXT NOT NULL DEFAULT 'approved';
ALTER TABLE "realisasi" ADD COLUMN IF NOT EXISTS "logistikId" INTEGER;
ALTER TABLE "realisasi" ADD COLUMN IF NOT EXISTS "invoiceId" INTEGER;
