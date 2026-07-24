import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { snapshotAHSP } from '@/lib/snapshot'
import { createAuditLog } from '@/lib/audit'

/**
 * POST /api/pekerjaan/[id]/recalculate
 * Recalculate pekerjaan with latest prices from master
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
    const pekerjaanId = parseInt(idParam)

    // Get pekerjaan with detail
    const pekerjaan = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaanId },
      include: {
        proyek: {
          include: {
            timProyek: true
          }
        },
        detailAnalisa: true
      }
    })

    if (!pekerjaan) {
      return NextResponse.json({ error: 'Pekerjaan not found' }, { status: 404 })
    }

    // Check access (owner/editor only)
    const canEdit =
      session.role === 'ADMIN' ||
      pekerjaan.proyek.userId === session.userId ||
      pekerjaan.proyek.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!canEdit) {
      return NextResponse.json(
        { error: 'Forbidden: only owner/editor can recalculate' },
        { status: 403 }
      )
    }

    // Only for AHSP mode
    if (pekerjaan.metodeHitung !== 'ahsp') {
      return NextResponse.json(
        { error: 'Recalculate only available for AHSP mode' },
        { status: 400 }
      )
    }

    // Get masterAnalisaId from existing detail
    const masterAnalisaId = pekerjaan.detailAnalisa[0]?.masterAnalisaId

    if (!masterAnalisaId) {
      return NextResponse.json(
        { error: 'No master analisa reference found' },
        { status: 400 }
      )
    }

    const oldTotal = Number(pekerjaan.totalBiaya)

    // Re-snapshot with current prices
    const snapshot = await snapshotAHSP(masterAnalisaId, Number(pekerjaan.volume))

    // Delete old snapshot
    await prisma.detailAnalisa.deleteMany({
      where: { pekerjaanId }
    })

    // Create new snapshot
    await prisma.pekerjaan.update({
      where: { id: pekerjaanId },
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
    const percentChange = oldTotal > 0 ? (diff / oldTotal) * 100 : 0

    // ✅ Create audit log
    await createAuditLog({
      proyekId: pekerjaan.proyekId,
      pekerjaanId: pekerjaan.id,
      userId: session.userId,
      action: 'recalculate',
      entityType: 'pekerjaan',
      entityId: pekerjaan.id,
      oldValue: { totalBiaya: oldTotal },
      newValue: { totalBiaya: newTotal },
      description: `Recalculated pekerjaan "${pekerjaan.uraianPekerjaan}" with latest prices`,
    })

    return NextResponse.json({
      success: true,
      message: 'Pekerjaan recalculated with latest prices',
      pekerjaanId,
      oldTotal,
      newTotal,
      diff,
      percentChange
    })
  } catch (error: any) {
    console.error('[pekerjaan/[id]/recalculate POST]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
  }
}
