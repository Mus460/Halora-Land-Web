import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

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
        }
      }
    })

    if (!pekerjaan) {
      return NextResponse.json({ error: 'Pekerjaan tidak ditemukan' }, { status: 404 })
    }

    // Check edit permission
    const canEdit = 
      session.role === 'ADMIN' ||
      pekerjaan.proyek.userId === session.userId ||
      pekerjaan.proyek.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!canEdit) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const { masterHargaId, nama, satuan, koef, hargaSatuan, tipe } = body

    if (!nama || !satuan || koef === undefined || hargaSatuan === undefined || !tipe) {
      return NextResponse.json(
        { error: 'Field required: nama, satuan, koef, hargaSatuan, tipe' },
        { status: 400 }
      )
    }

    const totalBiaya = parseFloat(koef) * parseFloat(hargaSatuan)

    const detailAnalisa = await prisma.detailAnalisa.create({
      data: {
        pekerjaanId,
        masterHargaId: masterHargaId || null,
        nama,
        satuan,
        koef: parseFloat(koef),
        hargaSatuan: parseFloat(hargaSatuan),
        totalBiaya,
        tipe,
      },
      include: {
        masterHarga: true,
      }
    })

    return NextResponse.json({ detailAnalisa }, { status: 201 })
  } catch (error) {
    console.error('Add detail analisa error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

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

    const detailAnalisa = await prisma.detailAnalisa.findMany({
      where: { pekerjaanId },
      include: {
        masterHarga: true,
      },
      orderBy: { id: 'asc' }
    })

    return NextResponse.json({ detailAnalisa })
  } catch (error) {
    console.error('Get detail analisa error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
