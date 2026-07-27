/**
 * Zod-vs-Prisma enum drift regression tests.
 *
 * These tests assert the CORRECT behaviour and are expected to FAIL until the
 * schemas are aligned with prisma/schema.prisma. Each failure is a real defect:
 * a value that passes API validation but is then rejected by the database at
 * insert time, surfacing to the user as a generic 500.
 *
 * See BUG-03/04/05 in SECURITY_AUDIT_REPORT.md.
 */
import { describe, it, expect } from 'vitest'
import {
  createPekerjaanSchema,
  createProyekSchema,
  createInvoiceSchema,
} from '@/lib/schemas'

// Values taken directly from prisma/schema.prisma
const PRISMA_KATEGORI_PEKERJAAN = [
  'persiapan','pondasi','beton','kanopi','baja','tangga','atap','dinding',
  'plesteran','acian','keramik','paving','pengecatan','pintu','interior',
  'toilet','mep','custom',
]
const PRISMA_TIPE_PROYEK = ['gedung', 'infra']
const PRISMA_STATUS_INVOICE = ['draft', 'sent', 'paid']

function zodEnumValues(schema: any, key: string): string[] {
  const def = schema.shape?.[key]
  const inner = def?._def?.innerType ?? def
  return inner?._def?.entries
    ? Object.values(inner._def.entries) as string[]
    : (inner?._def?.values ?? [])
}

describe('Zod schema vs Prisma enum drift', () => {
  it('pekerjaan kategori must be a subset of Prisma KategoriPekerjaan', () => {
    const zodVals = zodEnumValues(createPekerjaanSchema, 'kategori')
    const invalid = zodVals.filter(v => !PRISMA_KATEGORI_PEKERJAAN.includes(v))
    console.log('zod kategori:', zodVals)
    console.log('NOT accepted by DB:', invalid)
    expect(invalid).toEqual([])
  })

  it('proyek jenisProyek must be a subset of Prisma TipeProyek', () => {
    const zodVals = zodEnumValues(createProyekSchema, 'jenisProyek')
    const invalid = zodVals.filter(v => !PRISMA_TIPE_PROYEK.includes(v))
    console.log('zod jenisProyek:', zodVals, '| NOT accepted by DB:', invalid)
    expect(invalid).toEqual([])
  })

  it('invoice status must be a subset of Prisma StatusInvoice', () => {
    const zodVals = zodEnumValues(createInvoiceSchema, 'status')
    const invalid = zodVals.filter(v => !PRISMA_STATUS_INVOICE.includes(v))
    console.log('zod status:', zodVals, '| NOT accepted by DB:', invalid)
    expect(invalid).toEqual([])
  })
})
