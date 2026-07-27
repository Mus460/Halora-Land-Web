import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { Prisma } from '@prisma/client'
import { handleError, getJsonBody } from '@/lib/api-utils'
import { createProyekSchema } from '@/lib/schemas'

export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { searchParams } = new URL(request.url)
    const tipe = searchParams.get('tipe') as 'gedung' | 'infra' | null
    const search = searchParams.get('search')

    const where: Prisma.ProyekWhereInput = {}

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
    return handleError(error)
  }
}

export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await getJsonBody(request)
    const validated = createProyekSchema.parse(body)

    const proyek = await prisma.proyek.create({
      data: {
        userId: session.userId,
        namaProyek: validated.namaProyek,
        lokasi: validated.lokasi || null,
        tipe: (validated.jenisProyek as 'gedung' | 'infra') || 'gedung',
        nilaiKontrak: validated.nilaiKontrak || null,
        timeline: null,
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
    return handleError(error)
  }
}
