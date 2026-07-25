   # HitungBangun V3 - Deployment Guide

## Pre-Deployment Checklist

### 1. Environment Variables
Copy dari `.env.local` ke Vercel dashboard:

```bash
# Database (Supabase)
DATABASE_URL="postgresql://postgres.yqswxbslojancdasipwe:Hapebaru_16@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres?pgbouncer=true"
DIRECT_URL="postgresql://postgres.yqswxbslojancdasipwe:Hapebaru_16@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres"

# Supabase Auth
NEXT_PUBLIC_SUPABASE_URL="https://yqswxbslojancdasipwe.supabase.co"
NEXT_PUBLIC_SUPABASE_ANON_KEY="<your-anon-key>"

# JWT (legacy, still used for some validation)
JWT_SECRET="<your-jwt-secret>"

# Sentry (optional - get from sentry.io)
NEXT_PUBLIC_SENTRY_DSN="https://xxx@xxx.ingest.sentry.io/xxx"

# PostHog (optional - get from posthog.com)
NEXT_PUBLIC_POSTHOG_KEY="phc_xxx"
NEXT_PUBLIC_POSTHOG_HOST="https://app.posthog.com"

# App Config
NODE_ENV="production"
NEXT_PUBLIC_APP_NAME="HitungBangun V3"
NEXT_PUBLIC_APP_VERSION="3.0.0"
```

### 2. Build Check

```bash
npm run build
```

**Expected:** Build succeeds without errors.

### 3. Type Check

```bash
npm run type-check
```

**Expected:** No type errors.

### 4. Lint Check

```bash
npm run lint
```

**Expected:** No critical lint errors.

---

## Deployment Steps

### Option A: Vercel CLI (Recommended)

1. Install Vercel CLI:
```bash
npm i -g vercel
```

2. Login:
```bash
vercel login
```

3. Deploy preview:
```bash
vercel
```

4. Deploy production:
```bash
vercel --prod
```

### Option B: Vercel Dashboard

1. Push to GitHub
2. Import repository at vercel.com
3. Add environment variables
4. Deploy

---

## Post-Deployment Verification

### 1. Smoke Test Checklist

- [ ] Landing page loads
- [ ] Register new account
- [ ] Verify email (check Supabase Auth dashboard)
- [ ] Login works
- [ ] Dashboard displays stats
- [ ] Create new proyek
- [ ] Add master harga
- [ ] Add pekerjaan (test 1 kategori)
- [ ] Check kurva-s page
- [ ] Check monitoring page
- [ ] Submit feedback
- [ ] Logout works

### 2. Performance Check

- [ ] Lighthouse score > 90
- [ ] First Contentful Paint < 2s
- [ ] Time to Interactive < 3s
- [ ] No console errors

### 3. Security Check

- [ ] HTTPS enforced
- [ ] Security headers present (check DevTools Network tab)
- [ ] Rate limiting works (try 11 login attempts)
- [ ] CORS configured properly

### 4. Monitoring Check

- [ ] Sentry receiving errors (test by triggering an error)
- [ ] PostHog tracking pageviews
- [ ] Vercel Analytics active

---

## Rollback Plan

If deployment fails:

1. **Via Vercel Dashboard:**
   - Go to Deployments tab
   - Click "..." on previous working deployment
   - Click "Promote to Production"

2. **Via CLI:**
```bash
vercel rollback
```

---

## Common Issues & Fixes

### Issue: "Database does not exist"
**Fix:** Database connection from local may be blocked. Deploy to Vercel first (Vercel can connect to Supabase).

### Issue: "Module not found"
**Fix:** 
```bash
rm -rf node_modules .next
npm install
npm run build
```

### Issue: "Prisma client not generated"
**Fix:** Prisma generates on `npm install` via postinstall. If not:
```bash
npx prisma generate
```

### Issue: Rate limit headers not showing
**Fix:** Headers only work in production/preview. Test on deployed URL, not localhost.

### Issue: Sentry not receiving events
**Fix:** Check `NEXT_PUBLIC_SENTRY_DSN` is set in Vercel env vars.

---

## Monitoring URLs

After deployment, bookmark these:

- **App:** https://your-app.vercel.app
- **Vercel Dashboard:** https://vercel.com/dashboard
- **Supabase Dashboard:** https://supabase.com/dashboard/project/yqswxbslojancdasipwe
- **Sentry:** https://sentry.io/organizations/your-org/issues/
- **PostHog:** https://app.posthog.com

---

## Next Steps (Post-Launch)

1. Monitor error rate in Sentry for 24h
2. Check user feedback in `/feedback` page
3. Review PostHog analytics for user behavior
4. Scale database if needed (Supabase dashboard → Database → Upgrade plan)
5. Add custom domain (Vercel dashboard → Domains)

---

## Success Criteria ✅

- [ ] All 31 pages load from real API
- [ ] All CRUD operations work
- [ ] Auth complete (register → verify → login → logout)
- [ ] Calculations accurate
- [ ] Rate limiting active (test with 11 login attempts)
- [ ] Security headers present (check Network tab)
- [ ] Sentry tracking errors
- [ ] PostHog tracking pageviews
- [ ] Vercel Analytics active
- [ ] No console errors on production
- [ ] Lighthouse score > 90
