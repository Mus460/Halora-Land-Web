# Phase 3: Supabase Auth Migration Plan

## Current State (Custom JWT Auth)
- Using jose library for JWT signing/verification
- Password hashing with bcryptjs
- Manual cookie management (auth-token)
- Session stored in HTTP-only cookie (7 days)
- Manual password reset flow (needs implementation)
- No email verification
- No OAuth providers

## Target State (Supabase Auth)
- Supabase Auth handles JWT signing/verification
- Built-in email verification
- Built-in password reset flow
- OAuth providers (Google, GitHub, etc.) - optional
- Secure session management via Supabase SDK
- RLS (Row Level Security) policies - optional

## Migration Strategy

### Option A: Full Migration (Recommended)
Replace custom JWT entirely with Supabase Auth.

**Pros:**
- Get email verification for free
- Get password reset for free
- OAuth ready
- Less code to maintain
- Security best practices built-in

**Cons:**
- Requires user password reset (can't migrate hashed passwords)
- Breaking change for existing sessions
- Moderate effort

**Steps:**
1. Keep existing User table schema
2. Add Supabase Auth integration
3. Update login/register to use Supabase Auth
4. Link Supabase auth.users.id to User.id (via trigger or manual)
5. Update session middleware to use Supabase session
6. Deprecate custom JWT (keep for 1 week grace period)
7. Force logout all users, send password reset emails

### Option B: Hybrid (Backward Compatible)
Keep custom JWT, add Supabase Auth alongside.

**Pros:**
- No breaking changes
- Gradual migration
- Users can keep existing sessions

**Cons:**
- Two auth systems to maintain
- More complexity
- Delays full benefits

**Steps:**
1. Add Supabase Auth for new registrations
2. Keep custom JWT for existing users
3. Offer migration path (re-login with password reset)
4. After 90% migrated, deprecate custom JWT

## Recommendation: **Option A (Full Migration)**

### Implementation Plan (5-7 tasks)

**Task 3.1:** Create Supabase Auth helper functions
- `src/lib/supabase-auth.ts` with login/register/logout helpers
- Server-side session validation

**Task 3.2:** Update register endpoint
- Use Supabase `signUp()` instead of manual user creation
- Link Supabase auth UID to User table

**Task 3.3:** Update login endpoint
- Use Supabase `signInWithPassword()` instead of JWT signing
- Set Supabase session cookie

**Task 3.4:** Update session middleware
- Replace `getCurrentUser()` to use Supabase session
- Update all API routes that call `getCurrentUser()`

**Task 3.5:** Add email verification flow
- Enable email confirmation in Supabase dashboard
- Create email confirmation page

**Task 3.6:** Add password reset flow
- Create password reset request page
- Create password reset confirmation page

**Task 3.7:** Deprecate custom JWT
- Remove `src/lib/auth.ts` (jwt signing)
- Remove `src/lib/session.ts` (cookie management)
- Keep `hashPassword` util if needed for other purposes

## Database Schema Impact

**User table:** No changes needed! Supabase Auth uses separate `auth.users` table.

**Link tables:** Add `supabaseAuthId` field to User table:
```sql
ALTER TABLE "User" ADD COLUMN "supabaseAuthId" TEXT UNIQUE;
```

**Trigger to auto-create User row when Supabase user signs up:**
```sql
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public."User" (id, email, "supabaseAuthId", "namaLengkap", role, "accountType")
  VALUES (
    gen_random_uuid()::int, -- or use sequence
    NEW.email,
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'namaLengkap', NEW.email),
    'USER',
    'FREE'
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
```

## Rollback Plan

If migration fails:
1. Revert code to use custom JWT
2. Keep Supabase Auth disabled
3. No data loss (User table unchanged)

## Timeline Estimate

- Task 3.1-3.4: ~2 hours (core migration)
- Task 3.5-3.6: ~1 hour (email features)
- Task 3.7: ~30 min (cleanup)
- Testing: ~1 hour

**Total: ~4.5 hours**

## Next Steps

1. Confirm this plan
2. Start with Task 3.1 (create Supabase Auth helpers)
