import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError, getJsonBody } from '@/lib/api-utils'
import { createLogistikSchema } from '@/lib/schemas'

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
    const id = parseId(logistikId, 'logistikId')

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
    return handleError(error)
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
    const id = parseId(logistikId, 'logistikId')

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

    const body = await getJsonBody(request)
    const validated = createLogistikSchema.partial().parse(body)

    const newVolume = validated.volume ?? logistik.volume
    const newHargaSatuan = validated.hargaSatuan ?? logistik.hargaSatuan
    const totalBiaya = Number(newVolume) * Number(newHargaSatuan)

    const updated = await prisma.logistik.update({
      where: { id },
      data: {
        ...(validated.namaMaterial && { namaMaterial: validated.namaMaterial }),
        ...(validated.satuan && { satuan: validated.satuan }),
        ...(validated.volume !== undefined && { volume: validated.volume }),
        ...(validated.hargaSatuan !== undefined && { hargaSatuan: validated.hargaSatuan }),
        ...(validated.tanggal && { tanggal: new Date(validated.tanggal) }),
        totalBiaya,
      }
    })

    return NextResponse.json({ logistik: updated })
  } catch (error) {
    return handleError(error)
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
    const id = parseId(logistikId, 'logistikId')

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
    return handleError(error)
  }
}
