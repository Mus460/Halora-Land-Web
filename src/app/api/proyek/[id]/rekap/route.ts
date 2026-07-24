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

    // Get all pekerjaan grouped by kategori
    const pekerjaan = await prisma.pekerjaan.findMany({
      where: { proyekId },
      include: {
        detailAnalisa: true,
      },
      orderBy: { kategori: 'asc' }
    })

    // Group by kategori
    const grouped = pekerjaan.reduce((acc, p) => {
      const kategori = p.kategori
      if (!acc[kategori]) {
        acc[kategori] = []
      }
      acc[kategori].push(p)
      return acc
    }, {} as Record<string, typeof pekerjaan>)

    // Calculate totals
    const subtotals: Record<string, number> = {}
    let grandTotal = 0

    Object.keys(grouped).forEach(kategori => {
      const items = grouped[kategori]
      const subtotal = items.reduce((sum, item) => sum + Number(item.totalBiaya), 0)
      subtotals[kategori] = subtotal
      grandTotal += subtotal
    })

    // Get margin settings from rekap table (if exists)
    const rekapSettings = await prisma.rekap.findFirst({
      where: { 
        proyekId,
        kategori: 'settings'
      }
    })

    const margin = rekapSettings?.margin ? Number(rekapSettings.margin) : 0
    const overhead = 0 // TODO: make configurable
    const profit = 0 // TODO: make configurable
    const ppn = 0.11 // 11% PPN

    const subtotalWithMargin = grandTotal * (1 + margin / 100)
    const subtotalBeforeTax = subtotalWithMargin + overhead + profit
    const totalPPN = subtotalBeforeTax * ppn
    const totalAkhir = subtotalBeforeTax + totalPPN

    return NextResponse.json({
      proyek: {
        id: proyek.id,
        namaProyek: proyek.namaProyek,
        lokasi: proyek.lokasi,
        nilaiKontrak: proyek.nilaiKontrak,
      },
      grouped,
      subtotals,
      summary: {
        subtotal: grandTotal,
        margin: margin,
        subtotalWithMargin,
        overhead,
        profit,
        subtotalBeforeTax,
        ppn: ppn * 100, // return as percentage
        totalPPN,
        totalAkhir,
      }
    })
  } catch (error) {
    console.error('Get rekap error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

export async function PUT(
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
      include: {
        timProyek: true,
      }
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
    const { margin } = body

    if (margin === undefined) {
      return NextResponse.json({ error: 'margin is required' }, { status: 400 })
    }

    // Find existing settings
    const existing = await prisma.rekap.findFirst({
      where: { 
        proyekId,
        kategori: 'settings'
      }
    })

    let rekap
    if (existing) {
      rekap = await prisma.rekap.update({
        where: { id: existing.id },
        data: { margin }
      })
    } else {
      rekap = await prisma.rekap.create({
        data: {
          proyekId,
          kategori: 'settings',
          uraian: 'Margin & Overhead Settings',
          urutan: 0,
          margin,
        }
      })
    }

    return NextResponse.json({ 
      success: true,
      rekap 
    })
  } catch (error) {
    console.error('Update rekap settings error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
