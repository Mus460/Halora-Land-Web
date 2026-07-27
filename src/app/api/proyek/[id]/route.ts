import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { parseId, handleError, getJsonBody } from '@/lib/api-utils'
import { createProyekSchema } from '@/lib/schemas'

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
    return handleError(error)
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
    const proyekId = parseId(id, 'proyekId')

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

    const body = await getJsonBody(request)
    const validated = createProyekSchema.partial().parse(body)

    const proyek = await prisma.proyek.update({
      where: { id: proyekId },
      data: {
        namaProyek: validated.namaProyek ?? existing.namaProyek,
        lokasi: validated.lokasi ?? existing.lokasi,
        tipe: (validated.jenisProyek as 'gedung' | 'infra') ?? existing.tipe,
        nilaiKontrak: validated.nilaiKontrak ?? existing.nilaiKontrak,
        timeline: existing.timeline,
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
    return handleError(error)
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
    const proyekId = parseId(id, 'proyekId')

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
    return handleError(error)
  }
}
