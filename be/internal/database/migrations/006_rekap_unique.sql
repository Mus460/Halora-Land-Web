-- Satu baris settings per proyek: unik pada (proyekId, kategori) agar
-- UpsertMargin dapat memakai ON CONFLICT.
-- Idempotent: safe to re-run.
CREATE UNIQUE INDEX IF NOT EXISTS "rekap_proyek_kategori_unique" ON "rekap" ("proyekId", "kategori");
