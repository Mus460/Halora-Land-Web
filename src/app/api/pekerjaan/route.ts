import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { snapshotAHSP } from '@/lib/snapshot'

export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { searchParams } = new URL(request.url)
    const proyekId = searchParams.get('proyekId')
    const kategori = searchParams.get('kategori')
    const search = searchParams.get('search')

    if (!proyekId) {
      return NextResponse.json(
        { error: 'proyekId is required' },
        { status: 400 }
      )
    }

    // Check access to project
    const proyek = await prisma.proyek.findUnique({
      where: { id: parseInt(proyekId) },
      include: {
        timProyek: true,
      }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    const hasAccess = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const where: any = {
      proyekId: parseInt(proyekId)
    }

    if (kategori) {
      where.kategori = kategori
    }

    if (search) {
      where.uraianPekerjaan = {
        contains: search,
        mode: 'insensitive'
      }
    }

    const pekerjaan = await prisma.pekerjaan.findMany({
      where,
      include: {
        detailAnalisa: {
          include: {
            masterHarga: true,
          }
        }
      },
      orderBy: { createdAt: 'desc' }
    })

    return NextResponse.json({ pekerjaan })
  } catch (error) {
    console.error('Get pekerjaan error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await request.json()
    const {
      proyekId,
      kategori,
      uraianPekerjaan,
      volume,
      satuan,
      metodeHitung,
      levelPekerjaan,
      tipePekerjaan,
      
      // For AHSP mode
      masterAnalisaId,
      
      // For manual mode
      hargaSatuan: manualHargaSatuan,
      detailAnalisa: manualDetailAnalisa
    } = body

    if (!proyekId || !kategori || !uraianPekerjaan || !volume || !satuan) {
      return NextResponse.json(
        { error: 'Field required: proyekId, kategori, uraianPekerjaan, volume, satuan' },
        { status: 400 }
      )
    }

    // Check access to project
    const proyek = await prisma.proyek.findUnique({
      where: { id: parseInt(proyekId) },
      include: {
        timProyek: true,
      }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    const canEdit = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!canEdit) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    let hargaSatuan: number
    let totalBiaya: number
    let detailAnalisaData: any[] = []

    // Handle calculation based on metode hitung
    if (metodeHitung === 'ahsp' && masterAnalisaId) {
      // ✅ AHSP Mode: Create snapshot
      const snapshot = await snapshotAHSP(masterAnalisaId, parseFloat(volume))
      hargaSatuan = snapshot.hargaSatuan
      totalBiaya = snapshot.totalBiaya
      detailAnalisaData = snapshot.components
      
    } else if (metodeHitung === 'manual') {
      // Manual mode: No snapshot needed
      hargaSatuan = parseFloat(manualHargaSatuan || 0)
      totalBiaya = parseFloat(volume) * hargaSatuan
      
      if (manualDetailAnalisa && Array.isArray(manualDetailAnalisa)) {
        detailAnalisaData = manualDetailAnalisa.map((item: any) => ({
          masterHargaId: item.masterHargaId || null,
          masterAnalisaId: null,
          nama: item.nama,
          satuan: item.satuan,
          koef: parseFloat(item.koef),
          hargaSatuan: parseFloat(item.hargaSatuan),
          totalBiaya: parseFloat(item.koef) * parseFloat(item.hargaSatuan),
          tipe: item.tipe,
          snapshotAt: new Date(),
          sourceKode: null
        }))
      }
    } else {
      return NextResponse.json(
        { error: 'Invalid metode hitung or missing masterAnalisaId for AHSP mode' },
        { status: 400 }
      )
    }

    const pekerjaan = await prisma.pekerjaan.create({
      data: {
        proyekId: parseInt(proyekId),
        kategori,
        uraianPekerjaan,
        volume: parseFloat(volume),
        satuan,
        hargaSatuan,
        totalBiaya,
        metodeHitung: metodeHitung || 'manual',
        levelPekerjaan: levelPekerjaan || null,
        tipePekerjaan: tipePekerjaan || null,
        detailAnalisa: detailAnalisaData.length > 0 ? {
          create: detailAnalisaData
        } : undefined
      },
      include: {
        detailAnalisa: {
          include: {
            masterHarga: true,
            masterAnalisa: true
          }
        }
      }
    })

    return NextResponse.json({ pekerjaan }, { status: 201 })
  } catch (error: any) {
    console.error('Create pekerjaan error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server', message: error.message },
      { status: 500 }
    )
  }
}
