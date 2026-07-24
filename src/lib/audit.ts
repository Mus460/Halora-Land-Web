import { prisma } from '@/lib/prisma'

/**
 * Input for creating audit log
 */
export interface AuditLogInput {
  proyekId?: number
  pekerjaanId?: number
  userId: number
  action: string
  entityType: string
  entityId?: number
  oldValue?: any
  newValue?: any
  description?: string
  ipAddress?: string
  userAgent?: string
}

/**
 * Create audit log entry
 * 
 * @param input - Audit log data
 * @returns Created audit log
 */
export async function createAuditLog(input: AuditLogInput) {
  return await prisma.auditLog.create({
    data: {
      proyekId: input.proyekId,
      pekerjaanId: input.pekerjaanId,
      userId: input.userId,
      action: input.action,
      entityType: input.entityType,
      entityId: input.entityId,
      oldValue: input.oldValue || null,
      newValue: input.newValue || null,
      description: input.description,
      ipAddress: input.ipAddress,
      userAgent: input.userAgent,
    }
  })
}

/**
 * Get audit trail for an entity
 * 
 * @param entityType - Type of entity (pekerjaan, master_harga, etc)
 * @param entityId - ID of entity
 * @param limit - Max number of logs to return
 * @returns Audit logs with user info
 */
export async function getAuditTrail(
  entityType: string,
  entityId: number,
  limit: number = 50
) {
  return await prisma.auditLog.findMany({
    where: {
      entityType,
      entityId
    },
    include: {
      user: {
        select: {
          id: true,
          namaLengkap: true,
          email: true
        }
      }
    },
    orderBy: {
      createdAt: 'desc'
    },
    take: limit
  })
}

/**
 * Get audit logs for a project
 * 
 * @param proyekId - Project ID
 * @param limit - Max number of logs
 * @returns Audit logs
 */
export async function getProjectAuditLogs(
  proyekId: number,
  limit: number = 100
) {
  return await prisma.auditLog.findMany({
    where: {
      proyekId
    },
    include: {
      user: {
        select: {
          id: true,
          namaLengkap: true,
          email: true
        }
      },
      pekerjaan: {
        select: {
          id: true,
          uraianPekerjaan: true
        }
      }
    },
    orderBy: {
      createdAt: 'desc'
    },
    take: limit
  })
}

/**
 * Get audit logs by action type
 * 
 * @param action - Action type (e.g., 'recalculate', 'price_update')
 * @param limit - Max number of logs
 * @returns Audit logs
 */
export async function getAuditLogsByAction(
  action: string,
  limit: number = 50
) {
  return await prisma.auditLog.findMany({
    where: {
      action
    },
    include: {
      user: {
        select: {
          id: true,
          namaLengkap: true,
          email: true
        }
      }
    },
    orderBy: {
      createdAt: 'desc'
    },
    take: limit
  })
}

/**
 * Get audit logs for a user
 * 
 * @param userId - User ID
 * @param limit - Max number of logs
 * @returns Audit logs
 */
export async function getUserAuditLogs(
  userId: number,
  limit: number = 100
) {
  return await prisma.auditLog.findMany({
    where: {
      userId
    },
    include: {
      proyek: {
        select: {
          id: true,
          namaProyek: true
        }
      },
      pekerjaan: {
        select: {
          id: true,
          uraianPekerjaan: true
        }
      }
    },
    orderBy: {
      createdAt: 'desc'
    },
    take: limit
  })
}
