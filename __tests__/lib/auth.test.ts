import { describe, it, expect, vi } from 'vitest'
import { hashPassword, verifyPassword, signToken, verifyToken } from '@/lib/auth'

describe('hashPassword', () => {
  it('should hash password successfully', async () => {
    const password = 'MySecurePassword123'
    const hashed = await hashPassword(password)
    
    expect(hashed).toBeTruthy()
    expect(hashed).not.toBe(password)
    expect(hashed.length).toBeGreaterThan(50) // bcrypt hashes are long
    expect(hashed.startsWith('$2')).toBe(true) // bcrypt hash format
  })

  it('should generate different hashes for same password', async () => {
    const password = 'SamePassword123'
    const hash1 = await hashPassword(password)
    const hash2 = await hashPassword(password)
    
    expect(hash1).not.toBe(hash2) // bcrypt uses random salt
  })
})

describe('verifyPassword', () => {
  it('should verify correct password', async () => {
    const password = 'TestPassword123'
    const hashed = await hashPassword(password)
    
    const isValid = await verifyPassword(password, hashed)
    expect(isValid).toBe(true)
  })

  it('should reject incorrect password', async () => {
    const password = 'TestPassword123'
    const wrongPassword = 'WrongPassword456'
    const hashed = await hashPassword(password)
    
    const isValid = await verifyPassword(wrongPassword, hashed)
    expect(isValid).toBe(false)
  })

  it('should reject empty password', async () => {
    const password = 'TestPassword123'
    const hashed = await hashPassword(password)
    
    const isValid = await verifyPassword('', hashed)
    expect(isValid).toBe(false)
  })
})

describe('signToken', () => {
  it('should create valid JWT token', async () => {
    const payload = {
      userId: 123,
      email: 'test@example.com',
      role: 'USER',
    }

    const token = await signToken(payload)
    
    expect(token).toBeTruthy()
    expect(typeof token).toBe('string')
    expect(token.split('.')).toHaveLength(3) // JWT format: header.payload.signature
  })

  it('should include payload data in token', async () => {
    const payload = {
      userId: 456,
      email: 'admin@example.com',
      role: 'ADMIN',
    }

    const token = await signToken(payload)
    const verified = await verifyToken(token)
    
    expect(verified).toBeTruthy()
    expect(verified?.userId).toBe(456)
    expect(verified?.email).toBe('admin@example.com')
    expect(verified?.role).toBe('ADMIN')
  })
})

describe('verifyToken', () => {
  it('should verify valid token', async () => {
    const payload = {
      userId: 789,
      email: 'user@example.com',
      role: 'USER',
    }

    const token = await signToken(payload)
    const verified = await verifyToken(token)
    
    expect(verified).not.toBeNull()
    expect(verified?.userId).toBe(789)
  })

  it('should reject invalid token', async () => {
    const invalidToken = 'invalid.token.here'
    
    const verified = await verifyToken(invalidToken)
    expect(verified).toBeNull()
  })

  it('should reject empty token', async () => {
    const verified = await verifyToken('')
    expect(verified).toBeNull()
  })

  it('should reject tampered token', async () => {
    const payload = {
      userId: 111,
      email: 'user@example.com',
      role: 'USER',
    }

    const token = await signToken(payload)
    const tampered = token.slice(0, -5) + 'XXXXX' // Tamper with signature
    
    const verified = await verifyToken(tampered)
    expect(verified).toBeNull()
  })
})

describe('JWT Security', () => {
  it('should have expiration time', async () => {
    const payload = {
      userId: 999,
      email: 'test@example.com',
      role: 'USER',
    }

    const token = await signToken(payload)
    const verified = await verifyToken(token)
    
    expect(verified?.exp).toBeTruthy()
    expect(verified?.exp).toBeGreaterThan(Date.now() / 1000) // Not expired yet
  })

  it('should have issued at timestamp', async () => {
    const payload = {
      userId: 888,
      email: 'test@example.com',
      role: 'USER',
    }

    const token = await signToken(payload)
    const verified = await verifyToken(token)
    
    expect(verified?.iat).toBeTruthy()
    expect(verified?.iat).toBeLessThanOrEqual(Date.now() / 1000)
  })
})
