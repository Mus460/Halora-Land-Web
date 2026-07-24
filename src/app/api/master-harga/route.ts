import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { searchParams } = new URL(request.url)
    const kategori = searchParams.get('kategori') as 'material' | 'upah' | 'alat' | null
    const search = searchParams.get('search')
    const isGlobal = searchParams.get('isGlobal')

    const where: any = {}

    // Filter by kategori
    if (kategori) {
      where.kategori = kategori
    }

    // Filter by isGlobal or user-owned
    if (isGlobal === 'true') {
      where.isGlobal = true
    } else if (isGlobal === 'false') {
      where.AND = [
        { isGlobal: false },
        { userId: session.userId }
      ]
    } else {
      // Show both global and user-owned
      where.OR = [
        { isGlobal: true },
        { userId: session.userId }
      ]
    }

    // Search by nama
    if (search) {
      where.nama = {
        contains: search,
        mode: 'insensitive'
      }
    }

    const masterHarga = await prisma.masterHarga.findMany({
      where,
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true,
          }
        }
      },
      orderBy: [
        { isGlobal: 'desc' },
        { nama: 'asc' }
      ]
    })

    return NextResponse.json({ masterHarga })
  } catch (error) {
    console.error('Get master harga error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await request.json()
    const { nama, satuan, harga, kategori, isGlobal } = body

    if (!nama || !satuan || harga === undefined || !kategori) {
      return NextResponse.json(
        { error: 'Field required: nama, satuan, harga, kategori' },
        { status: 400 }
      )
    }

    // Only admin can create global master harga
    const canCreateGlobal = session.role === 'ADMIN'
    const finalIsGlobal = isGlobal && canCreateGlobal

    const masterHarga = await prisma.masterHarga.create({
      data: {
        nama,
        satuan,
        harga: parseFloat(harga),
        kategori,
        isGlobal: finalIsGlobal,
        userId: finalIsGlobal ? null : session.userId,
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

    return NextResponse.json({ masterHarga }, { status: 201 })
  } catch (error) {
    console.error('Create master harga error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
