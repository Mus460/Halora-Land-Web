import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; realisasiId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { realisasiId } = await params
    const id = parseInt(realisasiId)

    const realisasi = await prisma.realisasi.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!realisasi) {
      return NextResponse.json({ error: 'Realisasi tidak ditemukan' }, { status: 404 })
    }

    const hasAccess = 
      session.role === 'ADMIN' ||
      realisasi.proyek.userId === session.userId ||
      realisasi.proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json({ realisasi })
  } catch (error) {
    console.error('Get realisasi error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; realisasiId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { realisasiId } = await params
    const id = parseInt(realisasiId)

    const realisasi = await prisma.realisasi.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!realisasi) {
      return NextResponse.json({ error: 'Realisasi tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      realisasi.proyek.userId === session.userId ||
      realisasi.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const { tanggal, kategori, jumlah, keterangan } = body

    const updated = await prisma.realisasi.update({
      where: { id },
      data: {
        ...(tanggal && { tanggal: new Date(tanggal) }),
        ...(kategori && { kategori }),
        ...(jumlah && { jumlah }),
        ...(keterangan !== undefined && { keterangan }),
      }
    })

    return NextResponse.json({ realisasi: updated })
  } catch (error) {
    console.error('Update realisasi error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; realisasiId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { realisasiId } = await params
    const id = parseInt(realisasiId)

    const realisasi = await prisma.realisasi.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!realisasi) {
      return NextResponse.json({ error: 'Realisasi tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      realisasi.proyek.userId === session.userId ||
      realisasi.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    await prisma.realisasi.delete({ where: { id } })

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Delete realisasi error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}
