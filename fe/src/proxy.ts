import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const SESSION_COOKIE = 'halora_session'

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl

  const sessionCookie = request.cookies.get(SESSION_COOKIE)
  const hasSession = isSessionValid(sessionCookie?.value)

  if (pathname.startsWith('/dashboard') && !hasSession) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  if (pathname === '/login' && hasSession) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}

// The session cookie is an HS256 JWT (uid/exp/iat). This only checks the
// unauthenticated exp claim to keep the edge guard cheap; the backend verifies
// the signature on every API call.
function isSessionValid(value: string | undefined): boolean {
  if (!value) return false
  const parts = value.split('.')
  if (parts.length !== 3) return false
  try {
    const b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const payload = JSON.parse(atob(b64))
    if (typeof payload.exp === 'number' && Date.now() >= payload.exp * 1000) {
      return false
    }
  } catch {
    return false
  }
  return true
}

export const config = {
  matcher: ['/dashboard/:path*', '/login', '/admin/:path*'],
}
