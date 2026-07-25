import { prisma } from './prisma'

export type AuditAction =
  | 'CREATE'
  | 'UPDATE'
  | 'DELETE'
  | 'LOGIN'
  | 'LOGOUT'
  | 'REGISTER'
  | 'EXPORT'

export type AuditResource =
  | 'USER'
  | 'PROYEK'
  | 'PEKERJAAN'
  | 'MASTER_HARGA'
  | 'MASTER_ANALISA'
  | 'INVOICE'
  | 'LOGISTIK'
  | 'REALISASI'
  | 'FEEDBACK'
  | 'NEWS'

export interface AuditLogData {
  userId: number
  action: AuditAction
  resource: AuditResource
  resourceId?: number | string
  metadata?: Record<string, any>
  ip?: string
  userAgent?: string
}

/**
 * Log audit trail to database
 * Non-blocking - errors are logged but don't throw
 */
export async function logAudit(data: AuditLogData): Promise<void> {
  try {
    await prisma.auditLog.create({
      data: {
        userId: data.userId,
        action: data.action,
        entityType: data.resource,
        entityId: data.resourceId ? parseInt(data.resourceId.toString()) : null,
        description: data.metadata ? JSON.stringify(data.metadata) : null,
        ipAddress: data.ip,
        userAgent: data.userAgent,
      },
    })
  } catch (error) {
    console.error('Audit log failed:', error)
    // Don't throw - audit logging should not break user operations
  }
}

/**
 * Helper to extract client info from Request
 */
export function getClientInfo(request: Request) {
  const forwarded = request.headers.get('x-forwarded-for')
  const ip = forwarded ? forwarded.split(',')[0].trim() : 'unknown'
  const userAgent = request.headers.get('user-agent') || 'unknown'
  
  return { ip, userAgent }
}
