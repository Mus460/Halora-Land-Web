import { prisma } from '@/lib/prisma'
import { Prisma, TipeKomponen } from '@prisma/client'

type Decimal = Prisma.Decimal

/**
 * Snapshot result from AHSP calculation
 */
export interface SnapshotResult {
  hargaSatuan: number
  totalBiaya: number
  components: SnapshotComponent[]
  metadata: {
    analisaNama: string
    analisaKode: string
    satuan: string | null
  }
}

/**
 * Individual component in snapshot
 */
export interface SnapshotComponent {
  masterHargaId: number
  masterAnalisaId: number
  nama: string
  satuan: string
  koef: Decimal
  hargaSatuan: Decimal
  totalBiaya: number
  tipe: TipeKomponen
  snapshotAt: Date
  sourceKode: string
}

/**
 * Validation result for snapshot
 */
export interface ValidationResult {
  isValid: boolean
  changes: PriceChange[]
}

/**
 * Price change detail
 */
export interface PriceChange {
  nama: string
  oldHarga: number
  newHarga: number
  diff: number
  percentChange: number
}

/**
 * Create snapshot of AHSP components for a pekerjaan
 * 
 * @param masterAnalisaId - ID of master analisa to snapshot
 * @param volume - Volume of work
 * @returns Snapshot result with harga satuan, total biaya, and components
 */
export async function snapshotAHSP(
  masterAnalisaId: number,
  volume: number
): Promise<SnapshotResult> {
  // 1. Get master analisa with rincian
  const masterAnalisa = await prisma.masterAnalisa.findUnique({
    where: { id: masterAnalisaId },
    include: {
      rincianAnalisa: {
        include: {
          komponen: true
        }
      }
    }
  })

  if (!masterAnalisa) {
    throw new Error(`Master Analisa with ID ${masterAnalisaId} not found`)
  }

  if (!masterAnalisa.rincianAnalisa || masterAnalisa.rincianAnalisa.length === 0) {
    throw new Error(`Master Analisa ${masterAnalisa.kode} has no rincian components`)
  }

  // 2. Build snapshot components (freeze current prices)
  const snapshotComponents: SnapshotComponent[] = masterAnalisa.rincianAnalisa.map(rincian => {
    const koefNum = Number(rincian.koef)
    const hargaNum = Number(rincian.komponen.harga)
    const totalBiaya = koefNum * hargaNum

    return {
      masterHargaId: rincian.komponenId,
      masterAnalisaId: masterAnalisaId,
      nama: rincian.komponen.nama,
      satuan: rincian.komponen.satuan,
      koef: rincian.koef,
      hargaSatuan: rincian.komponen.harga,
      totalBiaya,
      tipe: rincian.tipe,
      snapshotAt: new Date(),
      sourceKode: masterAnalisa.kode,
    }
  })

  // 3. Calculate totals
  const hargaSatuan = snapshotComponents.reduce((sum, c) => sum + c.totalBiaya, 0)
  const totalBiaya = hargaSatuan * volume

  return {
    hargaSatuan,
    totalBiaya,
    components: snapshotComponents,
    metadata: {
      analisaNama: masterAnalisa.nama,
      analisaKode: masterAnalisa.kode,
      satuan: masterAnalisa.satuan,
    }
  }
}

/**
 * Validate if snapshot is stale (harga sudah berubah di master)
 * 
 * @param pekerjaanId - ID of pekerjaan to validate
 * @returns Validation result with changes detail
 */
export async function validateSnapshot(pekerjaanId: number): Promise<ValidationResult> {
  const pekerjaan = await prisma.pekerjaan.findUnique({
    where: { id: pekerjaanId },
    include: {
      detailAnalisa: true
    }
  })

  if (!pekerjaan || pekerjaan.detailAnalisa.length === 0) {
    return { isValid: true, changes: [] }
  }

  const changes: PriceChange[] = []

  for (const detail of pekerjaan.detailAnalisa) {
    if (!detail.masterHargaId) continue

    const currentHarga = await prisma.masterHarga.findUnique({
      where: { id: detail.masterHargaId },
      select: { harga: true }
    })

    if (!currentHarga) continue

    const oldHarga = Number(detail.hargaSatuan)
    const newHarga = Number(currentHarga.harga)

    if (oldHarga !== newHarga) {
      const diff = newHarga - oldHarga
      const percentChange = oldHarga > 0 ? (diff / oldHarga) * 100 : 0

      changes.push({
        nama: detail.nama,
        oldHarga,
        newHarga,
        diff,
        percentChange
      })
    }
  }

  return {
    isValid: changes.length === 0,
    changes
  }
}

/**
 * Compare snapshot with current master prices (for summary)
 * 
 * @param pekerjaanId - ID of pekerjaan to compare
 * @returns Summary of old vs new costs
 */
export async function compareSnapshot(pekerjaanId: number): Promise<{
  totalOldCost: number
  totalNewCost: number
  totalDiff: number
  percentChange: number
}> {
  const validation = await validateSnapshot(pekerjaanId)
  
  if (validation.isValid) {
    const pekerjaan = await prisma.pekerjaan.findUnique({
      where: { id: pekerjaanId },
      select: { totalBiaya: true, volume: true }
    })
    
    const totalOldCost = pekerjaan ? Number(pekerjaan.totalBiaya) : 0
    
    return {
      totalOldCost,
      totalNewCost: totalOldCost,
      totalDiff: 0,
      percentChange: 0
    }
  }

  // Recalculate with new prices
  const pekerjaan = await prisma.pekerjaan.findUnique({
    where: { id: pekerjaanId },
    include: {
      detailAnalisa: true
    }
  })

  if (!pekerjaan) {
    throw new Error(`Pekerjaan with ID ${pekerjaanId} not found`)
  }

  const totalOldCost = Number(pekerjaan.totalBiaya)
  
  let newHargaSatuan = 0
  for (const detail of pekerjaan.detailAnalisa) {
    if (!detail.masterHargaId) {
      // Use snapshot if no reference
      newHargaSatuan += Number(detail.totalBiaya)
      continue
    }

    const currentHarga = await prisma.masterHarga.findUnique({
      where: { id: detail.masterHargaId },
      select: { harga: true }
    })

    if (currentHarga) {
      newHargaSatuan += Number(detail.koef) * Number(currentHarga.harga)
    } else {
      // Use snapshot if master deleted
      newHargaSatuan += Number(detail.totalBiaya)
    }
  }

  const totalNewCost = newHargaSatuan * Number(pekerjaan.volume)
  const totalDiff = totalNewCost - totalOldCost
  const percentChange = totalOldCost > 0 ? (totalDiff / totalOldCost) * 100 : 0

  return {
    totalOldCost,
    totalNewCost,
    totalDiff,
    percentChange
  }
}
