import { NextRequest, NextResponse } from 'next/server'
import { createRouteHandlerClient } from '@/lib/supabase-auth'
import { rateLimit, getClientIp } from '@/lib/rate-limit'

/**
 * Resend Email Verification
 * Allows users to request a new verification email
 * Rate limited: 3 requests per 15 minutes per email
 */
export async function POST(request: NextRequest) {
  const ip = getClientIp(request)
  
  try {
    const { email } = await request.json()

    if (!email) {
      return NextResponse.json(
        { error: 'Email harus diisi' },
        { status: 400 }
      )
    }

    // Rate limit: 3 requests per 15 minutes per email
    const rateLimitResult = rateLimit({
      id: `resend:${email}`,
      limit: 3,
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

    const supabase = await createRouteHandlerClient(request)
    
    const { error } = await supabase.auth.resend({
      type: 'signup',
      email,
    })

    if (error) {
      throw error
    }

    return NextResponse.json({
      message: 'Email verifikasi telah dikirim. Cek inbox Anda.'
    })
  } catch (error: any) {
    console.error('Resend verification error:', error)
    return NextResponse.json(
      { error: error.message || 'Gagal mengirim email' },
      { status: 500 }
    )
  }
}
