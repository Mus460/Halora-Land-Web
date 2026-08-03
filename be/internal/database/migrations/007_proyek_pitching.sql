-- Status proyek dinyatakan sebagai boolean: "isPitching" = true artinya proyek
-- sedang dalam tahap penawaran/negosiasi, false artinya proyek aktif dikerjakan.
-- Idempotent: safe to re-run.
ALTER TABLE "proyek" ADD COLUMN IF NOT EXISTS "isPitching" BOOLEAN NOT NULL DEFAULT false;