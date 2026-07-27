import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { signInWithPassword } from '@/lib/supabase-auth'
import { rateLimit, getClientIp } from '@/lib/rate-limit'
import { logAudit, getClientInfo } from '@/lib/audit-log'
import { loginSchema } from '@/lib/schemas'
import { getJsonBody, handleError } from '@/lib/api-utils'

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
    const body = await getJsonBody(request)
    const { email, password } = loginSchema.parse(body)

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

    // Create response with user data
    const response = NextResponse.json({
      user: {
        id: user.id,
        namaLengkap: user.namaLengkap,
        email: user.email,
        role: user.role,
        accountType: user.accountType,
        isDemo: user.isDemo,
      },
    })

    // Set Supabase session cookies for SSR
    if (session) {
      // Extract project ref from Supabase URL
      const projectRef = process.env.NEXT_PUBLIC_SUPABASE_URL!.split('//')[1].split('.')[0]
      
      // Set auth token cookie (Supabase SSR format)
      response.cookies.set({
        name: `sb-${projectRef}-auth-token`,
        value: JSON.stringify({
          access_token: session.access_token,
          refresh_token: session.refresh_token,
          expires_at: session.expires_at,
          expires_in: session.expires_in,
          token_type: session.token_type,
          user: session.user,
        }),
        path: '/',
        maxAge: session.expires_in,
        httpOnly: true,
        secure: true,
        sameSite: 'strict',
      })
    }

    return response
  } catch {
    console.error('Login error')
    
    return NextResponse.json(
      { error: 'Email atau password salah' },
      { status: 401 }
    )
  }
}
