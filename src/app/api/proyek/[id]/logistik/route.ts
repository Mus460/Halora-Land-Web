import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError, getJsonBody } from '@/lib/api-utils'
import { createLogistikSchema } from '@/lib/schemas'

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
    const proyekId = parseId(id, 'proyekId')

    // Check access
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      include: { timProyek: true }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    const hasAccess = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const logistik = await prisma.logistik.findMany({
      where: { proyekId },
      orderBy: { tanggal: 'desc' }
    })

    const totalBiaya = logistik.reduce((sum, item) => sum + Number(item.totalBiaya), 0)

    return NextResponse.json({ logistik, totalBiaya })
  } catch (error) {
    return handleError(error)
  }
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { id } = await params
    const proyekId = parseId(id, 'proyekId')

    // Check edit access
    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
      include: { timProyek: true }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await getJsonBody(request)
    const validated = createLogistikSchema.parse(body)

    const totalBiaya = validated.volume * validated.hargaSatuan

    const logistik = await prisma.logistik.create({
      data: {
        proyekId,
        namaMaterial: validated.namaMaterial,
        satuan: validated.satuan,
        volume: validated.volume,
        hargaSatuan: validated.hargaSatuan,
        totalBiaya,
        tanggal: new Date(validated.tanggal),
      }
    })

    return NextResponse.json({ logistik }, { status: 201 })
  } catch (error) {
    return handleError(error)
  }
}
