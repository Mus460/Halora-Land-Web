import { NextRequest, NextResponse } from 'next/server'
import { createRouteHandlerClient } from '@/lib/supabase-auth'

/**
 * Update Password
 * Updates user password after password reset flow
 */
export async function POST(request: NextRequest) {
  try {
    const { password } = await request.json()

    if (!password) {
      return NextResponse.json(
        { error: 'Password harus diisi' },
        { status: 400 }
      )
    }

    if (password.length < 6) {
      return NextResponse.json(
        { error: 'Password minimal 6 karakter' },
        { status: 400 }
      )
    }

    const supabase = await createRouteHandlerClient(request)
    
    const { error } = await supabase.auth.updateUser({
      password,
    })

    if (error) {
      throw error
    }

    return NextResponse.json({
      message: 'Password berhasil diubah'
    })
  } catch (error: any) {
    console.error('Update password error:', error)
    return NextResponse.json(
      { error: error.message || 'Gagal mengubah password' },
      { status: 500 }
    )
  }
}
