import { SignJWT, jwtVerify } from 'jose'
import { hash, compare } from 'bcryptjs'

const secret = process.env.JWT_SECRET
if (!secret) {
  throw new Error('JWT_SECRET environment variable is required')
}
if (secret.length < 32) {
  throw new Error('JWT_SECRET must be at least 32 characters')
}
const JWT_SECRET = new TextEncoder().encode(secret)

export interface JWTPayload {
  userId: number
  email: string
  role: string
  iat?: number
  exp?: number
}

export async function hashPassword(password: string): Promise<string> {
  return hash(password, 12)
}

export async function verifyPassword(password: string, hashedPassword: string): Promise<boolean> {
  return compare(password, hashedPassword)
}

export async function signToken(payload: JWTPayload): Promise<string> {
  return new SignJWT(payload as any)
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('7d')
    .sign(JWT_SECRET)
}

export async function verifyToken(token: string): Promise<JWTPayload | null> {
  try {
    const { payload } = await jwtVerify(token, JWT_SECRET)
    return payload as unknown as JWTPayload
  } catch {
    return null
  }
}
