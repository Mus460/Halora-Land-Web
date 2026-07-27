import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError, getJsonBody } from '@/lib/api-utils'
import { createMasterHargaSchema } from '@/lib/schemas'

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
    const masterHargaId = parseId(id, 'masterHargaId')

    const masterHarga = await prisma.masterHarga.findUnique({
      where: { id: masterHargaId },
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true,
          }
        }
      }
    })

    if (!masterHarga) {
      return NextResponse.json({ error: 'Master harga tidak ditemukan' }, { status: 404 })
    }

    // Check access: global items are visible to all, user items only to owner
    const hasAccess = 
      masterHarga.isGlobal ||
      masterHarga.userId === session.userId ||
      session.role === 'ADMIN'

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json({ masterHarga })
  } catch (error) {
    return handleError(error)
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id } = await params
    const masterHargaId = parseId(id, 'masterHargaId')

    const existing = await prisma.masterHarga.findUnique({
      where: { id: masterHargaId }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Master harga tidak ditemukan' }, { status: 404 })
    }

    // Check edit permission
    const canEdit = 
      session.role === 'ADMIN' ||
      (!existing.isGlobal && existing.userId === session.userId)

    if (!canEdit) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await getJsonBody(request)
    const validated = createMasterHargaSchema.partial().parse(body)

    const masterHarga = await prisma.masterHarga.update({
      where: { id: masterHargaId },
      data: {
        nama: validated.nama ?? existing.nama,
        satuan: validated.satuan ?? existing.satuan,
        harga: validated.harga ?? existing.harga,
        kategori: validated.kategori ?? existing.kategori,
      },
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true,
          }
        }
      }
    })

    return NextResponse.json({ masterHarga })
  } catch (error) {
    return handleError(error)
  }
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id } = await params
    const masterHargaId = parseId(id, 'masterHargaId')

    const existing = await prisma.masterHarga.findUnique({
      where: { id: masterHargaId }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Master harga tidak ditemukan' }, { status: 404 })
    }

    // Check delete permission
    const canDelete = 
      session.role === 'ADMIN' ||
      (!existing.isGlobal && existing.userId === session.userId)

    if (!canDelete) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    await prisma.masterHarga.delete({
      where: { id: masterHargaId }
    })

    return NextResponse.json({ message: 'Master harga berhasil dihapus' })
  } catch (error) {
    return handleError(error)
  }
}
