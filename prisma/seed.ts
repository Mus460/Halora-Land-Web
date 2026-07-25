import { PrismaClient } from '@prisma/client'
import { PrismaPg } from '@prisma/adapter-pg'
import { Pool } from 'pg'
import { hash } from 'bcryptjs'

const connectionString = process.env.DATABASE_URL || 'postgresql://postgres:postgres@localhost:5432/template1'
const pool = new Pool({ connectionString })
const adapter = new PrismaPg(pool)
const prisma = new PrismaClient({ adapter })

async function main() {
  console.log('🌱 Starting seed...')

  // Clear existing data
  await prisma.feedbackReply.deleteMany()
  await prisma.feedback.deleteMany()
  await prisma.news.deleteMany()
  await prisma.realisasi.deleteMany()
  await prisma.logistik.deleteMany()
  await prisma.invoice.deleteMany()
  await prisma.rekap.deleteMany()
  await prisma.detailAnalisa.deleteMany()
  await prisma.pekerjaan.deleteMany()
  await prisma.rincianAnalisa.deleteMany()
  await prisma.masterAnalisa.deleteMany()
  await prisma.masterHarga.deleteMany()
  await prisma.timProyek.deleteMany()
  await prisma.proyek.deleteMany()
  await prisma.user.deleteMany()

  console.log('✅ Cleared existing data')

  // Create users
  const hashedPassword = await hash('password123', 12)

  const admin = await prisma.user.create({
    data: {
      namaLengkap: 'Admin Utama',
      email: 'admin@haloraland.id',
      password: hashedPassword,
      role: 'ADMIN',
      accountType: 'pro',
      isDemo: false,
    },
  })

  const user1 = await prisma.user.create({
    data: {
      namaLengkap: 'Budi Kontraktor',
      email: 'budi@example.com',
      password: hashedPassword,
      role: 'USER',
      accountType: 'free',
      isDemo: false,
    },
  })

  const demoUser = await prisma.user.create({
    data: {
      namaLengkap: 'Demo User',
      email: 'demo@haloraland.id',
      password: hashedPassword,
      role: 'DEMO',
      accountType: 'free',
      isDemo: true,
    },
  })

  console.log('✅ Created users')

  // Create projects
  const proyek1 = await prisma.proyek.create({
    data: {
      userId: user1.id,
      namaProyek: 'Pembangunan Rumah 2 Lantai',
      lokasi: 'Jakarta Selatan',
      tipe: 'gedung',
      nilaiKontrak: 850000000,
      timeline: '6 bulan',
    },
  })

  const proyek2 = await prisma.proyek.create({
    data: {
      userId: user1.id,
      namaProyek: 'Renovasi Kantor',
      lokasi: 'Tangerang',
      tipe: 'gedung',
      nilaiKontrak: 450000000,
      timeline: '3 bulan',
    },
  })

  console.log('✅ Created projects')

  // Create sample master harga
  const semen = await prisma.masterHarga.create({
    data: {
      nama: 'Semen Portland',
      satuan: 'zak',
      harga: 75000,
      kategori: 'material',
      isGlobal: true,
    },
  })

  const pasir = await prisma.masterHarga.create({
    data: {
      nama: 'Pasir Pasang',
      satuan: 'm3',
      harga: 350000,
      kategori: 'material',
      isGlobal: true,
    },
  })

  const tukang = await prisma.masterHarga.create({
    data: {
      nama: 'Tukang Batu',
      satuan: 'hari',
      harga: 150000,
      kategori: 'upah',
      isGlobal: true,
    },
  })

  console.log('✅ Created master harga')

  // Create sample pekerjaan
  const pekerjaan1 = await prisma.pekerjaan.create({
    data: {
      proyekId: proyek1.id,
      kategori: 'pondasi',
      uraianPekerjaan: 'Galian Tanah Pondasi',
      volume: 25.5,
      satuan: 'm3',
      hargaSatuan: 125000,
      totalBiaya: 3187500,
      metodeHitung: 'manual',
    },
  })

  const pekerjaan2 = await prisma.pekerjaan.create({
    data: {
      proyekId: proyek1.id,
      kategori: 'beton',
      uraianPekerjaan: 'Pasang Bata Merah',
      volume: 120,
      satuan: 'm2',
      hargaSatuan: 185000,
      totalBiaya: 22200000,
      metodeHitung: 'ahsp',
    },
  })

  console.log('✅ Created pekerjaan')

  // Create detail analisa
  await prisma.detailAnalisa.createMany({
    data: [
      {
        pekerjaanId: pekerjaan2.id,
        masterHargaId: semen.id,
        nama: 'Semen Portland',
        satuan: 'zak',
        koef: 1.5,
        hargaSatuan: 75000,
        totalBiaya: 112500,
        tipe: 'material',
      },
      {
        pekerjaanId: pekerjaan2.id,
        masterHargaId: pasir.id,
        nama: 'Pasir Pasang',
        satuan: 'm3',
        koef: 0.5,
        hargaSatuan: 350000,
        totalBiaya: 175000,
        tipe: 'material',
      },
      {
        pekerjaanId: pekerjaan2.id,
        masterHargaId: tukang.id,
        nama: 'Tukang Batu',
        satuan: 'hari',
        koef: 1,
        hargaSatuan: 150000,
        totalBiaya: 150000,
        tipe: 'upah',
      },
    ],
  })

  console.log('✅ Created detail analisa')

  // Create news
  await prisma.news.create({
    data: {
      title: 'Selamat Datang di HitungBangun V3',
      content: 'Versi terbaru dengan fitur lengkap untuk RAB konstruksi',
      isActive: true,
    },
  })

  console.log('✅ Created news')

  console.log('🎉 Seed completed!')
  console.log('\n📝 Test credentials:')
  console.log('Admin: admin@haloraland.id / password123')
  console.log('User: budi@example.com / password123')
  console.log('Demo: demo@haloraland.id / password123')
}

main()
  .catch((e) => {
    console.error('❌ Seed error:', e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
