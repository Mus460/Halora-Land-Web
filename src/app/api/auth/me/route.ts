import { NextRequest, NextResponse } from 'next/server'
import { getCurrentSupabaseUser } from '@/lib/supabase-auth'
import { prisma } from '@/lib/prisma'
import { z } from 'zod'

const profileUpdateSchema = z.object({
  namaLengkap: z.string().min(2).max(100).optional(),
  email: z.string().email('Format email tidak valid').optional()
})

export async function GET() {
  try {
    const session = await getCurrentSupabaseUser()
    
    if (!session) {
      return NextResponse.json(
        { error: 'Unauthorized' },
        { status: 401 }
      )
    }

    const user = await prisma.user.findUnique({
      where: { id: session.userId },
      select: {
        id: true,
        namaLengkap: true,
        email: true,
        role: true,
        accountType: true,
        isDemo: true,
        createdAt: true,
        updatedAt: true,
      },
    })

    if (!user) {
      return NextResponse.json(
        { error: 'User not found' },
        { status: 404 }
      )
    }

    return NextResponse.json({ user })
  } catch (error) {
    console.error('Get user error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}

export async function PUT(request: NextRequest) {
  try {
    const session = await getCurrentSupabaseUser()
    
    if (!session) {
      return NextResponse.json(
        { error: 'Unauthorized' },
        { status: 401 }
      )
    }

    const body = await request.json()
    const { namaLengkap, email } = profileUpdateSchema.parse(body)

    if (email && email !== session.email) {
      const existing = await prisma.user.findUnique({
        where: { email }
      })
      
      if (existing) {
        return NextResponse.json(
          { error: 'Email sudah digunakan' },
          { status: 409 }
        )
      }
    }

    const user = await prisma.user.update({
      where: { id: session.userId },
      data: {
        namaLengkap: namaLengkap || undefined,
        email: email || undefined,
      },
      select: {
        id: true,
        namaLengkap: true,
        email: true,
        role: true,
        accountType: true,
        isDemo: true,
        createdAt: true,
        updatedAt: true,
      },
    })

    return NextResponse.json({ user })
  } catch (error) {
    console.error('Update user error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
