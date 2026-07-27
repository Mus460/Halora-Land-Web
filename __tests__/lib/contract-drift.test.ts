/**
 * Contract-drift regression tests.
 *
 * These tests assert the CORRECT behaviour. They are expected to FAIL until the
 * corresponding bugs are fixed. Each failure documents a real defect found
 * during the round-2 security/correctness audit — see SECURITY_AUDIT_REPORT.md.
 *
 * Do not "fix" these tests by weakening the assertion. Fix the source.
 */
import { describe, it, expect } from 'vitest'
import { registerSchema } from '@/lib/schemas'
import { getJsonBody } from '@/lib/api-utils'

describe('BUG-01: registerSchema field name does not match register route', () => {
  /**
   * src/app/api/auth/register/route.ts:38 destructures `namaLengkap`:
   *   const { namaLengkap, email, password } = registerSchema.parse(body)
   * but registerSchema (src/lib/schemas.ts:16) defines the field as `name`.
   * Result: namaLengkap === undefined, then prisma.user.create() receives
   * undefined for the required String column `namaLengkap` and throws.
   * Registration is broken for every request.
   */
  it('schema should expose namaLengkap so the route destructure resolves', () => {
    const parsed = registerSchema.parse({
      email: 'user@example.com',
      password: 'Password123',
      namaLengkap: 'John Doe',
    }) as Record<string, unknown>

    expect(parsed.namaLengkap).toBe('John Doe')
  })
})

describe('BUG-02: getJsonBody rejects charset-qualified JSON content types', () => {
  /**
   * src/lib/api-utils.ts:11 uses strict equality:
   *   if (contentType !== 'application/json')
   * Browsers, axios and fetch commonly send
   * "application/json; charset=utf-8", which is a valid JSON content type per
   * RFC 9110. Every POST/PUT route using getJsonBody rejects those requests
   * with 400 "Content-Type must be application/json".
   */
  function jsonRequest(contentType: string) {
    return new Request('http://localhost/api/test', {
      method: 'POST',
      headers: { 'content-type': contentType },
      body: JSON.stringify({ ok: true }),
    })
  }

  it('accepts bare application/json', async () => {
    await expect(getJsonBody(jsonRequest('application/json'))).resolves.toEqual({ ok: true })
  })

  it('should accept application/json; charset=utf-8', async () => {
    await expect(
      getJsonBody(jsonRequest('application/json; charset=utf-8'))
    ).resolves.toEqual({ ok: true })
  })

  it('should accept application/json with uppercase and spacing variance', async () => {
    await expect(
      getJsonBody(jsonRequest('Application/JSON;charset=UTF-8'))
    ).resolves.toEqual({ ok: true })
  })

  it('still rejects genuinely wrong content types', async () => {
    await expect(getJsonBody(jsonRequest('text/plain'))).rejects.toThrow()
  })
})
