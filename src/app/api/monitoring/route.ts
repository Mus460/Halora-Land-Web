import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError } from '@/lib/api-utils'

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
    const proyekIdRaw = searchParams.get('proyekId')

    if (!proyekIdRaw) {
      return NextResponse.json({ error: 'proyekId required' }, { status: 400 })
    }

    const proyekId = parseId(proyekIdRaw, 'proyekId')

    // Verify project ownership or team access
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      select: { 
        userId: true,
        timProyek: {
          select: { userId: true }
        }
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
    return handleError(error)
  }
}
