import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

export async function GET(
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

    // Check access
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
      proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const { searchParams } = new URL(request.url)
    const startDate = searchParams.get('startDate')
    const endDate = searchParams.get('endDate')

    const where: any = { proyekId }

    if (startDate || endDate) {
      where.tanggal = {}
      if (startDate) where.tanggal.gte = new Date(startDate)
      if (endDate) where.tanggal.lte = new Date(endDate)
    }

    const realisasi = await prisma.realisasi.findMany({
      where,
      orderBy: { tanggal: 'desc' }
    })

    // Calculate totals
    const totalPengeluaran = realisasi.reduce((sum, item) => sum + Number(item.jumlah), 0)

    // Get RAB total from rekap
    const pekerjaan = await prisma.pekerjaan.findMany({
      where: { proyekId },
      select: { totalBiaya: true }
    })
    const totalRAB = pekerjaan.reduce((sum, p) => sum + Number(p.totalBiaya), 0)

    const selisih = totalRAB - totalPengeluaran

    // Monthly trend
    const monthlyTrend = realisasi.reduce((acc, item) => {
      const month = new Date(item.tanggal).toISOString().slice(0, 7) // YYYY-MM
      if (!acc[month]) acc[month] = 0
      acc[month] += Number(item.jumlah)
      return acc
    }, {} as Record<string, number>)

    return NextResponse.json({ 
      realisasi,
      summary: {
        totalRAB,
        totalPengeluaran,
        selisih,
      },
      monthlyTrend
    })
  } catch (error) {
    console.error('Get realisasi error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

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

    // Check edit access
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      include: { timProyek: true }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const { tanggal, kategori, jumlah, keterangan } = body

    if (!tanggal || !kategori || !jumlah) {
      return NextResponse.json({ 
        error: 'tanggal, kategori, jumlah are required' 
      }, { status: 400 })
    }

    const realisasi = await prisma.realisasi.create({
      data: {
        proyekId,
        tanggal: new Date(tanggal),
        kategori,
        jumlah,
        keterangan: keterangan || '',
      }
    })

    return NextResponse.json({ realisasi }, { status: 201 })
  } catch (error) {
    console.error('Create realisasi error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}
