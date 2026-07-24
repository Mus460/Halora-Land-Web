import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

/**
 * GET /api/master-analisa
 * List master analisa with tree structure
 * Query params:
 *   - level: filter by level (0=root, 1=divisi, 2=kelompok, 3=sub, 4=item)
 *   - parentId: filter by parent
 *   - search: search by kode or nama
 *   - isGlobal: filter global vs user-owned
 */
export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { searchParams } = new URL(request.url)
    const level = searchParams.get('level')
    const parentId = searchParams.get('parentId')
    const search = searchParams.get('search')
    const isGlobal = searchParams.get('isGlobal')

    const where: any = {}

    // Filter by level
    if (level !== null) {
      where.level = parseInt(level)
    }

    // Filter by parent
    if (parentId !== null) {
      where.parentId = parentId === 'null' ? null : parseInt(parentId)
    }

    // Search by kode or nama
    if (search) {
      where.OR = [
        { kode: { contains: search, mode: 'insensitive' } },
        { nama: { contains: search, mode: 'insensitive' } }
      ]
    }

    // Filter global vs user-owned
    if (isGlobal !== null) {
      if (isGlobal === 'true') {
        where.isGlobal = true
      } else {
        where.OR = [
          { isGlobal: false, userId: session.userId },
          { isGlobal: true }
        ]
      }
    } else {
      // Default: show global + user-owned
      where.OR = [
        { isGlobal: true },
        { userId: session.userId }
      ]
    }

    const masterAnalisa = await prisma.masterAnalisa.findMany({
      where,
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true
          }
        },
        children: {
          select: {
            id: true,
            kode: true,
            nama: true,
            level: true
          }
        },
        rincianAnalisa: {
          include: {
            komponen: {
              select: {
                id: true,
                nama: true,
                satuan: true,
                harga: true,
                kategori: true
              }
            }
          }
        }
      },
      orderBy: [
        { level: 'asc' },
        { kode: 'asc' }
      ]
    })

    return NextResponse.json({
      total: masterAnalisa.length,
      data: masterAnalisa
    })
  } catch (error: any) {
    console.error('[master-analisa GET]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
  }
}

/**
 * POST /api/master-analisa
 * Create new master analisa node
 */
export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await request.json()
    const { kode, nama, level, parentId, satuan, isGlobal } = body

    // Validation
    if (!kode || !nama || level === undefined) {
      return NextResponse.json(
        { error: 'Missing required fields: kode, nama, level' },
        { status: 400 }
      )
    }

    // Only ADMIN can create global master analisa
    if (isGlobal && session.role !== 'ADMIN') {
      return NextResponse.json(
        { error: 'Only admin can create global master analisa' },
        { status: 403 }
      )
    }

    // Check if kode already exists
    const existing = await prisma.masterAnalisa.findFirst({
      where: {
        kode,
        userId: isGlobal ? null : session.userId
      }
    })

    if (existing) {
      return NextResponse.json(
        { error: 'Kode already exists' },
        { status: 409 }
      )
    }

    // Verify parent exists if parentId provided
    if (parentId) {
      const parent = await prisma.masterAnalisa.findUnique({
        where: { id: parentId }
      })
      if (!parent) {
        return NextResponse.json(
          { error: 'Parent not found' },
          { status: 404 }
        )
      }
    }

    const masterAnalisa = await prisma.masterAnalisa.create({
      data: {
        kode,
        nama,
        level,
        parentId: parentId || null,
        satuan: satuan || null,
        isGlobal: isGlobal || false,
        userId: isGlobal ? null : session.userId
      },
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true
          }
        }
      }
    })

    return NextResponse.json(masterAnalisa, { status: 201 })
  } catch (error: any) {
    console.error('[master-analisa POST]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
  }
}
