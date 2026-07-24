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

    if (isNaN(proyekId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const proyek = await prisma.proyek.findUnique({
      where: { id: proyekId },
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
        pekerjaan: {
          orderBy: { createdAt: 'desc' },
          take: 10,
        },
        _count: {
          select: {
            pekerjaan: true,
            rekap: true,
            invoice: true,
          }
        }
      }
    })

    if (!proyek) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    // Check access permission
    const hasAccess = 
      session.role === 'ADMIN' ||
      proyek.userId === session.userId ||
      proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json({ proyek })
  } catch (error) {
    console.error('Get proyek detail error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

export async function PUT(
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

    if (isNaN(proyekId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const existing = await prisma.proyek.findUnique({
      where: { id: proyekId },
      include: {
        timProyek: true,
      }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    // Check edit permission (owner or editor)
    const canEdit = 
      session.role === 'ADMIN' ||
      existing.userId === session.userId ||
      existing.timProyek.some(t => t.userId === session.userId && t.role !== 'viewer')

    if (!canEdit) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const { namaProyek, lokasi, tipe, nilaiKontrak, timeline } = body

    const proyek = await prisma.proyek.update({
      where: { id: proyekId },
      data: {
        namaProyek: namaProyek !== undefined ? namaProyek : existing.namaProyek,
        lokasi: lokasi !== undefined ? lokasi : existing.lokasi,
        tipe: tipe !== undefined ? tipe : existing.tipe,
        nilaiKontrak: nilaiKontrak !== undefined ? (nilaiKontrak ? parseFloat(nilaiKontrak) : null) : existing.nilaiKontrak,
        timeline: timeline !== undefined ? timeline : existing.timeline,
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

    return NextResponse.json({ proyek })
  } catch (error) {
    console.error('Update proyek error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

export async function DELETE(
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

    if (isNaN(proyekId)) {
      return NextResponse.json({ error: 'Invalid ID' }, { status: 400 })
    }

    const existing = await prisma.proyek.findUnique({
      where: { id: proyekId }
    })

    if (!existing) {
      return NextResponse.json({ error: 'Proyek tidak ditemukan' }, { status: 404 })
    }

    // Only owner or admin can delete
    if (session.role !== 'ADMIN' && existing.userId !== session.userId) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    await prisma.proyek.delete({
      where: { id: proyekId }
    })

    return NextResponse.json({ message: 'Proyek berhasil dihapus' })
  } catch (error) {
    console.error('Delete proyek error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
