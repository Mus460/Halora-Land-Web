-- Backfill koefisien waktu (jam per satuan = 1/koefisien) ke semua master analisa:
-- - baris "Tukang batu" tetap jadi acuan bila ada (konsisten dengan import awal),
-- - selain itu dipakai baris tenaga kerja utama (koef upah terbesar).
-- Total waktu per item = SUM(waktu) rincian, lalu disalin ke pekerjaan yang sudah ada.
-- Idempotent: safe to re-run.
UPDATE "rincian_analisa" r SET waktu = ROUND(1.0 / r.koef, 4)
WHERE r.waktu IS NULL
  AND r.koef > 0
  AND r.tipe = 'upah'
  AND r.id = (
    SELECT r2.id FROM "rincian_analisa" r2
    WHERE r2."masterAnalisaId" = r."masterAnalisaId"
      AND r2.tipe = 'upah'
      AND r2.koef > 0
    ORDER BY (r2.nama ILIKE '%tukang batu%') DESC, r2.koef DESC, r2.id ASC
    LIMIT 1
  );

UPDATE "pekerjaan" p SET waktu = sub.w
FROM (
  SELECT "masterAnalisaId", SUM(waktu) AS w
  FROM "rincian_analisa" WHERE waktu IS NOT NULL
  GROUP BY "masterAnalisaId"
) sub
WHERE p."masterAnalisaId" = sub."masterAnalisaId" AND p.waktu IS NULL;

-- Pekerjaan lama tanpa tautan: cocokkan persis uraian dengan master analisa sistem.
UPDATE "pekerjaan" p SET "masterAnalisaId" = ma.id, waktu = sub.w
FROM "master_analisa" ma
JOIN (
  SELECT "masterAnalisaId", SUM(waktu) AS w
  FROM "rincian_analisa" WHERE waktu IS NOT NULL
  GROUP BY "masterAnalisaId"
) sub ON sub."masterAnalisaId" = ma.id
WHERE p."masterAnalisaId" IS NULL AND p.waktu IS NULL
  AND ma."isSystem" = true
  AND LOWER(BTRIM(p."uraianPekerjaan")) = LOWER(BTRIM(ma.nama))
  AND ma.id = (
    SELECT ma2.id FROM "master_analisa" ma2
    WHERE LOWER(BTRIM(ma2.nama)) = LOWER(BTRIM(p."uraianPekerjaan")) AND ma2."isSystem" = true
    ORDER BY ma2.id ASC LIMIT 1
  );
