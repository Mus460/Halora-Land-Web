import * as XLSX from 'xlsx'

export interface AHSPBreakdownItem {
  tipe: 'upah' | 'material' | 'alat'
  nama: string
  kodeReferensi?: string
  satuan: string
  koefisien: number
  hargaSatuan: number
  jumlahHarga: number
  urutan: number
}

export interface AHSPItem {
  kode: string           // "2.2.1.1.1"
  nama: string
  satuan: string
  hargaSatuan: number
  biayaUmum: number      // 0.10 = 10%
  breakdown: AHSPBreakdownItem[]
}

export interface AHSPParseResult {
  items: AHSPItem[]
  kategori: string
  sheetName: string
  count: number
}

const SHEET_TO_KATEGORI: Record<string, string> = {
  'Persiapan': 'persiapan',
  'Beton': 'beton',
  'Pondasi': 'pondasi',
  'BAJA': 'baja',
  'Rangka Atap': 'atap',
  'Penutup Atap': 'atap',
  'Pasangan Dinding': 'dinding',
  'Plesteran Dan Acian': 'plesteran',
  'Pengecatan dan Pelituran': 'pengecatan',
  'Penutup Lantai dan Dinding': 'keramik',
  'Pintu dan Jendela': 'pintu',
  'Kaca': 'pintu',
  'Sanitair': 'toilet',
  'Jaringan Listrik': 'mep',
  'Perpipaan dalam Gedung': 'mep',
  'Lansekap': 'persiapan',
  'Plafon': 'atap',
  'Besi dan Aluminium': 'baja',
  'Kayu': 'pintu',
}

/**
 * Parse AHSP Excel sheet and extract work items with breakdown
 */
export function parseAHSPSheet(
  sheetName: string,
  workbook: XLSX.WorkBook
): AHSPParseResult {
  const sheet = workbook.Sheets[sheetName]
  if (!sheet) {
    throw new Error(`Sheet "${sheetName}" not found`)
  }

  const data = XLSX.utils.sheet_to_json(sheet, { header: 1, defval: null }) as any[][]
  const items: AHSPItem[] = []
  const kategori = SHEET_TO_KATEGORI[sheetName] || 'custom'

  let i = 0
  while (i < data.length) {
    const row = data[i]
    
    // Check if this is a work item row (column 2 has code with 4+ levels)
    const kode = row[2] ? String(row[2]).trim() : ''
    const nama = row[3] ? String(row[3]).trim() : ''
    const hargaSatuan = row[8] // Price in column 8 (index 8)
    
    // Work item must have kode with 4+ levels (e.g., 2.2.1.1.1)
    if (kode && nama && hargaSatuan && typeof hargaSatuan === 'number') {
      const parts = kode.split('.')
      const validParts = parts.filter(p => p.trim() && /^\d+$/.test(p.trim()))
      
      if (validParts.length >= 4) {
        // This is a valid work item
        // Extract breakdown from next rows
        const breakdown = extractBreakdown(data, i + 1)
        
        // Extract satuan from nama (usually in parentheses or after description)
        const satuan = extractSatuan(row)
        
        items.push({
          kode,
          nama,
          satuan: satuan || 'unit',
          hargaSatuan,
          biayaUmum: 0.10, // Default 10%
          breakdown
        })
        
        // Skip to next item (breakdown rows + separator)
        i += breakdown.length + 10
        continue
      }
    }
    
    i++
  }

  return {
    items,
    kategori,
    sheetName,
    count: items.length
  }
}

/**
 * Extract breakdown (A: Upah, B: Material, C: Alat) from rows
 */
function extractBreakdown(data: any[][], startRow: number): AHSPBreakdownItem[] {
  const breakdown: AHSPBreakdownItem[] = []
  let currentSection: 'upah' | 'material' | 'alat' | null = null
  let urutan = 0
  
  for (let i = startRow; i < Math.min(startRow + 100, data.length); i++) {
    const row = data[i]
    const col2 = row[2] ? String(row[2]).trim() : ''
    
    // Check for section markers
    if (col2 === 'A') {
      currentSection = 'upah'
      continue
    } else if (col2 === 'B') {
      currentSection = 'material'
      continue
    } else if (col2 === 'C') {
      currentSection = 'alat'
      continue
    } else if (col2 === 'D' || col2 === 'E' || col2 === 'F') {
      // End of breakdown
      break
    }
    
    // Extract item if in a section and col2 is a number
    if (currentSection && col2 && /^\d+$/.test(col2)) {
      const nama = row[3] ? String(row[3]).trim() : ''
      const kodeRef = row[4] ? String(row[4]).trim() : undefined
      const satuan = row[4] || row[5] ? String(row[4] || row[5]).trim() : ''
      const koefisien = parseFloat(row[5] || row[6]) || 0
      const hargaSatuan = parseFloat(row[6] || row[7]) || 0
      const jumlahHarga = parseFloat(row[7] || row[9]) || 0
      
      if (nama && koefisien > 0) {
        breakdown.push({
          tipe: currentSection,
          nama,
          kodeReferensi: kodeRef,
          satuan: satuan || 'unit',
          koefisien,
          hargaSatuan,
          jumlahHarga,
          urutan: urutan++
        })
      }
    }
  }
  
  return breakdown
}

/**
 * Extract satuan from row (usually in column 4 or embedded in nama)
 */
function extractSatuan(row: any[]): string | null {
  // Try column 4 first
  if (row[4] && typeof row[4] === 'string') {
    const val = row[4].trim()
    // Common satuan patterns
    if (/^(m|m2|m3|kg|ton|unit|buah|ls|liter|OH|hari|jam)$/i.test(val)) {
      return val
    }
  }
  
  // Try to extract from nama (e.g., "1 m3 beton K-300")
  const nama = row[3] ? String(row[3]) : ''
  const match = nama.match(/\b(m'|m2|m3|kg|ton|unit|buah|ls|liter|OH|hari|jam)\b/i)
  if (match) {
    return match[1]
  }
  
  return null
}

/**
 * Load and parse AHSP workbook from file path
 */
export function loadAHSPWorkbook(filePath: string): XLSX.WorkBook {
  return XLSX.readFile(filePath)
}

/**
 * Get all available detail sheet names from workbook
 */
export function getDetailSheets(workbook: XLSX.WorkBook): string[] {
  // Skip first 2 sheets (main list & master data)
  return workbook.SheetNames.slice(2)
}

/**
 * Get kategori from sheet name
 */
export function getKategoriFromSheet(sheetName: string): string {
  return SHEET_TO_KATEGORI[sheetName] || 'custom'
}
