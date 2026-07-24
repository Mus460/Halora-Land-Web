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

    // Get all pekerjaan grouped by kategori with progress
    const pekerjaan = await prisma.pekerjaan.findMany({
      where: { proyekId },
      orderBy: { kategori: 'asc' }
    })

    // Group by kategori
    const grouped = pekerjaan.reduce((acc, p) => {
      const kategori = p.kategori
      if (!acc[kategori]) {
        acc[kategori] = {
          items: [],
          totalBiaya: 0,
          avgProgress: 0,
        }
      }
      acc[kategori].items.push(p)
      acc[kategori].totalBiaya += Number(p.totalBiaya)
      return acc
    }, {} as Record<string, any>)

    // Calculate progress per kategori (mock for now - should come from separate Progress table)
    Object.keys(grouped).forEach(kategori => {
      const items = grouped[kategori].items
      // Mock progress calculation - in real app, fetch from Progress table
      grouped[kategori].avgProgress = 0
    })

    // Overall project progress
    const totalBiaya = pekerjaan.reduce((sum, p) => sum + Number(p.totalBiaya), 0)
    const overallProgress = 0 // Mock - calculate from actual progress data

    return NextResponse.json({
      proyek: {
        id: proyek.id,
        namaProyek: proyek.namaProyek,
        nilaiKontrak: proyek.nilaiKontrak,
      },
      grouped,
      summary: {
        totalBiaya,
        overallProgress,
        totalKategori: Object.keys(grouped).length,
        totalPekerjaan: pekerjaan.length,
      }
    })
  } catch (error) {
    console.error('Get progress error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
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
    const { pekerjaanId, progress, notes } = body

    if (!pekerjaanId || progress === undefined) {
      return NextResponse.json({ 
        error: 'pekerjaanId and progress are required' 
      }, { status: 400 })
    }

    // Verify pekerjaan belongs to this project
    const pekerjaan = await prisma.pekerjaan.findFirst({
      where: { 
        id: parseInt(pekerjaanId),
        proyekId 
      }
    })

    if (!pekerjaan) {
      return NextResponse.json({ error: 'Pekerjaan tidak ditemukan' }, { status: 404 })
    }

    // TODO: Store progress in separate Progress table
    // For now, return mock success
    return NextResponse.json({ 
      success: true,
      progress: {
        pekerjaanId: parseInt(pekerjaanId),
        progress: parseFloat(progress),
        notes: notes || '',
        updatedAt: new Date(),
      }
    })
  } catch (error) {
    console.error('Update progress error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}
