import { getCurrentSupabaseUser } from './supabase-auth'

export async function getCurrentUser() {
  return await getCurrentSupabaseUser()
}
