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
    const tipe = searchParams.get('tipe') as 'gedung' | 'infra' | null
    const search = searchParams.get('search')

    const where: any = {}

    // Filter by user (non-admin only see their own projects + shared projects)
    if (session.role !== 'ADMIN') {
      where.OR = [
        { userId: session.userId },
        { timProyek: { some: { userId: session.userId } } }
      ]
    }

    // Filter by tipe
    if (tipe) {
      where.tipe = tipe
    }

    // Search by name or location
    if (search) {
      where.AND = [
        where.AND || {},
        {
          OR: [
            { namaProyek: { contains: search, mode: 'insensitive' } },
            { lokasi: { contains: search, mode: 'insensitive' } }
          ]
        }
      ]
    }

    const proyek = await prisma.proyek.findMany({
      where,
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true,
            role: true,
          }
        },
        timProyek: {
          include: {
            user: {
              select: {
                id: true,
                namaLengkap: true,
                email: true,
              }
            }
          }
        },
        _count: {
          select: {
            pekerjaan: true,
          }
        }
      },
      orderBy: { createdAt: 'desc' }
    })

    return NextResponse.json({ proyek })
  } catch (error) {
    console.error('Get proyek error:', error)
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
    const { namaProyek, lokasi, tipe, nilaiKontrak, timeline } = body

    if (!namaProyek) {
      return NextResponse.json(
        { error: 'Nama proyek harus diisi' },
        { status: 400 }
      )
    }

    const proyek = await prisma.proyek.create({
      data: {
        userId: session.userId,
        namaProyek,
        lokasi: lokasi || null,
        tipe: tipe || 'gedung',
        nilaiKontrak: nilaiKontrak ? parseFloat(nilaiKontrak) : null,
        timeline: timeline || null,
      },
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true,
            role: true,
          }
        },
        _count: {
          select: {
            pekerjaan: true,
          }
        }
      }
    })

    return NextResponse.json({ proyek }, { status: 201 })
  } catch (error) {
    console.error('Create proyek error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
