import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'
import { getCurrentUser } from '@/lib/session'

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; invoiceId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { invoiceId } = await params
    const id = parseInt(invoiceId)

    const invoice = await prisma.invoice.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!invoice) {
      return NextResponse.json({ error: 'Invoice tidak ditemukan' }, { status: 404 })
    }

    const hasAccess = 
      session.role === 'ADMIN' ||
      invoice.proyek.userId === session.userId ||
      invoice.proyek.timProyek.some(t => t.userId === session.userId)

    if (!hasAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    return NextResponse.json({ invoice })
  } catch (error) {
    console.error('Get invoice error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; invoiceId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { invoiceId } = await params
    const id = parseInt(invoiceId)

    const invoice = await prisma.invoice.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!invoice) {
      return NextResponse.json({ error: 'Invoice tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      invoice.proyek.userId === session.userId ||
      invoice.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    const body = await request.json()
    const { tanggal, total, status } = body

    const updated = await prisma.invoice.update({
      where: { id },
      data: {
        ...(tanggal && { tanggal: new Date(tanggal) }),
        ...(total && { total }),
        ...(status && { status }),
      }
    })

    return NextResponse.json({ invoice: updated })
  } catch (error) {
    console.error('Update invoice error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string; invoiceId: string }> }
) {
  try {
    const session = await getCurrentUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { invoiceId } = await params
    const id = parseInt(invoiceId)

    const invoice = await prisma.invoice.findUnique({
      where: { id },
      include: {
        proyek: {
          include: { timProyek: true }
        }
      }
    })

    if (!invoice) {
      return NextResponse.json({ error: 'Invoice tidak ditemukan' }, { status: 404 })
    }

    const hasEditAccess = 
      session.role === 'ADMIN' ||
      invoice.proyek.userId === session.userId ||
      invoice.proyek.timProyek.some(t => t.userId === session.userId && t.role === 'editor')

    if (!hasEditAccess) {
      return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
    }

    await prisma.invoice.delete({ where: { id } })

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Delete invoice error:', error)
    return NextResponse.json({ error: 'Terjadi kesalahan server' }, { status: 500 })
  }
}
