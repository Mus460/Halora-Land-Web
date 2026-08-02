import { cookies } from 'next/headers'
import { createServerClient } from '@supabase/ssr'
import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const requestUrl = new URL(request.url)
  const code = requestUrl.searchParams.get('code')
  const type = requestUrl.searchParams.get('type')

  if (code) {
    try {
      const cookieStore = await cookies()
      const supabase = createServerClient(
        process.env.NEXT_PUBLIC_SUPABASE_URL!,
        process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
        {
          cookies: {
            get(name: string) { return cookieStore.get(name)?.value },
            set(name: string, value: string, options: any) { cookieStore.set({ name, value, ...options }) },
            remove(name: string, options: any) { cookieStore.set({ name, value: '', ...options }) },
          },
        },
      )
      const { error } = await supabase.auth.exchangeCodeForSession(code)
      if (error) {
        return NextResponse.redirect(new URL(`/login?error=${encodeURIComponent(error.message)}`, requestUrl.origin))
      }
    } catch (error: any) {
      return NextResponse.redirect(new URL(`/login?error=${encodeURIComponent(error.message || 'Authentication failed')}`, requestUrl.origin))
    }
  }

  if (type === 'recovery') {
    return NextResponse.redirect(new URL('/reset-password', requestUrl.origin))
  }
  return NextResponse.redirect(new URL('/login?verified=true', requestUrl.origin))
}
