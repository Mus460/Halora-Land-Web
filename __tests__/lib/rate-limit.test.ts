import { describe, it, expect, vi, beforeEach } from 'vitest'
import { rateLimit, getClientIp } from '@/lib/rate-limit'

describe('rateLimit', () => {
  beforeEach(() => {
    // Clear rate limit store between tests
    vi.useFakeTimers()
  })

  it('should allow requests within limit', () => {
    const config = {
      id: 'test-user',
      limit: 5,
      windowSec: 60,
    }

    // First 5 requests should succeed
    for (let i = 0; i < 5; i++) {
      const result = rateLimit(config)
      expect(result.success).toBe(true)
      expect(result.remaining).toBe(4 - i)
    }
  })

  it('should block requests exceeding limit', () => {
    const config = {
      id: 'test-user-2',
      limit: 3,
      windowSec: 60,
    }

    // First 3 requests succeed
    for (let i = 0; i < 3; i++) {
      const result = rateLimit(config)
      expect(result.success).toBe(true)
    }

    // 4th request should fail
    const blocked = rateLimit(config)
    expect(blocked.success).toBe(false)
    expect(blocked.remaining).toBe(0)
  })

  it('should reset after window expires', () => {
    const config = {
      id: 'test-user-3',
      limit: 2,
      windowSec: 60,
    }

    // Use up the limit
    rateLimit(config)
    rateLimit(config)
    
    const blocked = rateLimit(config)
    expect(blocked.success).toBe(false)

    // Advance time past window
    vi.advanceTimersByTime(61 * 1000)

    // Should allow requests again
    const afterReset = rateLimit(config)
    expect(afterReset.success).toBe(true)
    expect(afterReset.remaining).toBe(1)
  })

  it('should track different IDs independently', () => {
    const config1 = {
      id: 'user-a',
      limit: 2,
      windowSec: 60,
    }

    const config2 = {
      id: 'user-b',
      limit: 2,
      windowSec: 60,
    }

    // User A uses up limit
    rateLimit(config1)
    rateLimit(config1)
    expect(rateLimit(config1).success).toBe(false)

    // User B should still have limit
    expect(rateLimit(config2).success).toBe(true)
    expect(rateLimit(config2).success).toBe(true)
  })

  it('should return correct reset timestamp', () => {
    const config = {
      id: 'test-reset',
      limit: 5,
      windowSec: 60,
    }

    const now = Date.now()
    const result = rateLimit(config)
    
    expect(result.reset).toBeGreaterThan(now)
    expect(result.reset).toBeLessThanOrEqual(now + 60 * 1000)
  })
})

describe('getClientIp', () => {
  it('should extract IP from x-forwarded-for header', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') {
            return '203.0.113.1, 192.0.2.1'
          }
          return null
        }
      }
    } as any

    const ip = getClientIp(mockRequest)
    expect(ip).toBe('203.0.113.1')
  })

  it('should handle single IP in x-forwarded-for', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') {
            return '203.0.113.1'
          }
          return null
        }
      }
    } as any

    const ip = getClientIp(mockRequest)
    expect(ip).toBe('203.0.113.1')
  })

  it('should return dev-client when no forwarded header', () => {
    const mockRequest = {
      headers: {
        get: () => null
      }
    } as any

    const ip = getClientIp(mockRequest)
    expect(ip).toBe('dev-client')
  })

  it('should trim whitespace from IP', () => {
    const mockRequest = {
      headers: {
        get: (name: string) => {
          if (name === 'x-forwarded-for') {
            return '  203.0.113.1  , 192.0.2.1'
          }
          return null
        }
      }
    } as any

    const ip = getClientIp(mockRequest)
    expect(ip).toBe('203.0.113.1')
  })
})
