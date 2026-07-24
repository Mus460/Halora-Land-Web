import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

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
    const id = parseInt(idParam)

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
  } catch (error: any) {
    console.error('[master-analisa/[id] GET]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
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
    const id = parseInt(idParam)
    const body = await request.json()
    const { kode, nama, satuan, parentId } = body

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
    if (kode && kode !== existing.kode) {
      const duplicate = await prisma.masterAnalisa.findFirst({
        where: {
          kode,
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
        kode: kode || existing.kode,
        nama: nama || existing.nama,
        satuan: satuan !== undefined ? satuan : existing.satuan,
        parentId: parentId !== undefined ? parentId : existing.parentId
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
  } catch (error: any) {
    console.error('[master-analisa/[id] PUT]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
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
    const id = parseInt(idParam)

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
  } catch (error: any) {
    console.error('[master-analisa/[id] DELETE]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
  }
}
