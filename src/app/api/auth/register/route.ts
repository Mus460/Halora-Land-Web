import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { hashPassword, signToken } from '@/lib/auth'
import { setAuthCookie } from '@/lib/session'

export async function POST(request: NextRequest) {
  try {
    const { namaLengkap, email, password } = await request.json()

    if (!namaLengkap || !email || !password) {
      return NextResponse.json(
        { error: 'Semua field harus diisi' },
        { status: 400 }
      )
    }

    if (password.length < 6) {
      return NextResponse.json(
        { error: 'Password minimal 6 karakter' },
        { status: 400 }
      )
    }

    const existingUser = await prisma.user.findUnique({
      where: { email },
    })

    if (existingUser) {
      return NextResponse.json(
        { error: 'Email sudah terdaftar' },
        { status: 409 }
      )
    }

    const hashedPassword = await hashPassword(password)

    const user = await prisma.user.create({
      data: {
        namaLengkap,
        email,
        password: hashedPassword,
        role: 'USER',
        accountType: 'free',
        isDemo: false,
      },
    })

    const token = await signToken({
      userId: user.id,
      email: user.email,
      role: user.role,
    })

    await setAuthCookie(token)

    return NextResponse.json({
      user: {
        id: user.id,
        namaLengkap: user.namaLengkap,
        email: user.email,
        role: user.role,
        accountType: user.accountType,
        isDemo: user.isDemo,
      },
    }, { status: 201 })
  } catch (error) {
    console.error('Register error:', error)
    return NextResponse.json(
      { error: 'Terjadi kesalahan server' },
      { status: 500 }
    )
  }
}
