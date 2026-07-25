import { createRouteHandlerClient } from '@/lib/supabase-auth'
import { NextRequest, NextResponse } from 'next/server'

/**
 * Auth Callback Route
 * Handles Supabase email verification and password reset callbacks
 * 
 * Flow:
 * 1. User clicks email link → Supabase redirects here with code
 * 2. Exchange code for session
 * 3. Redirect based on type:
 *    - signup → /login?verified=true
 *    - recovery → /reset-password
 */
export async function GET(request: NextRequest) {
  const requestUrl = new URL(request.url)
  const code = requestUrl.searchParams.get('code')
  const type = requestUrl.searchParams.get('type')
  const next = requestUrl.searchParams.get('next') || '/login'

  if (code) {
    try {
      const supabase = await createRouteHandlerClient(request)
      const { error } = await supabase.auth.exchangeCodeForSession(code)
      
      if (error) {
        console.error('Auth callback error:', error)
        return NextResponse.redirect(
          new URL(`/login?error=${encodeURIComponent(error.message)}`, requestUrl.origin)
        )
      }
    } catch (error: any) {
      console.error('Auth callback exception:', error)
      return NextResponse.redirect(
        new URL(`/login?error=${encodeURIComponent(error.message || 'Authentication failed')}`, requestUrl.origin)
      )
    }
  }

  // Redirect based on type
  if (type === 'recovery') {
    // Password reset → redirect to reset password form
    return NextResponse.redirect(new URL('/reset-password', requestUrl.origin))
  }

  // Email verification → redirect to login with success message
  return NextResponse.redirect(
    new URL('/login?verified=true', requestUrl.origin)
  )
}
