import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id } = await params
    const proyekId = parseInt(id)
    const body = await request.json()
    const { masterAnalisaId, volume, applyBreakdown = true } = body

    if (!masterAnalisaId || !volume) {
      return NextResponse.json(
        { error: 'masterAnalisaId and volume are required' },
        { status: 400 }
      )
    }

    // Check project access
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      include: { timProyek: true }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    const hasAccess = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Get MasterAnalisa with breakdown
    const masterAnalisa = await prisma.masterAnalisa.findUnique({
      where: { id: masterAnalisaId },
      include: {
        rincianAnalisa: {
          orderBy: { urutan: 'asc' }
        }
      }
    })

    if (!masterAnalisa) {
      return NextResponse.json({ error: 'Master analisa tidak ditemukan' }, { status: 404 })
    }

    // Calculate total harga
    const hargaSatuan = masterAnalisa.hargaSatuan || 0
    const totalBiaya = hargaSatuan * parseFloat(volume.toString())

    // Create Pekerjaan
    const pekerjaan = await prisma.pekerjaan.create({
      data: {
        proyekId,
        kategori: masterAnalisa.kategori as any || 'custom',
        uraianPekerjaan: masterAnalisa.nama,
        volume: parseFloat(volume.toString()),
        satuan: masterAnalisa.satuan || 'unit',
        hargaSatuan: hargaSatuan,
        totalBiaya: totalBiaya,
        metodeHitung: 'ahsp',
        levelPekerjaan: masterAnalisa.level.toString(),
      }
    })

    // Create DetailAnalisa from breakdown if requested
    if (applyBreakdown && masterAnalisa.rincianAnalisa.length > 0) {
      const detailAnalisaData = masterAnalisa.rincianAnalisa.map(rincian => ({
        pekerjaanId: pekerjaan.id,
        masterHargaId: rincian.komponenId,
        masterAnalisaId: masterAnalisa.id,
        nama: rincian.nama || '',
        satuan: rincian.satuan || '',
        koef: rincian.koef,
        hargaSatuan: rincian.hargaSatuan || 0,
        totalBiaya: (rincian.jumlahHarga || 0) * parseFloat(volume.toString()),
        tipe: rincian.tipe,
        snapshotAt: new Date(),
        sourceKode: rincian.kodeReferensi,
      }))

      await prisma.detailAnalisa.createMany({
        data: detailAnalisaData
      })
    }

    // Fetch created pekerjaan with details
    const result = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaan.id },
      include: {
        detailAnalisa: true
      }
    })

    return NextResponse.json({ 
      success: true,
      pekerjaan: result 
    })

  } catch (error) {
    console.error('Create pekerjaan from AHSP error:', error)
    return NextResponse.json(
      { 
        error: 'Terjadi kesalahan server',
        details: error instanceof Error ? error.message : String(error)
      },
      { status: 500 }
    )
  }
}
