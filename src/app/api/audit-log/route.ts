import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

/**
 * GET /api/audit-log
 * Get audit logs with filters
 * Query params:
 *   - proyekId: filter by project
 *   - entityType: filter by entity type
 *   - action: filter by action
 *   - userId: filter by user
 *   - limit: max results (default 50)
 */
export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { searchParams } = new URL(request.url)
    const proyekId = searchParams.get('proyekId')
    const entityType = searchParams.get('entityType')
    const action = searchParams.get('action')
    const userId = searchParams.get('userId')
    const limit = parseInt(searchParams.get('limit') || '50')

    const where: any = {}

    // Filter by project
    if (proyekId) {
      const proyekIdInt = parseInt(proyekId)
      
      // Check access to project
      const proyek = await prisma.proyek.findUnique({
        where: { id: proyekIdInt },
        include: { timProyek: true }
      })

      if (!proyek) {
        return NextResponse.json({ error: 'Proyek not found' }, { status: 404 })
      }

      const hasAccess =
        session.role === 'ADMIN' ||
        proyek.userId === session.userId ||
        proyek.timProyek.some(t => t.userId === session.userId)

      if (!hasAccess) {
        return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
      }

      where.proyekId = proyekIdInt
    } else {
      // Non-admin can only see logs from their own projects
      if (session.role !== 'ADMIN') {
        const userProjects = await prisma.proyek.findMany({
          where: {
            OR: [
              { userId: session.userId },
              { timProyek: { some: { userId: session.userId } } }
            ]
          },
          select: { id: true }
        })

        const projectIds = userProjects.map(p => p.id)
        where.proyekId = { in: projectIds }
      }
    }

    // Filter by entity type
    if (entityType) {
      where.entityType = entityType
    }

    // Filter by action
    if (action) {
      where.action = action
    }

    // Filter by user
    if (userId) {
      where.userId = parseInt(userId)
    }

    const auditLogs = await prisma.auditLog.findMany({
      where,
      include: {
        user: {
          select: {
            id: true,
            namaLengkap: true,
            email: true
          }
        },
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

    // Get total count
    const total = await prisma.auditLog.count({ where })

    return NextResponse.json({
      total,
      limit,
      logs: auditLogs
    })
  } catch (error: any) {
    console.error('[audit-log GET]', error)
    return NextResponse.json(
      { error: 'Internal server error', message: error.message },
      { status: 500 }
    )
  }
}
