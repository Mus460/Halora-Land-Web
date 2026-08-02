import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || ''
  const projectRef = supabaseUrl.split('//')[1]?.split('.')[0]
  if (!projectRef) {
    return NextResponse.next()
  }

  const cookieName = `sb-${projectRef}-auth-token`
  const sessionCookie = request.cookies.get(cookieName)
  let hasSession = false
  if (sessionCookie) {
    try {
      const session = JSON.parse(sessionCookie.value)
      const expiresAt = session.expires_at
      hasSession = !(expiresAt && new Date(expiresAt * 1000) < new Date())
    } catch {
      hasSession = !!sessionCookie.value
    }
  }

  if (pathname.startsWith('/dashboard') && !hasSession) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  if (pathname === '/login' && hasSession) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/dashboard/:path*', '/login', '/admin/:path*'],
}
