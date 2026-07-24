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
    const proyekId = parseInt(id)

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
    console.error('Get logistik error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
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
    const proyekId = parseInt(id)

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

    const body = await request.json()
    const { namaMaterial, satuan, volume, hargaSatuan, tanggal } = body

    if (!namaMaterial || !satuan || !volume || !hargaSatuan || !tanggal) {
      return NextResponse.json({ 
        error: 'namaMaterial, satuan, volume, hargaSatuan, tanggal are required' 
      }, { status: 400 })
    }

    const totalBiaya = parseFloat(volume) * parseFloat(hargaSatuan)

    const logistik = await prisma.logistik.create({
      data: {
        proyekId,
        namaMaterial,
        satuan,
        volume,
        hargaSatuan,
        totalBiaya,
        tanggal: new Date(tanggal),
      }
    })

    return NextResponse.json({ logistik }, { status: 201 })
  } catch (error) {
    console.error('Create logistik error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}
