// DEPRECATED: This file is replaced by supabase-auth.ts
// Keeping for backward compatibility during migration
// TODO: Remove after all references are migrated

import { getCurrentSupabaseUser } from './supabase-auth'

/**
 * @deprecated Use getCurrentSupabaseUser() from supabase-auth.ts instead
 */
export async function getCurrentUser() {
  return getCurrentSupabaseUser()
}
