import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'
import { loadAHSPWorkbook, parseAHSPSheet, getKategoriFromSheet } from '@/lib/ahsp-parser'
import path from 'path'

export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session || session.role !== 'ADMIN') {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await request.json()
    const { kategori, sheetName, forceReimport } = body

    if (!sheetName) {
      return NextResponse.json({ error: 'sheetName is required' }, { status: 400 })
    }

    // Check if already imported
    if (!forceReimport) {
      const existing = await prisma.masterAnalisa.count({
        where: {
          isSystem: true,
          ahspSheet: sheetName
        }
      })

      if (existing > 0) {
        return NextResponse.json({
          message: 'Already imported',
          count: existing,
          skipped: true
        })
      }
    }

    // Load Excel file
    const filePath = path.join(process.cwd(), 'public', 'data', 'ahsp-2026.xlsx')
    const workbook = loadAHSPWorkbook(filePath)

    // Parse sheet
    const result = parseAHSPSheet(sheetName, workbook)

    // Delete existing if reimport
    if (forceReimport) {
      await prisma.masterAnalisa.deleteMany({
        where: {
          isSystem: true,
          ahspSheet: sheetName
        }
      })
    }

    // Batch check existing
    const existingKodes = await prisma.masterAnalisa.findMany({
      where: { ahspKode: { in: result.items.map(i => i.kode) } },
      select: { ahspKode: true }
    })
    const existingSet = new Set(existingKodes.map(e => e.ahspKode))
    const newItems = result.items.filter(i => !existingSet.has(i.kode))

    if (newItems.length === 0) {
      return NextResponse.json({ message: 'Semua sudah imported', skipped: true })
    }

    // Import in transaction (N+1 → ~2 queries)
    const importedCount = await prisma.$transaction(async (tx) => {
      const analisaData = newItems.map(item => ({
        kode: item.kode,
        nama: item.nama,
        level: item.kode.split('.').filter(p => p.trim()).length,
        satuan: item.satuan,
        hargaSatuan: item.hargaSatuan,
        kategori: result.kategori,
        isGlobal: true,
        isSystem: true,
        ahspKode: item.kode,
        ahspSheet: sheetName,
        biayaUmum: item.biayaUmum,
        userId: null,
      }))

      const createdAnalisa = await tx.masterAnalisa.createManyAndReturn({ data: analisaData })
      const analisaMap = new Map(createdAnalisa.map(a => [a.ahspKode, a.id]))

      const rincianData = newItems.flatMap(item => 
        item.breakdown.map(detail => ({
          masterAnalisaId: analisaMap.get(item.kode)!,
          komponenId: null,
          koef: detail.koefisien,
          tipe: detail.tipe,
          nama: detail.nama,
          satuan: detail.satuan,
          hargaSatuan: detail.hargaSatuan,
          jumlahHarga: detail.jumlahHarga,
          kodeReferensi: detail.kodeReferensi,
          urutan: detail.urutan
        }))
      )

      await tx.rincianAnalisa.createMany({ data: rincianData })
      return createdAnalisa.length
    })

    return NextResponse.json({
      success: true,
      sheetName,
      kategori: result.kategori,
      imported: importedCount,
      total: result.count
    })

  } catch (error) {
    console.error('Import AHSP error:', error)
    return NextResponse.json(
      { 
        error: 'Terjadi kesalahan import', 
        details: error instanceof Error ? error.message : String(error)
      },
      { status: 500 }
    )
  }
}

export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session || session.role !== 'ADMIN') {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Get import status for all sheets
    const filePath = path.join(process.cwd(), 'public', 'data', 'ahsp-2026.xlsx')
    const workbook = loadAHSPWorkbook(filePath)
    const sheets = workbook.SheetNames.slice(2) // Skip first 2

    const status = await Promise.all(
      sheets.map(async (sheetName) => {
        const count = await prisma.masterAnalisa.count({
          where: {
            isSystem: true,
            ahspSheet: sheetName
          }
        })

        return {
          sheetName,
          kategori: getKategoriFromSheet(sheetName),
          imported: count > 0,
          count
        }
      })
    )

    return NextResponse.json({ status })

  } catch (error) {
    console.error('Get import status error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
