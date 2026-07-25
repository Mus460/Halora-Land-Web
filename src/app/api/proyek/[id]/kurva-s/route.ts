import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

/**
 * GET /api/proyek/:id/kurva-s
 * Returns S-curve data (planned vs actual progress)
 */
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

    if (isNaN(proyekId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    // Verify project ownership
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      select: {
        id: true,
        userId: true,
      }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    if (proyek.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // TODO: Calculate real kurva-s from realisasi data
    // For now, return dummy data
    const labels = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun']
    const planned = [10, 25, 45, 65, 85, 100]
    const actual = [8, 22, 40, 60, 78, 92]

    return NextResponse.json({
      labels,
      planned,
      actual,
    })
  } catch (error) {
    console.error('Kurva S error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
