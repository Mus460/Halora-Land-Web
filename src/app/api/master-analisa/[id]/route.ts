import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError, getJsonBody } from '@/lib/api-utils'
import { createMasterAnalisaSchema } from '@/lib/schemas'

/**
 * GET /api/master-analisa/[id]
 * Get master analisa detail by ID
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
    const id = parseId(idParam, 'id')

    const masterAnalisa = await prisma.masterAnalisa.findUnique({
      where: { id },
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true
          }
        },
        parent: {
          select: {
            id: true,
            kode: true,
            nama: true,
            level: true
          }
        },
        children: {
          select: {
            id: true,
            kode: true,
            nama: true,
            level: true,
            satuan: true
          }
        },
        rincianAnalisa: {
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
        }
      }
    })

    if (!masterAnalisa) {
      return NextResponse.json({ error: 'Master analisa not found' }, { status: 404 })
    }

    // Check access: global or owned by user
    if (!masterAnalisa.isGlobal && masterAnalisa.userId !== session.userId && session.role !== 'ADMIN') {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json(masterAnalisa)
  } catch (error) {
    return handleError(error)
  }
}

/**
 * PUT /api/master-analisa/[id]
 * Update master analisa
 */
export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id: idParam } = await params
    const id = parseId(idParam, 'id')
    
    const body = await getJsonBody(request)
    const validated = createMasterAnalisaSchema.partial().parse(body)

    // Get existing
    const existing = await prisma.masterAnalisa.findUnique({
      where: { id }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Master analisa not found' }, { status: 404 })
    }

    // Check permission
    if (existing.isGlobal && session.role !== 'ADMIN') {
      return NextResponse.json(
        { error: 'Only admin can edit global master analisa' },
        { status: 403 }
      )
    }

    if (!existing.isGlobal && existing.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Check kode uniqueness if changed
    if (validated.kode && validated.kode !== existing.kode) {
      const duplicate = await prisma.masterAnalisa.findFirst({
        where: {
          kode: validated.kode,
          userId: existing.userId,
          id: { not: id }
        }
      })

      if (duplicate) {
        return NextResponse.json(
          { error: 'Kode already exists' },
          { status: 409 }
        )
      }
    }

    const updated = await prisma.masterAnalisa.update({
      where: { id },
      data: {
        kode: validated.kode ?? existing.kode,
        nama: validated.nama ?? existing.nama,
        satuan: validated.satuan !== undefined ? validated.satuan : existing.satuan,
        parentId: validated.parentId !== undefined ? validated.parentId : existing.parentId
      },
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true
          }
        },
        rincianAnalisa: {
          include: {
            komponen: true
          }
        }
      }
    })

    return NextResponse.json(updated)
  } catch (error) {
    return handleError(error)
  }
}

/**
 * DELETE /api/master-analisa/[id]
 * Delete master analisa (cascades to children & rincian)
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
    const id = parseId(idParam, 'id')

    // Get existing
    const existing = await prisma.masterAnalisa.findUnique({
      where: { id },
      include: {
        children: true
      }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Master analisa not found' }, { status: 404 })
    }

    // Check permission
    if (existing.isGlobal && session.role !== 'ADMIN') {
      return NextResponse.json(
        { error: 'Only admin can delete global master analisa' },
        { status: 403 }
      )
    }

    if (!existing.isGlobal && existing.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    // Warn if has children
    if (existing.children.length > 0) {
      return NextResponse.json(
        {
          error: 'Cannot delete: has children',
          childrenCount: existing.children.length
        },
        { status: 409 }
      )
    }

    await prisma.masterAnalisa.delete({
      where: { id }
    })

    return NextResponse.json({ message: 'Master analisa deleted successfully' })
  } catch (error) {
    return handleError(error)
  }
}
