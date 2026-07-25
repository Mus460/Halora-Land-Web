import { NextResponse } from 'next/server'
import { signOut } from '@/lib/supabase-auth'

export async function POST() {
  try {
    await signOut()
    return NextResponse.json({ message: 'Logout berhasil' })
  } catch (error: any) {
    console.error('Logout error:', error)
    return NextResponse.json(
      { error: error.message || 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
