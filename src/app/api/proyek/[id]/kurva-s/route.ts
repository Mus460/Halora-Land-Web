import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError } from '@/lib/api-utils'

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
    const proyekId = parseId(id, 'proyekId')

    // Verify project ownership
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      select: {
        id: true,
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
    return handleError(error)
  }
}
