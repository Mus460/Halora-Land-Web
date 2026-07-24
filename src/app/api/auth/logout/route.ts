import { NextResponse } from 'next/server'
import { deleteAuthCookie } from '@/lib/session'

export async function POST() {
  try {
    await deleteAuthCookie()
    return NextResponse.json({ message: 'Logout berhasil' })
  } catch (error) {
    console.error('Logout error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
