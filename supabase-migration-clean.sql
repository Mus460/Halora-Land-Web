-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "public";

-- CreateEnum
CREATE TYPE "Role" AS ENUM ('ADMIN', 'OWNER', 'USER', 'DEMO');

-- CreateEnum
CREATE TYPE "TipeProyek" AS ENUM ('gedung', 'infra');

-- CreateEnum
CREATE TYPE "RoleTimProyek" AS ENUM ('owner', 'editor', 'viewer');

-- CreateEnum
CREATE TYPE "KategoriPekerjaan" AS ENUM ('persiapan', 'pondasi', 'beton', 'kanopi', 'baja', 'tangga', 'atap', 'dinding', 'plesteran', 'acian', 'keramik', 'paving', 'pengecatan', 'pintu', 'interior', 'toilet', 'mep', 'custom');

-- CreateEnum
CREATE TYPE "MetodeHitung" AS ENUM ('ahsp', 'manual', 'harga_borong', 'harga_manual', 'harga_custom');

-- CreateEnum
CREATE TYPE "TipeKomponen" AS ENUM ('material', 'upah', 'alat');

-- CreateEnum
CREATE TYPE "StatusInvoice" AS ENUM ('draft', 'sent', 'paid');

-- CreateEnum
CREATE TYPE "StatusFeedback" AS ENUM ('open', 'in_progress', 'resolved', 'closed');

-- CreateTable
CREATE TABLE "users" (
    "id" SERIAL NOT NULL,
    "namaLengkap" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "role" "Role" NOT NULL DEFAULT 'USER',
    "accountType" TEXT NOT NULL DEFAULT 'free',
    "isDemo" BOOLEAN NOT NULL DEFAULT false,
    "supabaseAuthId" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "proyek" (
    "id" SERIAL NOT NULL,
    "userId" INTEGER NOT NULL,
    "namaProyek" TEXT NOT NULL,
    "lokasi" TEXT,
    "tipe" "TipeProyek" NOT NULL DEFAULT 'gedung',
    "nilaiKontrak" DECIMAL(15,2),
    "timeline" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "proyek_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "tim_proyek" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER NOT NULL,
    "userId" INTEGER NOT NULL,
    "role" "RoleTimProyek" NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "tim_proyek_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "pekerjaan" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER NOT NULL,
    "kategori" "KategoriPekerjaan" NOT NULL,
    "uraianPekerjaan" TEXT NOT NULL,
    "volume" DECIMAL(15,4) NOT NULL,
    "satuan" TEXT NOT NULL,
    "hargaSatuan" DECIMAL(15,2) NOT NULL,
    "totalBiaya" DECIMAL(15,2) NOT NULL,
    "metodeHitung" "MetodeHitung" NOT NULL,
    "levelPekerjaan" TEXT,
    "tipePekerjaan" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "pekerjaan_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "detail_analisa" (
    "id" SERIAL NOT NULL,
    "pekerjaanId" INTEGER NOT NULL,
    "masterHargaId" INTEGER,
    "masterAnalisaId" INTEGER,
    "nama" TEXT NOT NULL,
    "satuan" TEXT NOT NULL,
    "koef" DECIMAL(10,6) NOT NULL,
    "hargaSatuan" DECIMAL(15,2) NOT NULL,
    "totalBiaya" DECIMAL(15,2) NOT NULL,
    "tipe" "TipeKomponen" NOT NULL,
    "snapshotAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "sourceKode" TEXT,

    CONSTRAINT "detail_analisa_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "master_analisa" (
    "id" SERIAL NOT NULL,
    "kode" TEXT NOT NULL,
    "nama" TEXT NOT NULL,
    "level" INTEGER NOT NULL,
    "parentId" INTEGER,
    "satuan" TEXT,
    "isGlobal" BOOLEAN NOT NULL DEFAULT false,
    "userId" INTEGER,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "master_analisa_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "rincian_analisa" (
    "id" SERIAL NOT NULL,
    "masterAnalisaId" INTEGER NOT NULL,
    "komponenId" INTEGER NOT NULL,
    "koef" DECIMAL(10,6) NOT NULL,
    "tipe" "TipeKomponen" NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "rincian_analisa_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "master_harga" (
    "id" SERIAL NOT NULL,
    "nama" TEXT NOT NULL,
    "satuan" TEXT NOT NULL,
    "harga" DECIMAL(15,2) NOT NULL,
    "kategori" "TipeKomponen" NOT NULL,
    "isGlobal" BOOLEAN NOT NULL DEFAULT false,
    "userId" INTEGER,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "master_harga_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "rekap" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER NOT NULL,
    "kategori" TEXT NOT NULL,
    "uraian" TEXT NOT NULL,
    "urutan" INTEGER NOT NULL,
    "margin" DECIMAL(5,2),
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "rekap_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "invoice" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER NOT NULL,
    "nomor" TEXT NOT NULL,
    "tanggal" TIMESTAMP(3) NOT NULL,
    "total" DECIMAL(15,2) NOT NULL,
    "status" "StatusInvoice" NOT NULL DEFAULT 'draft',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "invoice_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "logistik" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER NOT NULL,
    "namaMaterial" TEXT NOT NULL,
    "satuan" TEXT NOT NULL,
    "volume" DECIMAL(15,4) NOT NULL,
    "hargaSatuan" DECIMAL(15,2) NOT NULL,
    "totalBiaya" DECIMAL(15,2) NOT NULL,
    "tanggal" TIMESTAMP(3),
    "keterangan" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "logistik_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "realisasi" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER NOT NULL,
    "tanggal" TIMESTAMP(3) NOT NULL,
    "kategori" TEXT NOT NULL,
    "jumlah" DECIMAL(15,2) NOT NULL,
    "keterangan" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "realisasi_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "feedback" (
    "id" SERIAL NOT NULL,
    "userId" INTEGER NOT NULL,
    "subject" TEXT NOT NULL,
    "message" TEXT NOT NULL,
    "status" "StatusFeedback" NOT NULL DEFAULT 'open',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "feedback_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "feedback_reply" (
    "id" SERIAL NOT NULL,
    "feedbackId" INTEGER NOT NULL,
    "userId" INTEGER NOT NULL,
    "message" TEXT NOT NULL,
    "isAdmin" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "feedback_reply_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "news" (
    "id" SERIAL NOT NULL,
    "title" TEXT NOT NULL,
    "content" TEXT NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "news_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "audit_log" (
    "id" SERIAL NOT NULL,
    "proyekId" INTEGER,
    "pekerjaanId" INTEGER,
    "userId" INTEGER NOT NULL,
    "action" TEXT NOT NULL,
    "entityType" TEXT NOT NULL,
    "entityId" INTEGER,
    "oldValue" JSONB,
    "newValue" JSONB,
    "description" TEXT,
    "ipAddress" TEXT,
    "userAgent" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "audit_log_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "users_email_key" ON "users"("email");

-- CreateIndex
CREATE UNIQUE INDEX "users_supabaseAuthId_key" ON "users"("supabaseAuthId");

-- CreateIndex
CREATE INDEX "tim_proyek_userId_idx" ON "tim_proyek"("userId");

-- CreateIndex
CREATE INDEX "tim_proyek_proyekId_idx" ON "tim_proyek"("proyekId");

-- CreateIndex
CREATE UNIQUE INDEX "tim_proyek_proyekId_userId_key" ON "tim_proyek"("proyekId", "userId");

-- CreateIndex
CREATE INDEX "pekerjaan_proyekId_idx" ON "pekerjaan"("proyekId");

-- CreateIndex
CREATE INDEX "pekerjaan_kategori_idx" ON "pekerjaan"("kategori");

-- CreateIndex
CREATE INDEX "detail_analisa_pekerjaanId_idx" ON "detail_analisa"("pekerjaanId");

-- CreateIndex
CREATE INDEX "detail_analisa_masterHargaId_idx" ON "detail_analisa"("masterHargaId");

-- CreateIndex
CREATE INDEX "detail_analisa_masterAnalisaId_idx" ON "detail_analisa"("masterAnalisaId");

-- CreateIndex
CREATE INDEX "master_analisa_userId_idx" ON "master_analisa"("userId");

-- CreateIndex
CREATE INDEX "master_analisa_parentId_idx" ON "master_analisa"("parentId");

-- CreateIndex
CREATE UNIQUE INDEX "master_analisa_kode_userId_key" ON "master_analisa"("kode", "userId");

-- CreateIndex
CREATE INDEX "rincian_analisa_masterAnalisaId_idx" ON "rincian_analisa"("masterAnalisaId");

-- CreateIndex
CREATE INDEX "rincian_analisa_komponenId_idx" ON "rincian_analisa"("komponenId");

-- CreateIndex
CREATE INDEX "master_harga_userId_idx" ON "master_harga"("userId");

-- CreateIndex
CREATE UNIQUE INDEX "master_harga_nama_userId_kategori_key" ON "master_harga"("nama", "userId", "kategori");

-- CreateIndex
CREATE INDEX "rekap_proyekId_idx" ON "rekap"("proyekId");

-- CreateIndex
CREATE UNIQUE INDEX "invoice_nomor_key" ON "invoice"("nomor");

-- CreateIndex
CREATE INDEX "invoice_proyekId_idx" ON "invoice"("proyekId");

-- CreateIndex
CREATE INDEX "logistik_proyekId_idx" ON "logistik"("proyekId");

-- CreateIndex
CREATE INDEX "realisasi_proyekId_idx" ON "realisasi"("proyekId");

-- CreateIndex
CREATE INDEX "feedback_userId_idx" ON "feedback"("userId");

-- CreateIndex
CREATE INDEX "feedback_reply_feedbackId_idx" ON "feedback_reply"("feedbackId");

-- CreateIndex
CREATE INDEX "feedback_reply_userId_idx" ON "feedback_reply"("userId");

-- CreateIndex
CREATE INDEX "audit_log_proyekId_idx" ON "audit_log"("proyekId");

-- CreateIndex
CREATE INDEX "audit_log_userId_idx" ON "audit_log"("userId");

-- CreateIndex
CREATE INDEX "audit_log_action_idx" ON "audit_log"("action");

-- CreateIndex
CREATE INDEX "audit_log_createdAt_idx" ON "audit_log"("createdAt");

-- AddForeignKey
ALTER TABLE "proyek" ADD CONSTRAINT "proyek_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "tim_proyek" ADD CONSTRAINT "tim_proyek_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "tim_proyek" ADD CONSTRAINT "tim_proyek_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "pekerjaan" ADD CONSTRAINT "pekerjaan_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "detail_analisa" ADD CONSTRAINT "detail_analisa_pekerjaanId_fkey" FOREIGN KEY ("pekerjaanId") REFERENCES "pekerjaan"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "detail_analisa" ADD CONSTRAINT "detail_analisa_masterHargaId_fkey" FOREIGN KEY ("masterHargaId") REFERENCES "master_harga"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "detail_analisa" ADD CONSTRAINT "detail_analisa_masterAnalisaId_fkey" FOREIGN KEY ("masterAnalisaId") REFERENCES "master_analisa"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "master_analisa" ADD CONSTRAINT "master_analisa_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "master_analisa" ADD CONSTRAINT "master_analisa_parentId_fkey" FOREIGN KEY ("parentId") REFERENCES "master_analisa"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "rincian_analisa" ADD CONSTRAINT "rincian_analisa_masterAnalisaId_fkey" FOREIGN KEY ("masterAnalisaId") REFERENCES "master_analisa"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "rincian_analisa" ADD CONSTRAINT "rincian_analisa_komponenId_fkey" FOREIGN KEY ("komponenId") REFERENCES "master_harga"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "master_harga" ADD CONSTRAINT "master_harga_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "rekap" ADD CONSTRAINT "rekap_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "invoice" ADD CONSTRAINT "invoice_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "logistik" ADD CONSTRAINT "logistik_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "realisasi" ADD CONSTRAINT "realisasi_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "feedback" ADD CONSTRAINT "feedback_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "feedback_reply" ADD CONSTRAINT "feedback_reply_feedbackId_fkey" FOREIGN KEY ("feedbackId") REFERENCES "feedback"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "feedback_reply" ADD CONSTRAINT "feedback_reply_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "audit_log" ADD CONSTRAINT "audit_log_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "audit_log" ADD CONSTRAINT "audit_log_proyekId_fkey" FOREIGN KEY ("proyekId") REFERENCES "proyek"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "audit_log" ADD CONSTRAINT "audit_log_pekerjaanId_fkey" FOREIGN KEY ("pekerjaanId") REFERENCES "pekerjaan"("id") ON DELETE CASCADE ON UPDATE CASCADE;

