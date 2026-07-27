import { describe, it, expect, vi, beforeEach } from 'vitest'
import { logAudit, getClientInfo } from '@/lib/audit-log'
import type { AuditLogData } from '@/lib/audit-log'

// Mock Prisma
vi.mock('@/lib/prisma', () => ({
  prisma: {
    auditLog: {
      create: vi.fn().mockResolvedValue({
        id: 1,
        userId: 123,
        action: 'LOGIN',
        entityType: 'USER',
        createdAt: new Date(),
      }),
    },
  },
}))

describe('logAudit', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should log audit data successfully', async () => {
    const auditData: AuditLogData = {
      userId: 123,
      action: 'LOGIN',
      resource: 'USER',
      resourceId: 123,
      ip: '192.168.1.1',
      userAgent: 'Mozilla/5.0',
    }

    await expect(logAudit(auditData)).resolves.not.toThrow()
  })

  it('should handle missing optional fields', async () => {
    const auditData: AuditLogData = {
      userId: 456,
      action: 'CREATE',
      resource: 'PROYEK',
    }

    await expect(logAudit(auditData)).resolves.not.toThrow()
  })

  it('should handle errors gracefully without throwing', async () => {
    const { prisma } = await import('@/lib/prisma')
    vi.mocked(prisma.auditLog.create).mockRejectedValueOnce(new Error('DB Error'))

    const auditData: AuditLogData = {
      userId: 789,
      action: 'DELETE',
      resource: 'PROYEK',
    }

    // Should not throw even when DB fails
    await expect(logAudit(auditData)).resolves.not.toThrow()
  })

  it('should log with metadata', async () => {
    const auditData: AuditLogData = {
      userId: 111,
      action: 'UPDATE',
      resource: 'PROYEK',
      resourceId: 555,
      metadata: {
        oldValue: { nama: 'Old Name' },
        newValue: { nama: 'New Name' },
      },
    }

    await expect(logAudit(auditData)).resolves.not.toThrow()
  })
})

describe('getClientInfo', () => {
  it('should extract IP from x-forwarded-for header', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') return '203.0.113.1, 198.51.100.1'
          if (name === 'user-agent') return 'Mozilla/5.0 (Test)'
          return null
        }
      }
    } as any

    const info = getClientInfo(mockRequest)
    expect(info.ip).toBe('203.0.113.1')
    expect(info.userAgent).toBe('Mozilla/5.0 (Test)')
  })

  it('should handle missing x-forwarded-for header', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'user-agent') return 'Mozilla/5.0'
          return null
        }
      }
    } as any

    const info = getClientInfo(mockRequest)
    expect(info.ip).toBe('unknown')
    expect(info.userAgent).toBe('Mozilla/5.0')
  })

  it('should handle missing user-agent header', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') return '203.0.113.1'
          return null
        }
      }
    } as any

    const info = getClientInfo(mockRequest)
    expect(info.ip).toBe('203.0.113.1')
    expect(info.userAgent).toBe('unknown')
  })

  it('should handle completely missing headers', () => {
    const mockRequest = {
      headers: {
        get: () => null
      }
    } as any

    const info = getClientInfo(mockRequest)
    expect(info.ip).toBe('unknown')
    expect(info.userAgent).toBe('unknown')
  })

  it('should trim whitespace from IP', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') return '  203.0.113.1  , 198.51.100.1  '
          if (name === 'user-agent') return 'Test Agent'
          return null
        }
      }
    } as any

    const info = getClientInfo(mockRequest)
    expect(info.ip).toBe('203.0.113.1')
  })

  it('should handle multiple IPs in x-forwarded-for', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') return '10.0.0.1, 172.16.0.1, 192.168.1.1'
          if (name === 'user-agent') return 'Test'
          return null
        }
      }
    } as any

    const info = getClientInfo(mockRequest)
    expect(info.ip).toBe('10.0.0.1') // Should get first IP
  })
})

describe('Audit Action Types', () => {
  it('should support all defined audit actions', async () => {
    const actions: Array<AuditLogData['action']> = [
      'CREATE',
      'UPDATE',
      'DELETE',
      'LOGIN',
      'LOGOUT',
      'REGISTER',
      'EXPORT',
    ]

    for (const action of actions) {
      const auditData: AuditLogData = {
        userId: 123,
        action,
        resource: 'USER',
      }

      await expect(logAudit(auditData)).resolves.not.toThrow()
    }
  })
})

describe('Audit Resource Types', () => {
  it('should support all defined resource types', async () => {
    const resources: Array<AuditLogData['resource']> = [
      'USER',
      'PROYEK',
      'PEKERJAAN',
      'MASTER_HARGA',
      'MASTER_ANALISA',
      'INVOICE',
      'LOGISTIK',
      'REALISASI',
      'FEEDBACK',
      'NEWS',
    ]

    for (const resource of resources) {
      const auditData: AuditLogData = {
        userId: 123,
        action: 'CREATE',
        resource,
      }

      await expect(logAudit(auditData)).resolves.not.toThrow()
    }
  })
})
