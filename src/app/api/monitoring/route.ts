import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

/**
 * GET /api/monitoring
 * Returns progress monitoring data grouped by kategori
 */
export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { searchParams } = new URL(request.url)
    const proyekId = searchParams.get('proyekId')

    if (!proyekId) {
      return NextResponse.json({ error: 'proyekId required' }, { status: 400 })
    }

    // Verify project ownership
    const proyek = await prisma.proyek.findUnique({
      where: { id: parseInt(proyekId) },
      select: { userId: true }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    if (proyek.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Get all pekerjaan grouped by kategori
    const pekerjaan = await prisma.pekerjaan.findMany({
      where: { proyekId: parseInt(proyekId) },
      orderBy: [
        { kategori: 'asc' },
        { uraianPekerjaan: 'asc' }
      ],
      select: {
        id: true,
        uraianPekerjaan: true,
        kategori: true,
        volume: true,
        satuan: true,
        totalBiaya: true,
        // Calculate progress from realisasi if exists
        // For now, set to 0 (TODO: add progress field or calculate from realisasi)
      }
    })

    // Group by kategori
    const grouped = pekerjaan.reduce((acc, p) => {
      if (!acc[p.kategori]) {
        acc[p.kategori] = []
      }
      acc[p.kategori].push({
        id: p.id,
        nama: p.uraianPekerjaan,
        progress: 0, // TODO: calculate from realisasi or add progress field
        volume: p.volume,
        satuan: p.satuan,
        totalBiaya: p.totalBiaya,
      })
      return acc
    }, {} as Record<string, any[]>)

    // Convert to array format
    const monitoring = Object.entries(grouped).map(([kategori, items]) => ({
      kategori,
      items,
    }))

    return NextResponse.json({ monitoring })
  } catch (error) {
    console.error('Monitoring error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
