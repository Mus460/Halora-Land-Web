import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { snapshotAHSP } from '@/lib/snapshot'
import { createAuditLog } from '@/lib/audit'

/**
 * POST /api/proyek/[id]/recalculate-all
 * Bulk recalculate all AHSP pekerjaan in a project
 */
export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id: idParam } = await params
    const proyekId = parseInt(idParam)

    // Check access to project
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      include: {
        timProyek: true
      }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek not found' }, { status: 404 })
    }

    // Check access (owner/editor only)
    const canEdit =
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!canEdit) {
      return NextResponse.json(
        { error: 'Forbidden: only owner/editor can recalculate' },
        { status: 403 }
      )
    }

    // Get all AHSP pekerjaan
    const pekerjaanList = await prisma.pekerjaan.findMany({
      where: {
        proyekId,
        metodeHitung: 'ahsp'
      },
      include: {
        detailAnalisa: true
      }
    })

    if (pekerjaanList.length === 0) {
      return NextResponse.json({
        success: true,
        message: 'No AHSP pekerjaan found',
        recalculatedCount: 0,
        totalOldCost: 0,
        totalNewCost: 0,
        totalDiff: 0,
        details: []
      })
    }

    const results = []
    let totalOldCost = 0
    let totalNewCost = 0
    let skippedCount = 0

    // Recalculate each pekerjaan
    for (const pekerjaan of pekerjaanList) {
      const masterAnalisaId = pekerjaan.detailAnalisa[0]?.masterAnalisaId

      if (!masterAnalisaId) {
        skippedCount++
        continue
      }

      try {
        const oldTotal = Number(pekerjaan.totalBiaya)
        const snapshot = await snapshotAHSP(masterAnalisaId, Number(pekerjaan.volume))

        // Delete old snapshot
        await prisma.detailAnalisa.deleteMany({
          where: { pekerjaanId: pekerjaan.id }
        })

        // Create new snapshot
        await prisma.pekerjaan.update({
          where: { id: pekerjaan.id },
          data: {
            hargaSatuan: snapshot.hargaSatuan,
            totalBiaya: snapshot.totalBiaya,
            detailAnalisa: {
              create: snapshot.components
            }
          }
        })

        const newTotal = snapshot.totalBiaya
        const diff = newTotal - oldTotal

        totalOldCost += oldTotal
        totalNewCost += newTotal

        // ✅ Create audit log for each recalculated pekerjaan
        await createAuditLog({
          proyekId,
          pekerjaanId: pekerjaan.id,
          userId: session.userId,
          action: 'bulk_recalculate',
          entityType: 'pekerjaan',
          entityId: pekerjaan.id,
          oldValue: { totalBiaya: oldTotal },
          newValue: { totalBiaya: newTotal },
          description: `Bulk recalculated pekerjaan "${pekerjaan.uraianPekerjaan}"`,
        })

        results.push({
          id: pekerjaan.id,
          uraian: pekerjaan.uraianPekerjaan,
          oldTotal,
          newTotal,
          diff,
          percentChange: oldTotal > 0 ? (diff / oldTotal) * 100 : 0
        })
      } catch (error: any) {
        console.error(`Failed to recalculate pekerjaan ${pekerjaan.id}:`, error)
        skippedCount++
      }
    }

    const totalDiff = totalNewCost - totalOldCost
    const percentChange = totalOldCost > 0 ? (totalDiff / totalOldCost) * 100 : 0

    return NextResponse.json({
      success: true,
      recalculatedCount: results.length,
      skippedCount,
      totalOldCost,
      totalNewCost,
      totalDiff,
      percentChange,
      details: results
    })
  } catch (error: any) {
    console.error('[proyek/[id]/recalculate-all POST]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
  }
}
