import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

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
    const masterHargaId = parseInt(id)

    if (isNaN(masterHargaId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

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
    console.error('Get master harga detail error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
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
    const masterHargaId = parseInt(id)

    if (isNaN(masterHargaId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

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

    const body = await request.json()
    const { nama, satuan, harga, kategori } = body

    const masterHarga = await prisma.masterHarga.update({
      where: { id: masterHargaId },
      data: {
        nama: nama !== undefined ? nama : existing.nama,
        satuan: satuan !== undefined ? satuan : existing.satuan,
        harga: harga !== undefined ? parseFloat(harga) : existing.harga,
        kategori: kategori !== undefined ? kategori : existing.kategori,
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
    console.error('Update master harga error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
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
    const masterHargaId = parseInt(id)

    if (isNaN(masterHargaId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

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
    console.error('Delete master harga error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
