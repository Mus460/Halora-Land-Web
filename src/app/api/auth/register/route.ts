import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { signUpWithPassword } from '@/lib/supabase-auth'
import { rateLimit, getClientIp } from '@/lib/rate-limit'
import { logAudit, getClientInfo } from '@/lib/audit-log'

export async function POST(request: NextRequest) {
  // Rate limit: 5 requests per 15 minutes per IP
  const ip = getClientIp(request)
  const rateLimitResult = rateLimit({
    id: `register:${ip}`,
    limit: 5,
    windowSec: 15 * 60,
  })

  if (!rateLimitResult.success) {
    return NextResponse.json(
      { 
        error: 'Terlalu banyak percobaan. Coba lagi nanti.',
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
    const { namaLengkap, email, password } = await request.json()

    if (!namaLengkap || !email || !password) {
      return NextResponse.json(
        { error: 'Semua field harus diisi' },
        { status: 400 }
      )
    }

    if (password.length < 6) {
      return NextResponse.json(
        { error: 'Password minimal 6 karakter' },
        { status: 400 }
      )
    }

    // Check if email already exists in our User table
    const existingUser = await prisma.user.findUnique({
      where: { email },
    })

    if (existingUser) {
      return NextResponse.json(
        { error: 'Email sudah terdaftar' },
        { status: 409 }
      )
    }

    // Sign up with Supabase Auth
    const { user: supabaseUser } = await signUpWithPassword(email, password, {
      namaLengkap,
    })

    if (!supabaseUser) {
      return NextResponse.json(
        { error: 'Gagal membuat akun' },
        { status: 500 }
      )
    }

    // Create user in our User table
    const user = await prisma.user.create({
      data: {
        namaLengkap,
        email,
        password: '', // No longer needed, Supabase handles auth
        role: 'USER',
        accountType: 'FREE',
        isDemo: false,
        supabaseAuthId: supabaseUser.id,
      },
    })

    // Audit log
    const clientInfo = getClientInfo(request)
    await logAudit({
      userId: user.id,
      action: 'REGISTER',
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
      message: 'Registrasi berhasil. Silakan cek email untuk verifikasi.',
    }, { status: 201 })
  } catch (error: any) {
    console.error('Register error:', error)
    return NextResponse.json(
      { error: error.message || 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
