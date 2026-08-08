-- =====================================================================
-- English standardization: rename tables, columns, enum types and enum
-- values from Indonesian to English. Every statement is guarded so the
-- migration is safe to re-run (e.g. after a partial failure).
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Table renames
-- ---------------------------------------------------------------------
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='proyek')
    THEN ALTER TABLE "proyek" RENAME TO "projects"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='tim_proyek')
    THEN ALTER TABLE "tim_proyek" RENAME TO "project_team"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='pekerjaan')
    THEN ALTER TABLE "pekerjaan" RENAME TO "work_items"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='detail_analisa')
    THEN ALTER TABLE "detail_analisa" RENAME TO "work_item_details"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='master_analisa')
    THEN ALTER TABLE "master_analisa" RENAME TO "analysis_masters"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='rincian_analisa')
    THEN ALTER TABLE "rincian_analisa" RENAME TO "analysis_components"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='master_harga')
    THEN ALTER TABLE "master_harga" RENAME TO "price_masters"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='rekap')
    THEN ALTER TABLE "rekap" RENAME TO "recaps"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='invoice')
    THEN ALTER TABLE "invoice" RENAME TO "invoices"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='logistik')
    THEN ALTER TABLE "logistik" RENAME TO "logistics"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename='realisasi')
    THEN ALTER TABLE "realisasi" RENAME TO "transactions"; END IF;
END $$;

-- ---------------------------------------------------------------------
-- 2. Column renames (new table names)
-- ---------------------------------------------------------------------
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='projects' AND column_name='namaProyek')
    THEN ALTER TABLE "projects" RENAME COLUMN "namaProyek" TO "name"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='projects' AND column_name='lokasi')
    THEN ALTER TABLE "projects" RENAME COLUMN "lokasi" TO "location"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='projects' AND column_name='tipe')
    THEN ALTER TABLE "projects" RENAME COLUMN "tipe" TO "type"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='projects' AND column_name='nilaiKontrak')
    THEN ALTER TABLE "projects" RENAME COLUMN "nilaiKontrak" TO "contractValue"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='namaLengkap')
    THEN ALTER TABLE "users" RENAME COLUMN "namaLengkap" TO "fullName"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='project_team' AND column_name='proyekId')
    THEN ALTER TABLE "project_team" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='proyekId')
    THEN ALTER TABLE "work_items" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='kategori')
    THEN ALTER TABLE "work_items" RENAME COLUMN "kategori" TO "category"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='uraianPekerjaan')
    THEN ALTER TABLE "work_items" RENAME COLUMN "uraianPekerjaan" TO "description"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='satuan')
    THEN ALTER TABLE "work_items" RENAME COLUMN "satuan" TO "unit"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='hargaSatuan')
    THEN ALTER TABLE "work_items" RENAME COLUMN "hargaSatuan" TO "unitPrice"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='totalBiaya')
    THEN ALTER TABLE "work_items" RENAME COLUMN "totalBiaya" TO "totalCost"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='metodeHitung')
    THEN ALTER TABLE "work_items" RENAME COLUMN "metodeHitung" TO "calculationMethod"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='levelPekerjaan')
    THEN ALTER TABLE "work_items" RENAME COLUMN "levelPekerjaan" TO "level"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='tipePekerjaan')
    THEN ALTER TABLE "work_items" RENAME COLUMN "tipePekerjaan" TO "type"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='masterAnalisaId')
    THEN ALTER TABLE "work_items" RENAME COLUMN "masterAnalisaId" TO "analysisMasterId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='waktu')
    THEN ALTER TABLE "work_items" RENAME COLUMN "waktu" TO "duration"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_items' AND column_name='totalWaktu')
    THEN ALTER TABLE "work_items" RENAME COLUMN "totalWaktu" TO "totalDuration"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='waktu')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "waktu" TO "duration"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='pekerjaanId')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "pekerjaanId" TO "workItemId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='masterHargaId')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "masterHargaId" TO "priceMasterId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='masterAnalisaId')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "masterAnalisaId" TO "analysisMasterId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='nama')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "nama" TO "name"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='satuan')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "satuan" TO "unit"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='koef')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "koef" TO "coefficient"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='hargaSatuan')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "hargaSatuan" TO "unitPrice"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='totalBiaya')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "totalBiaya" TO "totalCost"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='tipe')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "tipe" TO "type"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_item_details' AND column_name='sourceKode')
    THEN ALTER TABLE "work_item_details" RENAME COLUMN "sourceKode" TO "sourceCode"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='kode')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "kode" TO "code"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='nama')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "nama" TO "name"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='satuan')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "satuan" TO "unit"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='hargaSatuan')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "hargaSatuan" TO "unitPrice"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='kategori')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "kategori" TO "category"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='ahspKode')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "ahspKode" TO "ahspCode"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_masters' AND column_name='biayaUmum')
    THEN ALTER TABLE "analysis_masters" RENAME COLUMN "biayaUmum" TO "generalCost"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='masterAnalisaId')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "masterAnalisaId" TO "analysisMasterId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='komponenId')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "komponenId" TO "componentId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='koef')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "koef" TO "coefficient"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='tipe')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "tipe" TO "type"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='nama')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "nama" TO "name"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='satuan')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "satuan" TO "unit"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='hargaSatuan')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "hargaSatuan" TO "unitPrice"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='jumlahHarga')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "jumlahHarga" TO "totalPrice"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='kodeReferensi')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "kodeReferensi" TO "referenceCode"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analysis_components' AND column_name='urutan')
    THEN ALTER TABLE "analysis_components" RENAME COLUMN "urutan" TO "sequence"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='price_masters' AND column_name='nama')
    THEN ALTER TABLE "price_masters" RENAME COLUMN "nama" TO "name"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='price_masters' AND column_name='satuan')
    THEN ALTER TABLE "price_masters" RENAME COLUMN "satuan" TO "unit"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='price_masters' AND column_name='harga')
    THEN ALTER TABLE "price_masters" RENAME COLUMN "harga" TO "price"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='price_masters' AND column_name='kategori')
    THEN ALTER TABLE "price_masters" RENAME COLUMN "kategori" TO "type"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='price_masters' AND column_name='kodeAHSP')
    THEN ALTER TABLE "price_masters" RENAME COLUMN "kodeAHSP" TO "ahspCode"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='recaps' AND column_name='proyekId')
    THEN ALTER TABLE "recaps" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='recaps' AND column_name='kategori')
    THEN ALTER TABLE "recaps" RENAME COLUMN "kategori" TO "category"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='recaps' AND column_name='uraian')
    THEN ALTER TABLE "recaps" RENAME COLUMN "uraian" TO "description"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='recaps' AND column_name='urutan')
    THEN ALTER TABLE "recaps" RENAME COLUMN "urutan" TO "sequence"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='proyekId')
    THEN ALTER TABLE "invoices" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='nomor')
    THEN ALTER TABLE "invoices" RENAME COLUMN "nomor" TO "number"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='tanggal')
    THEN ALTER TABLE "invoices" RENAME COLUMN "tanggal" TO "date"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='proyekId')
    THEN ALTER TABLE "logistics" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='namaMaterial')
    THEN ALTER TABLE "logistics" RENAME COLUMN "namaMaterial" TO "materialName"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='satuan')
    THEN ALTER TABLE "logistics" RENAME COLUMN "satuan" TO "unit"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='hargaSatuan')
    THEN ALTER TABLE "logistics" RENAME COLUMN "hargaSatuan" TO "unitPrice"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='totalBiaya')
    THEN ALTER TABLE "logistics" RENAME COLUMN "totalBiaya" TO "totalCost"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='tanggal')
    THEN ALTER TABLE "logistics" RENAME COLUMN "tanggal" TO "date"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='logistics' AND column_name='keterangan')
    THEN ALTER TABLE "logistics" RENAME COLUMN "keterangan" TO "description"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='proyekId')
    THEN ALTER TABLE "transactions" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='tanggal')
    THEN ALTER TABLE "transactions" RENAME COLUMN "tanggal" TO "date"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='kategori')
    THEN ALTER TABLE "transactions" RENAME COLUMN "kategori" TO "category"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='jumlah')
    THEN ALTER TABLE "transactions" RENAME COLUMN "jumlah" TO "amount"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='keterangan')
    THEN ALTER TABLE "transactions" RENAME COLUMN "keterangan" TO "description"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='jenis')
    THEN ALTER TABLE "transactions" RENAME COLUMN "jenis" TO "type"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='logistikId')
    THEN ALTER TABLE "transactions" RENAME COLUMN "logistikId" TO "logisticsId"; END IF;
END $$;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='audit_log' AND column_name='proyekId')
    THEN ALTER TABLE "audit_log" RENAME COLUMN "proyekId" TO "projectId"; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='audit_log' AND column_name='pekerjaanId')
    THEN ALTER TABLE "audit_log" RENAME COLUMN "pekerjaanId" TO "workItemId"; END IF;
END $$;

-- ---------------------------------------------------------------------
-- 3. Enum conversions (values change → new type + typed column swap).
--    Runs only while the old type still exists and the new one doesn't.
-- ---------------------------------------------------------------------
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='TipeProyek')
     AND NOT EXISTS (SELECT 1 FROM pg_type WHERE typname='project_type') THEN
    ALTER TYPE "TipeProyek" RENAME TO project_type_old;
    CREATE TYPE project_type AS ENUM ('building','infrastructure');
    ALTER TABLE "projects" ALTER COLUMN "type" DROP DEFAULT;
    ALTER TABLE "projects" ALTER COLUMN "type" TYPE project_type
      USING (CASE "type"::text
        WHEN 'gedung' THEN 'building'
        WHEN 'infra' THEN 'infrastructure'
      END)::project_type;
    ALTER TABLE "projects" ALTER COLUMN "type" SET DEFAULT 'building';
    DROP TYPE project_type_old;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='KategoriPekerjaan')
     AND NOT EXISTS (SELECT 1 FROM pg_type WHERE typname='work_category') THEN
    ALTER TYPE "KategoriPekerjaan" RENAME TO work_category_old;
    CREATE TYPE work_category AS ENUM (
      'preparation','foundation','concrete','canopy','steel','stairs','roof',
      'wall','plastering','finishing','tiles','paving','painting',
      'doors','interior','toilet','mep','custom'
    );
    ALTER TABLE "work_items" ALTER COLUMN "category" TYPE work_category
      USING (CASE "category"::text
        WHEN 'persiapan' THEN 'preparation'
        WHEN 'pondasi' THEN 'foundation'
        WHEN 'beton' THEN 'concrete'
        WHEN 'kanopi' THEN 'canopy'
        WHEN 'baja' THEN 'steel'
        WHEN 'tangga' THEN 'stairs'
        WHEN 'atap' THEN 'roof'
        WHEN 'dinding' THEN 'wall'
        WHEN 'plesteran' THEN 'plastering'
        WHEN 'acian' THEN 'finishing'
        WHEN 'keramik' THEN 'tiles'
        WHEN 'paving' THEN 'paving'
        WHEN 'pengecatan' THEN 'painting'
        WHEN 'pintu' THEN 'doors'
        WHEN 'interior' THEN 'interior'
        WHEN 'toilet' THEN 'toilet'
        WHEN 'mep' THEN 'mep'
        WHEN 'custom' THEN 'custom'
      END)::work_category;
    DROP TYPE work_category_old;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='MetodeHitung')
     AND NOT EXISTS (SELECT 1 FROM pg_type WHERE typname='calculation_method') THEN
    ALTER TYPE "MetodeHitung" RENAME TO calculation_method_old;
    CREATE TYPE calculation_method AS ENUM (
      'ahsp','manual','lump_sum','manual_price','custom_price'
    );
    ALTER TABLE "work_items" ALTER COLUMN "calculationMethod" TYPE calculation_method
      USING (CASE "calculationMethod"::text
        WHEN 'ahsp' THEN 'ahsp'
        WHEN 'manual' THEN 'manual'
        WHEN 'harga_borong' THEN 'lump_sum'
        WHEN 'harga_manual' THEN 'manual_price'
        WHEN 'harga_custom' THEN 'custom_price'
      END)::calculation_method;
    DROP TYPE calculation_method_old;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='TipeKomponen')
     AND NOT EXISTS (SELECT 1 FROM pg_type WHERE typname='component_type') THEN
    ALTER TYPE "TipeKomponen" RENAME TO component_type_old;
    CREATE TYPE component_type AS ENUM ('material','labor','equipment');
    ALTER TABLE "work_item_details" ALTER COLUMN "type" TYPE component_type
      USING (CASE "type"::text
        WHEN 'material' THEN 'material'
        WHEN 'upah' THEN 'labor'
        WHEN 'alat' THEN 'equipment'
      END)::component_type;
    ALTER TABLE "analysis_components" ALTER COLUMN "type" TYPE component_type
      USING (CASE "type"::text
        WHEN 'material' THEN 'material'
        WHEN 'upah' THEN 'labor'
        WHEN 'alat' THEN 'equipment'
      END)::component_type;
    ALTER TABLE "price_masters" ALTER COLUMN "type" TYPE component_type
      USING (CASE "type"::text
        WHEN 'material' THEN 'material'
        WHEN 'upah' THEN 'labor'
        WHEN 'alat' THEN 'equipment'
      END)::component_type;
    DROP TYPE component_type_old;
  END IF;
END $$;

-- ---------------------------------------------------------------------
-- 4. Enum type renames (values unchanged)
-- ---------------------------------------------------------------------
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='RoleTimProyek')
    THEN ALTER TYPE "RoleTimProyek" RENAME TO team_role; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='StatusInvoice')
    THEN ALTER TYPE "StatusInvoice" RENAME TO invoice_status; END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname='StatusFeedback')
    THEN ALTER TYPE "StatusFeedback" RENAME TO feedback_status; END IF;
END $$;

-- ---------------------------------------------------------------------
-- 5. Text value updates (transactions type, TEXT column)
-- ---------------------------------------------------------------------
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM "transactions" WHERE "type" = 'pengeluaran') THEN
    UPDATE "transactions" SET "type" = 'expense' WHERE "type" = 'pengeluaran';
  END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM "transactions" WHERE "type" = 'pemasukan') THEN
    UPDATE "transactions" SET "type" = 'income' WHERE "type" = 'pemasukan';
  END IF;
END $$;
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='type'
             AND column_default LIKE '%pengeluaran%') THEN
    ALTER TABLE "transactions" ALTER COLUMN "type" SET DEFAULT 'expense';
  END IF;
END $$;
