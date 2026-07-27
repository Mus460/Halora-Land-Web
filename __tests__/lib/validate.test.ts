import { describe, it, expect } from 'vitest'
import { validateEmail, validatePassword, parseIntParam } from '@/lib/validate'

describe('validateEmail', () => {
  it('should accept valid email addresses', () => {
    expect(validateEmail('user@example.com')).toBe(true)
    expect(validateEmail('test.user@domain.co.id')).toBe(true)
    expect(validateEmail('admin+tag@company.org')).toBe(true)
  })

  it('should reject invalid email addresses', () => {
    expect(validateEmail('invalid')).toBe(false)
    expect(validateEmail('invalid@')).toBe(false)
    expect(validateEmail('@domain.com')).toBe(false)
    expect(validateEmail('user@')).toBe(false)
    expect(validateEmail('user domain.com')).toBe(false)
  })

  it('should reject empty string', () => {
    expect(validateEmail('')).toBe(false)
  })
})

describe('validatePassword', () => {
  it('should accept strong passwords', () => {
    expect(validatePassword('Password123')).toBeNull()
    expect(validatePassword('SecureP@ss1')).toBeNull()
    expect(validatePassword('MyP@ssw0rd')).toBeNull()
  })

  it('should reject password shorter than 8 characters', () => {
    expect(validatePassword('Pass1')).toBe('Minimal 8 karakter')
    expect(validatePassword('Aa1')).toBe('Minimal 8 karakter')
  })

  it('should reject password without uppercase letter', () => {
    expect(validatePassword('password123')).toBe('Harus ada huruf besar')
  })

  it('should reject password without lowercase letter', () => {
    expect(validatePassword('PASSWORD123')).toBe('Harus ada huruf kecil')
  })

  it('should reject password without number', () => {
    expect(validatePassword('PasswordOnly')).toBe('Harus ada angka')
  })

  it('should reject empty password', () => {
    expect(validatePassword('')).toBe('Minimal 8 karakter')
  })
})

describe('parseIntParam', () => {
  it('should parse valid integer strings', () => {
    expect(parseIntParam('123')).toBe(123)
    expect(parseIntParam('0')).toBe(0)
    expect(parseIntParam('999999')).toBe(999999)
  })

  it('should return null for invalid inputs', () => {
    expect(parseIntParam('abc')).toBeNull()
    expect(parseIntParam('12.34')).toBe(12) // parseInt truncates
    expect(parseIntParam(undefined)).toBeNull()
  })

  it('should return null for array inputs', () => {
    expect(parseIntParam(['123'])).toBeNull()
    expect(parseIntParam(['1', '2'])).toBeNull()
  })

  it('should handle negative numbers', () => {
    expect(parseIntParam('-123')).toBe(-123)
  })
})
