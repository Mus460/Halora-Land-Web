import { expect } from 'vitest'

// Mock environment variables for tests
process.env.JWT_SECRET = 'test-secret-key-for-unit-tests-min-32-chars'
process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://test.supabase.co'
process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'test-anon-key'
