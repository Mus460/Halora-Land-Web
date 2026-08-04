-- Local-only auth (Supabase removed): drop the Supabase link column.
-- Password hashes now live in users."passwordHash" for every account.
ALTER TABLE "users" DROP COLUMN IF EXISTS "supabaseAuthId";
DROP INDEX IF EXISTS "users_supabaseAuthId_key";
