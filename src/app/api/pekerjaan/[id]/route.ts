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
    const pekerjaanId = parseInt(id)

    if (isNaN(pekerjaanId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const pekerjaan = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaanId },
      include: {
        proyek: {
          include: {
            timProyek: true,
          }
        },
        detailAnalisa: {
          include: {
            masterHarga: true,
          }
        }
      }
    })

    if (!pekerjaan) {
      return NextResponse.json({ error: 'Pekerjaan tidak ditemukan' }, { status: 404 })
    }

    // Check access
    const hasAccess = 
      session.role === 'ADMIN' ||
      pekerjaan.proyek.userId === session.userId ||
      pekerjaan.proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json({ pekerjaan })
  } catch (error) {
    console.error('Get pekerjaan detail error:', error)
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
    const pekerjaanId = parseInt(id)

    if (isNaN(pekerjaanId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const existing = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaanId },
      include: {
        proyek: {
          include: {
            timProyek: true,
          }
        }
      }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Pekerjaan tidak ditemukan' }, { status: 404 })
    }

    // Check edit permission
    const canEdit = 
      session.role === 'ADMIN' ||
      existing.proyek.userId === session.userId ||
      existing.proyek.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!canEdit) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const {
      kategori,
      uraianPekerjaan,
      volume,
      satuan,
      hargaSatuan,
      metodeHitung,
      levelPekerjaan,
      tipePekerjaan,
    } = body

    const newVolume = volume !== undefined ? parseFloat(volume) : existing.volume
    const newHargaSatuan = hargaSatuan !== undefined ? parseFloat(hargaSatuan) : existing.hargaSatuan
    const totalBiaya = Number(newVolume) * Number(newHargaSatuan)

    const pekerjaan = await prisma.pekerjaan.update({
      where: { id: pekerjaanId },
      data: {
        kategori: kategori !== undefined ? kategori : existing.kategori,
        uraianPekerjaan: uraianPekerjaan !== undefined ? uraianPekerjaan : existing.uraianPekerjaan,
        volume: newVolume,
        satuan: satuan !== undefined ? satuan : existing.satuan,
        hargaSatuan: newHargaSatuan,
        totalBiaya,
        metodeHitung: metodeHitung !== undefined ? metodeHitung : existing.metodeHitung,
        levelPekerjaan: levelPekerjaan !== undefined ? levelPekerjaan : existing.levelPekerjaan,
        tipePekerjaan: tipePekerjaan !== undefined ? tipePekerjaan : existing.tipePekerjaan,
      },
      include: {
        detailAnalisa: {
          include: {
            masterHarga: true,
          }
        }
      }
    })

    return NextResponse.json({ pekerjaan })
  } catch (error) {
    console.error('Update pekerjaan error:', error)
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
    const pekerjaanId = parseInt(id)

    if (isNaN(pekerjaanId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const existing = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaanId },
      include: {
        proyek: {
          include: {
            timProyek: true,
          }
        }
      }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Pekerjaan tidak ditemukan' }, { status: 404 })
    }

    // Check delete permission (owner or admin)
    const canDelete = 
      session.role === 'ADMIN' ||
      existing.proyek.userId === session.userId ||
      existing.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'owner')

    if (!canDelete) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    await prisma.pekerjaan.delete({
      where: { id: pekerjaanId }
    })

    return NextResponse.json({ message: 'Pekerjaan berhasil dihapus' })
  } catch (error) {
    console.error('Delete pekerjaan error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
