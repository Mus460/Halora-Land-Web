import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; logistikId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { logistikId } = await params
    const id = parseInt(logistikId)

    const logistik = await prisma.logistik.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!logistik) {
      return NextResponse.json({ error: 'Logistik tidak ditemukan' }, { status: 404 })
    }

    const hasAccess = 
      session.role === 'ADMIN' ||
      logistik.proyek.userId === session.userId ||
      logistik.proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json({ logistik })
  } catch (error) {
    console.error('Get logistik error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; logistikId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { logistikId } = await params
    const id = parseInt(logistikId)

    const logistik = await prisma.logistik.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!logistik) {
      return NextResponse.json({ error: 'Logistik tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      logistik.proyek.userId === session.userId ||
      logistik.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const { namaMaterial, satuan, volume, hargaSatuan, tanggal } = body

    const newVolume = volume || logistik.volume
    const newHargaSatuan = hargaSatuan || logistik.hargaSatuan
    const totalBiaya = Number(newVolume) * Number(newHargaSatuan)

    const updated = await prisma.logistik.update({
      where: { id },
      data: {
        ...(namaMaterial && { namaMaterial }),
        ...(satuan && { satuan }),
        ...(volume && { volume }),
        ...(hargaSatuan && { hargaSatuan }),
        ...(tanggal && { tanggal: new Date(tanggal) }),
        totalBiaya,
      }
    })

    return NextResponse.json({ logistik: updated })
  } catch (error) {
    console.error('Update logistik error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; logistikId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { logistikId } = await params
    const id = parseInt(logistikId)

    const logistik = await prisma.logistik.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!logistik) {
      return NextResponse.json({ error: 'Logistik tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      logistik.proyek.userId === session.userId ||
      logistik.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    await prisma.logistik.delete({ where: { id } })

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Delete logistik error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}
