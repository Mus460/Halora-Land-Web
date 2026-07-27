import { z } from 'zod'

// === AUTH ===
export const loginSchema = z.object({
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(8, 'Password minimal 8 karakter')
})

export const registerSchema = z.object({
  email: z.string().email('Format email tidak valid'),
  password: z.string()
    .min(8, 'Password minimal 8 karakter')
    .regex(/[A-Z]/, 'Password harus ada huruf besar')
    .regex(/[a-z]/, 'Password harus ada huruf kecil')
    .regex(/[0-9]/, 'Password harus ada angka'),
  name: z.string().min(2, 'Nama minimal 2 karakter').max(100)
})

// === PEKERJAAN ===
export const createPekerjaanSchema = z.object({
  proyekId: z.number().int().positive(),
  kategori: z.enum([
    'pekerjaan_persiapan',
    'pekerjaan_tanah',
    'pekerjaan_dinding',
    'pekerjaan_atap',
    'pekerjaan_plafon',
    'pekerjaan_lantai',
    'pekerjaan_kusen',
    'pekerjaan_finish'
  ]),
  uraianPekerjaan: z.string().min(1).max(500),
  volume: z.number().positive('Volume harus > 0'),
  satuan: z.string().min(1).max(50),
  metodeHitung: z.enum(['ahsp', 'manual']),
  levelPekerjaan: z.number().int().min(1).max(3).default(1),
  tipePekerjaan: z.enum(['UTAMA', 'SUB']).default('UTAMA'),
  masterAnalisaId: z.number().int().positive().optional(),
  hargaSatuan: z.number().min(0).max(999_999_999_999).optional(),
  detailAnalisa: z.array(z.object({
    komponen: z.string().min(1).max(200),
    koefisien: z.number().min(0),
    satuan: z.string().min(1).max(50),
    hargaSatuan: z.number().min(0),
  })).optional(),
})

// === PROYEK ===
export const createProyekSchema = z.object({
  namaProyek: z.string().min(3, 'Nama proyek minimal 3 karakter').max(200),
  jenisProyek: z.enum(['gedung', 'infrastruktur']),
  lokasi: z.string().min(1).max(500),
  tanggalMulai: z.string().datetime().optional(),
  tanggalSelesai: z.string().datetime().optional(),
  nilaiKontrak: z.number().min(0).max(999_999_999_999).optional(),
  owner: z.string().min(1).max(200).optional(),
  deskripsi: z.string().max(2000).optional(),
})

// === FEEDBACK ===
export const feedbackSchema = z.object({
  message: z.string().min(10, 'Pesan minimal 10 karakter').max(2000),
  rating: z.number().int().min(1).max(5).optional(),
  category: z.enum(['bug', 'feature', 'question', 'other']).default('other')
})

// === MASTER HARGA ===
export const createMasterHargaSchema = z.object({
  nama: z.string().min(1, 'Nama harus diisi').max(200),
  satuan: z.string().min(1, 'Satuan harus diisi').max(50),
  harga: z.number().min(0, 'Harga harus >= 0').max(999_999_999_999),
  kategori: z.enum(['material', 'upah', 'alat']),
  isGlobal: z.boolean().optional(),
})

// === INVOICE ===
export const createInvoiceSchema = z.object({
  tanggal: z.string().datetime(),
  total: z.number().min(0).max(999_999_999_999),
  status: z.enum(['draft', 'sent', 'paid', 'cancelled']).default('draft'),
})

// === LOGISTIK ===
export const createLogistikSchema = z.object({
  namaMaterial: z.string().min(1, 'Nama material harus diisi').max(200),
  satuan: z.string().min(1, 'Satuan harus diisi').max(50),
  volume: z.number().positive('Volume harus > 0'),
  hargaSatuan: z.number().min(0, 'Harga satuan harus >= 0'),
  tanggal: z.string().datetime(),
})

// === REALISASI ===
export const createRealisasiSchema = z.object({
  tanggal: z.string().datetime(),
  kategori: z.string().min(1, 'Kategori harus diisi').max(100),
  jumlah: z.number().min(0, 'Jumlah harus >= 0'),
  keterangan: z.string().max(500).optional(),
})

// === MASTER ANALISA ===
export const createMasterAnalisaSchema = z.object({
  kode: z.string().min(1, 'Kode harus diisi').max(50),
  nama: z.string().min(1, 'Nama harus diisi').max(200),
  level: z.number().int().min(0).max(4),
  parentId: z.number().int().positive().optional(),
  satuan: z.string().max(50).optional(),
  isGlobal: z.boolean().optional(),
})

// === TYPE EXPORTS ===
export type LoginInput = z.infer<typeof loginSchema>
export type RegisterInput = z.infer<typeof registerSchema>
export type CreatePekerjaanInput = z.infer<typeof createPekerjaanSchema>
export type CreateProyekInput = z.infer<typeof createProyekSchema>
export type CreateMasterHargaInput = z.infer<typeof createMasterHargaSchema>
