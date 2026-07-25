import { createServerClient as createSupabaseServerClient } from '@supabase/ssr'
import { cookies } from 'next/headers'
import { NextRequest, NextResponse } from 'next/server'

/**
 * Create Supabase client for Server Components
 */
export async function createServerClient() {
  const cookieStore = await cookies()

  return createSupabaseServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        get(name: string) {
          return cookieStore.get(name)?.value
        },
        set(name: string, value: string, options: any) {
          cookieStore.set({ name, value, ...options })
        },
        remove(name: string, options: any) {
          cookieStore.set({ name, value: '', ...options })
        },
      },
    }
  )
}

/**
 * Create Supabase client for API Routes
 */
export async function createRouteHandlerClient(request: NextRequest) {
  const cookieStore = await cookies()

  return createSupabaseServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        get(name: string) {
          return cookieStore.get(name)?.value
        },
        set(name: string, value: string, options: any) {
          cookieStore.set({ name, value, ...options })
        },
        remove(name: string, options: any) {
          cookieStore.set({ name, value: '', ...options })
        },
      },
    }
  )
}

/**
 * Get current user session from Supabase Auth
 * Returns user data if authenticated, null otherwise
 */
export async function getSupabaseSession() {
  const supabase = await createServerClient()
  
  const { data: { session }, error } = await supabase.auth.getSession()
  
  if (error || !session) {
    return null
  }

  return {
    user: session.user,
    session,
  }
}

/**
 * Get current user with Prisma User data
 * Links Supabase auth with Prisma User table
 */
export async function getCurrentSupabaseUser() {
  const sessionData = await getSupabaseSession()
  
  if (!sessionData) {
    return null
  }

  const { prisma } = await import('./prisma')
  
  // Find user by Supabase auth ID or email
  const user = await prisma.user.findFirst({
    where: {
      OR: [
        { supabaseAuthId: sessionData.user.id },
        { email: sessionData.user.email },
      ]
    },
    select: {
      id: true,
      email: true,
      namaLengkap: true,
      role: true,
      accountType: true,
      isDemo: true,
      supabaseAuthId: true,
    }
  })

  if (!user) {
    return null
  }

  // Link Supabase auth ID if not already linked
  if (!user.supabaseAuthId) {
    await prisma.user.update({
      where: { id: user.id },
      data: { supabaseAuthId: sessionData.user.id }
    })
  }

  return {
    userId: user.id,
    email: user.email,
    role: user.role,
    namaLengkap: user.namaLengkap,
    accountType: user.accountType,
    isDemo: user.isDemo,
    supabaseUser: sessionData.user,
  }
}

/**
 * Sign in with email and password
 */
export async function signInWithPassword(email: string, password: string) {
  const supabase = await createServerClient()
  
  const { data, error } = await supabase.auth.signInWithPassword({
    email,
    password,
  })

  if (error) {
    throw new Error(error.message)
  }

  return data
}

/**
 * Sign up with email and password
 */
export async function signUpWithPassword(email: string, password: string, metadata?: any) {
  const supabase = await createServerClient()
  
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
    options: {
      data: metadata,
    }
  })

  if (error) {
    throw new Error(error.message)
  }

  return data
}

/**
 * Sign out
 */
export async function signOut() {
  const supabase = await createServerClient()
  
  const { error } = await supabase.auth.signOut()

  if (error) {
    throw new Error(error.message)
  }
}

/**
 * Send password reset email
 */
export async function sendPasswordResetEmail(email: string) {
  const supabase = await createServerClient()
  
  const { error } = await supabase.auth.resetPasswordForEmail(email, {
    redirectTo: `${process.env.NEXT_PUBLIC_APP_URL}/auth/reset-password`,
  })

  if (error) {
    throw new Error(error.message)
  }
}

/**
 * Update password
 */
export async function updatePassword(newPassword: string) {
  const supabase = await createServerClient()
  
  const { error } = await supabase.auth.updateUser({
    password: newPassword,
  })

  if (error) {
    throw new Error(error.message)
  }
}
