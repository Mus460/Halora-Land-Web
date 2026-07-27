import { NextResponse } from 'next/server'

export class ApiError extends Error {
  constructor(public message: string, public status: number = 400) {
    super(message)
  }
}

export async function getJsonBody<T>(request: Request): Promise<T> {
  const contentType = request.headers.get('content-type')
  if (contentType !== 'application/json') {
    throw new ApiError('Content-Type must be application/json', 400)
  }
  try {
    return await request.json() as T
  } catch {
    throw new ApiError('Invalid JSON body', 400)
  }
}

export function parseId(raw: string, name = 'id'): number {
  const id = parseInt(raw, 10)
  if (isNaN(id) || id <= 0) {
    throw new ApiError(`Invalid ${name}`, 400)
  }
  return id
}

export function parseNumber(raw: string | number, name: string, opts: {
  min?: number
  max?: number
}) {
  const num = typeof raw === 'string' ? parseFloat(raw) : raw
  if (isNaN(num)) throw new ApiError(`Invalid ${name}`, 400)
  if (opts.min !== undefined && num < opts.min) throw new ApiError(`${name} must be >= ${opts.min}`, 400)
  if (opts.max !== undefined && num > opts.max) throw new ApiError(`${name} must be <= ${opts.max}`, 400)
  return num
}

export function handleError(error: unknown) {
  console.error('[API Error]', error)
  
  if (error instanceof ApiError) {
    return NextResponse.json(
      { error: error.message },
      { status: error.status }
    )
  }
  
  if (typeof error === 'object' && error !== null && 'code' in error) {
    const prismaError = error as { code: string; meta?: any }
    
    if (prismaError.code === 'P2002') {
      return NextResponse.json(
        { error: 'Data sudah ada (duplicate)' },
        { status: 409 }
      )
    }
    if (prismaError.code === 'P2025') {
      return NextResponse.json(
        { error: 'Data tidak ditemukan' },
        { status: 404 }
      )
    }
  }
  
  return NextResponse.json(
    { error: 'Terjadi kesalahan server' },
    { status: 500 }
  )
}
