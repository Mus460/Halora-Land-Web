import { describe, it, expect } from 'vitest'
import { loginSchema, registerSchema, createProyekSchema, createPekerjaanSchema } from '@/lib/schemas'
import { z } from 'zod'

describe('loginSchema', () => {
  it('should accept valid login data', () => {
    const valid = {
      email: 'user@example.com',
      password: 'Password123',
    }
    expect(() => loginSchema.parse(valid)).not.toThrow()
  })

  it('should reject invalid email', () => {
    const invalid = {
      email: 'invalid-email',
      password: 'Password123',
    }
    expect(() => loginSchema.parse(invalid)).toThrow(z.ZodError)
  })

  it('should reject password shorter than 8 characters', () => {
    const invalid = {
      email: 'user@example.com',
      password: 'Pass1',
    }
    expect(() => loginSchema.parse(invalid)).toThrow(z.ZodError)
  })

  it('should reject missing fields', () => {
    expect(() => loginSchema.parse({ email: 'test@test.com' })).toThrow()
    expect(() => loginSchema.parse({ password: 'Password123' })).toThrow()
  })
})

describe('registerSchema', () => {
  it('should accept valid registration data', () => {
    const valid = {
      email: 'user@example.com',
      password: 'Password123',
      namaLengkap: 'John Doe',
    }
    expect(() => registerSchema.parse(valid)).not.toThrow()
  })

  it('should reject password without uppercase', () => {
    const invalid = {
      email: 'user@example.com',
      password: 'password123',
      name: 'John Doe',
    }
    expect(() => registerSchema.parse(invalid)).toThrow(/huruf besar/)
  })

  it('should reject password without lowercase', () => {
    const invalid = {
      email: 'user@example.com',
      password: 'PASSWORD123',
      name: 'John Doe',
    }
    expect(() => registerSchema.parse(invalid)).toThrow(/huruf kecil/)
  })

  it('should reject password without number', () => {
    const invalid = {
      email: 'user@example.com',
      password: 'PasswordOnly',
      namaLengkap: 'John Doe',
    }
    expect(() => registerSchema.parse(invalid)).toThrow(/angka/)
  })

  it('should reject name shorter than 2 characters', () => {
    const invalid = {
      email: 'user@example.com',
      password: 'Password123',
      namaLengkap: 'J',
    }
    expect(() => registerSchema.parse(invalid)).toThrow(/minimal 2 karakter/)
  })

  it('should reject name longer than 100 characters', () => {
    const invalid = {
      email: 'user@example.com',
      password: 'Password123',
      namaLengkap: 'A'.repeat(101),
    }
    expect(() => registerSchema.parse(invalid)).toThrow()
  })
})

describe('createProyekSchema', () => {
  it('should accept valid project data', () => {
    const valid = {
      namaProyek: 'Pembangunan Gedung',
      jenisProyek: 'gedung' as const,
      lokasi: 'Jakarta',
      nilaiKontrak: 1000000000,
    }
    expect(() => createProyekSchema.parse(valid)).not.toThrow()
  })

  it('should reject project name shorter than 3 characters', () => {
    const invalid = {
      namaProyek: 'AB',
      jenisProyek: 'gedung' as const,
      lokasi: 'Jakarta',
    }
    expect(() => createProyekSchema.parse(invalid)).toThrow(/minimal 3 karakter/)
  })

  it('should reject invalid jenisProyek', () => {
    const invalid = {
      namaProyek: 'Valid Project',
      jenisProyek: 'invalid-type',
      lokasi: 'Jakarta',
    }
    expect(() => createProyekSchema.parse(invalid)).toThrow()
  })

  it('should accept optional fields', () => {
    const minimal = {
      namaProyek: 'Valid Project',
      jenisProyek: 'gedung' as const,
      lokasi: 'Jakarta',
    }
    expect(() => createProyekSchema.parse(minimal)).not.toThrow()
  })
})

describe('createPekerjaanSchema', () => {
  it('should accept valid pekerjaan data', () => {
    const valid = {
      proyekId: 1,
      kategori: 'pondasi' as const,
      uraianPekerjaan: 'Galian tanah pondasi',
      volume: 100,
      satuan: 'm3',
      metodeHitung: 'manual' as const,
      levelPekerjaan: '1',
      tipePekerjaan: 'UTAMA' as const,
    }
    expect(() => createPekerjaanSchema.parse(valid)).not.toThrow()
  })

  it('should reject negative volume', () => {
    const invalid = {
      proyekId: 1,
      kategori: 'pondasi' as const,
      uraianPekerjaan: 'Test',
      volume: -10,
      satuan: 'm3',
      metodeHitung: 'manual' as const,
    }
    expect(() => createPekerjaanSchema.parse(invalid)).toThrow(/Volume harus > 0/)
  })

  it('should reject zero volume', () => {
    const invalid = {
      proyekId: 1,
      kategori: 'pondasi' as const,
      uraianPekerjaan: 'Test',
      volume: 0,
      satuan: 'm3',
      metodeHitung: 'manual' as const,
    }
    expect(() => createPekerjaanSchema.parse(invalid)).toThrow(/Volume harus > 0/)
  })

  it('should reject invalid kategori', () => {
    const invalid = {
      proyekId: 1,
      kategori: 'invalid_category',
      uraianPekerjaan: 'Test',
      volume: 100,
      satuan: 'm3',
      metodeHitung: 'manual' as const,
    }
    expect(() => createPekerjaanSchema.parse(invalid)).toThrow()
  })
})
