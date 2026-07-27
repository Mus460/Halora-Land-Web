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
    const query = searchParams.get('q') || ''
    const kategori = searchParams.get('kategori')
    const limit = parseInt(searchParams.get('limit') || '20')

    if (!query) {
      return NextResponse.json({ results: [] })
    }

    // PostgreSQL full-text search with fuzzy matching
    // Using ILIKE for simplicity (pg_trgm would need extension)
    const whereConditions: any = {
      isSystem: true,
      OR: [
        { nama: { contains: query, mode: 'insensitive' } },
        { ahspKode: { contains: query, mode: 'insensitive' } },
      ]
    }

    if (kategori) {
      whereConditions.kategori = kategori
    }

    const results = await prisma.masterAnalisa.findMany({
      where: whereConditions,
      select: {
        id: true,
        kode: true,
        nama: true,
        satuan: true,
        hargaSatuan: true,
        kategori: true,
        ahspKode: true,
        ahspSheet: true,
        biayaUmum: true,
      },
      orderBy: [
        { kategori: 'asc' },
        { kode: 'asc' }
      ],
      take: limit
    })

    return NextResponse.json({ 
      results,
      total: results.length,
      query
    })

  } catch (error) {
    console.error('Search master analisa error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
