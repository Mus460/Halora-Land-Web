import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Middleware for route protection and authentication
 * 
 * Protects:
 * - /dashboard/* routes (require authentication)
 * 
 * Redirects:
 * - Unauthenticated users to /login
 * - Authenticated users from /login to /dashboard
 */
export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  
  // Extract project ref from Supabase URL
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || ''
  const projectRef = supabaseUrl.split('//')[1]?.split('.')[0]
  
  if (!projectRef) {
    console.error('NEXT_PUBLIC_SUPABASE_URL not configured')
    return NextResponse.next()
  }

  // Get session cookie
  const cookieName = `sb-${projectRef}-auth-token`
  const sessionCookie = request.cookies.get(cookieName)
  
  let session = null
  if (sessionCookie) {
    try {
      session = JSON.parse(sessionCookie.value)
      
      // Check if session is expired
      const expiresAt = session.expires_at
      if (expiresAt && new Date(expiresAt * 1000) < new Date()) {
        session = null // Session expired
      }
    } catch (error) {
      console.error('Failed to parse session cookie:', error)
      session = null
    }
  }

  // Protect dashboard routes
  if (pathname.startsWith('/dashboard')) {
    if (!session) {
      const loginUrl = new URL('/login', request.url)
      return NextResponse.redirect(loginUrl)
    }
  }

  // Redirect authenticated users from login to dashboard
  if (pathname === '/login' && session) {
    const dashboardUrl = new URL('/dashboard', request.url)
    return NextResponse.redirect(dashboardUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/login',
  ],
}
