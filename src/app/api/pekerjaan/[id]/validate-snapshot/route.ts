import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { validateSnapshot, compareSnapshot } from '@/lib/snapshot'
import { parseId, handleError } from '@/lib/api-utils'

/**
 * GET /api/pekerjaan/[id]/validate-snapshot
 * Validate if snapshot is stale (harga sudah berubah di master)
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

    const { id: idParam } = await params
    const pekerjaanId = parseId(idParam, 'pekerjaanId')

    // Get pekerjaan with proyek
    const pekerjaan = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaanId },
      include: {
        proyek: {
          include: {
            timProyek: true
          }
        }
      }
    })

    if (!pekerjaan) {
      return NextResponse.json({ error: 'Pekerjaan not found' }, { status: 404 })
    }

    // Check access
    const hasAccess =
      session.role === 'ADMIN' ||
      pekerjaan.proyek.userId === session.userId ||
      pekerjaan.proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Validate snapshot
    const validation = await validateSnapshot(pekerjaanId)
    const comparison = await compareSnapshot(pekerjaanId)

    return NextResponse.json({
      pekerjaanId,
      isValid: validation.isValid,
      changes: validation.changes,
      summary: {
        totalOldCost: comparison.totalOldCost,
        totalNewCost: comparison.totalNewCost,
        totalDiff: comparison.totalDiff,
        percentChange: comparison.percentChange
      },
      message: validation.isValid
        ? 'Snapshot is up-to-date'
        : `${validation.changes.length} komponen harga berubah`
    })
  } catch (error) {
    return handleError(error)
  }
}
