import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentSupabaseUser } from '@/lib/supabase-auth'

/**
 * GET /api/dashboard/stats
 * Returns aggregated statistics for dashboard
 */
export async function GET(request: NextRequest) {
  try {
    const session = await getCurrentSupabaseUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Get user's projects
    const proyekCount = await prisma.proyek.count({
      where: { userId: session.userId }
    })

    // For now, assume all projects are active (no status field in schema)
    const proyekAktif = proyekCount

    // Get total RAB from all user's projects
    // Calculate total RAB from all pekerjaan
    const allPekerjaan = await prisma.pekerjaan.findMany({
      where: {
        proyek: {
          userId: session.userId
        }
      },
      select: {
        hargaSatuan: true,
        volume: true
      }
    })
    const totalRAB = allPekerjaan.reduce((sum, p) => {
      const harga = Number(p.hargaSatuan) || 0
      const vol = Number(p.volume) || 0
      return sum + (harga * vol)
    }, 0)

    // Get total pekerjaan count
    const totalPekerjaan = allPekerjaan.length

    // Get recent projects
    const recentProjects = await prisma.proyek.findMany({
      where: { userId: session.userId },
      orderBy: { createdAt: 'desc' },
      take: 5,
      select: {
        id: true,
        namaProyek: true,
        lokasi: true,
        createdAt: true,
        pekerjaan: {
          select: {
            hargaSatuan: true,
            volume: true
          }
        }
      }
    })

    // Calculate totalRAB per project
    const recentProjectsWithRAB = recentProjects.map(p => {
      const projectRAB = p.pekerjaan.reduce((sum, pk) => {
        const harga = Number(pk.hargaSatuan) || 0
        const vol = Number(pk.volume) || 0
        return sum + (harga * vol)
      }, 0)
      return {
        id: p.id,
        nama: p.namaProyek,
        lokasi: p.lokasi,
        totalRAB: projectRAB,
        createdAt: p.createdAt,
      }
    })

    // Get recent activity (last 10 audit logs)
    const recentActivity = await prisma.auditLog.findMany({
      where: { userId: session.userId },
      orderBy: { createdAt: 'desc' },
      take: 10,
      select: {
        id: true,
        action: true,
        entityType: true,
        entityId: true,
        createdAt: true,
      }
    })

    return NextResponse.json({
      stats: {
        totalProyek: proyekCount,
        proyekAktif,
        totalRAB,
        totalPekerjaan,
      },
      recentProjects: recentProjectsWithRAB,
      recentActivity,
    })
  } catch (error) {
    console.error('Dashboard stats error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
