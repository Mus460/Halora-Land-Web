import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError } from '@/lib/api-utils'

/**
 * GET /api/master-analisa/[id]/rincian
 * Get rincian analisa (komponen breakdown) for a master analisa
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
    const id = parseId(idParam, 'master-analisa id')

    // Verify master analisa exists
    const masterAnalisa = await prisma.masterAnalisa.findUnique({
      where: { id },
      select: { id: true, isGlobal: true, userId: true }
    })

    if (!masterAnalisa) {
      return NextResponse.json({ error: 'Master analisa not found' }, { status: 404 })
    }

    // Check access
    if (!masterAnalisa.isGlobal && masterAnalisa.userId !== session.userId && session.role !== 'ADMIN') {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const rincianAnalisa = await prisma.rincianAnalisa.findMany({
      where: { masterAnalisaId: id },
      include: {
        komponen: {
          select: {
            id: true,
            nama: true,
            satuan: true,
            harga: true,
            kategori: true
          }
        }
      },
      orderBy: { tipe: 'asc' }
    })

    // Calculate total harga satuan
    const totalHargaSatuan = rincianAnalisa.reduce(
      (sum, r) => sum + (Number(r.koef) * Number(r.komponen.harga)),
      0
    )

    return NextResponse.json({
      masterAnalisaId: id,
      total: rincianAnalisa.length,
      totalHargaSatuan,
      data: rincianAnalisa
    })
  } catch (error) {
    return handleError(error)
  }
}

/**
 * POST /api/master-analisa/[id]/rincian
 * Add komponen to rincian analisa
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
    const id = parseId(idParam, 'master-analisa id')
    const body = await request.json()
    const { komponenId, koef, tipe } = body

    // Validation
    if (!komponenId || !koef || !tipe) {
      return NextResponse.json(
        { error: 'Missing required fields: komponenId, koef, tipe' },
        { status: 400 }
      )
    }

    // Verify master analisa exists & check permission
    const masterAnalisa = await prisma.masterAnalisa.findUnique({
      where: { id }
    })

    if (!masterAnalisa) {
      return NextResponse.json({ error: 'Master analisa not found' }, { status: 404 })
    }

    if (masterAnalisa.isGlobal && session.role !== 'ADMIN') {
      return NextResponse.json(
        { error: 'Only admin can edit global master analisa' },
        { status: 403 }
      )
    }

    if (!masterAnalisa.isGlobal && masterAnalisa.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Verify komponen exists
    const komponen = await prisma.masterHarga.findUnique({
      where: { id: komponenId }
    })

    if (!komponen) {
      return NextResponse.json({ error: 'Komponen not found' }, { status: 404 })
    }

    // Check duplicate
    const existing = await prisma.rincianAnalisa.findFirst({
      where: {
        masterAnalisaId: id,
        komponenId
      }
    })

    if (existing) {
      return NextResponse.json(
        { error: 'Komponen already exists in this analisa' },
        { status: 409 }
      )
    }

    const rincian = await prisma.rincianAnalisa.create({
      data: {
        masterAnalisaId: id,
        komponenId,
        koef,
        tipe
      },
      include: {
        komponen: {
          select: {
            id: true,
            nama: true,
            satuan: true,
            harga: true,
            kategori: true
          }
        }
      }
    })

    return NextResponse.json(rincian, { status: 201 })
  } catch (error) {
    return handleError(error)
  }
}

/**
 * DELETE /api/master-analisa/[id]/rincian
 * Remove komponen from rincian analisa
 * Body: { rincianId: number }
 */
export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id: idParam } = await params
    const id = parseId(idParam, 'master-analisa id')
    const body = await request.json()
    const { rincianId } = body

    if (!rincianId) {
      return NextResponse.json(
        { error: 'Missing required field: rincianId' },
        { status: 400 }
      )
    }

    // Verify master analisa exists & check permission
    const masterAnalisa = await prisma.masterAnalisa.findUnique({
      where: { id }
    })

    if (!masterAnalisa) {
      return NextResponse.json({ error: 'Master analisa not found' }, { status: 404 })
    }

    if (masterAnalisa.isGlobal && session.role !== 'ADMIN') {
      return NextResponse.json(
        { error: 'Only admin can edit global master analisa' },
        { status: 403 }
      )
    }

    if (!masterAnalisa.isGlobal && masterAnalisa.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Verify rincian belongs to this master analisa
    const rincian = await prisma.rincianAnalisa.findUnique({
      where: { id: rincianId }
    })

    if (!rincian || rincian.masterAnalisaId !== id) {
      return NextResponse.json({ error: 'Rincian not found' }, { status: 404 })
    }

    await prisma.rincianAnalisa.delete({
      where: { id: rincianId }
    })

    return NextResponse.json({ message: 'Rincian deleted successfully' })
  } catch (error) {
    return handleError(error)
  }
}
