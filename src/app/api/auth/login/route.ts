import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { signInWithPassword } from '@/lib/supabase-auth'
import { rateLimit, getClientIp } from '@/lib/rate-limit'
import { logAudit, getClientInfo } from '@/lib/audit-log'

export async function POST(request: NextRequest) {
  // Rate limit: 10 requests per 15 minutes per IP
  const ip = getClientIp(request)
  const rateLimitResult = rateLimit({
    id: `login:${ip}`,
    limit: 10,
    windowSec: 15 * 60,
  })

  if (!rateLimitResult.success) {
    return NextResponse.json(
      { 
        error: 'Terlalu banyak percobaan login. Coba lagi nanti.',
        retryAfter: Math.ceil((rateLimitResult.reset - Date.now()) / 1000),
      },
      { 
        status: 429,
        headers: {
          'Retry-After': String(Math.ceil((rateLimitResult.reset - Date.now()) / 1000)),
          'X-RateLimit-Limit': String(rateLimitResult.limit),
          'X-RateLimit-Remaining': String(rateLimitResult.remaining),
          'X-RateLimit-Reset': String(rateLimitResult.reset),
        }
      }
    )
  }

  try {
    const { email, password } = await request.json()

    if (!email || !password) {
      return NextResponse.json(
        { error: 'Email dan password harus diisi' },
        { status: 400 }
      )
    }

    // Sign in with Supabase Auth
    const { user: supabaseUser, session } = await signInWithPassword(email, password)

    if (!supabaseUser || !session) {
      return NextResponse.json(
        { error: 'Email atau password salah' },
        { status: 401 }
      )
    }

    // Find or create user in our User table
    let user = await prisma.user.findFirst({
      where: {
        OR: [
          { supabaseAuthId: supabaseUser.id },
          { email: supabaseUser.email },
        ]
      }
    })

    // Link Supabase auth ID if user exists but not linked yet
    if (user && !user.supabaseAuthId) {
      user = await prisma.user.update({
        where: { id: user.id },
        data: { supabaseAuthId: supabaseUser.id }
      })
    }

    // If user doesn't exist in our table, create it (migration case)
    if (!user) {
      user = await prisma.user.create({
        data: {
          namaLengkap: supabaseUser.user_metadata?.namaLengkap || supabaseUser.email || 'User',
          email: supabaseUser.email!,
          password: '', // No longer needed
          role: 'USER',
          accountType: 'FREE',
          isDemo: false,
          supabaseAuthId: supabaseUser.id,
        }
      })
    }

    // Audit log
    const clientInfo = getClientInfo(request)
    await logAudit({
      userId: user.id,
      action: 'LOGIN',
      resource: 'USER',
      resourceId: user.id,
      ...clientInfo,
    })

    return NextResponse.json({
      user: {
        id: user.id,
        namaLengkap: user.namaLengkap,
        email: user.email,
        role: user.role,
        accountType: user.accountType,
        isDemo: user.isDemo,
      },
    })
  } catch (error: any) {
    console.error('Login error:', error)
    
    // Preserve exact error message for client detection
    const errorMessage = error.message || 'Email atau password salah'
    
    return NextResponse.json(
      { error: errorMessage },
      { status: 401 }
    )
  }
}
