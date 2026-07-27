# 🚀 Quick Start Guide - RAB + AHSP

## Prerequisites

✅ PostgreSQL database configured
✅ Next.js 16.2.11 running
✅ Node.js installed

---

## Step 1: Apply Database Migration

```bash
cd "/home/musyaffa/Dokumen/Bangunin clone/hitungbangun"

# Option A: Push schema directly (development)
npx prisma db push

# Option B: Create migration (production)
npx prisma migrate dev --name add_ahsp_support

# Generate Prisma client
npx prisma generate
```

**Expected output:**
```
✓ Database schema updated
✓ Prisma Client generated
```

---

## Step 2: Install Dependencies (if needed)

```bash
npm install xlsx
```

---

## Step 3: Start Development Server

```bash
npm run dev
```

Navigate to: http://localhost:3000

---

## Step 4: Import AHSP Data (Popular Categories)

### Login as ADMIN
1. Open browser → http://localhost:3000/login
2. Login with ADMIN account

### Import Categories
1. Navigate to: http://localhost:3000/admin/ahsp
2. You should see 41 categories listed
3. Click "Import Now" on these popular ones:
   - **Beton** (~55 items) ← START HERE
   - **Persiapan** (~35 items)
   - **Penutup Atap** (~45 items)
   - **Pasangan Dinding** (~27 items)

**Each import takes ~5-10 seconds**

Or click **"Import Semua"** to import all 41 categories at once (takes ~5 minutes)

---

## Step 5: Create Your First RAB

### 1. Select Project
- Click project selector in header
- Select an existing project

### 2. Navigate to RAB Page
- Sidebar → "RAB" (under "Utama" section)
- Or direct: http://localhost:3000/rab

### 3. Search for Items
Type in search box:
```
beton k-300
```

You should see results like:
- **2.2.2.4.1** - 1 m3 beton K-300 ready mix (Rp 1,205,400 / m3)
- **2.2.2.4.2** - 1 m3 beton K-300 site mix (Rp 985,200 / m3)

### 4. Add Item
1. Click **"Tambah"** on desired item
2. Enter volume: `15.5` (for 15.5 m3)
3. Click OK
4. Item added to RAB with auto-calculated breakdown!

### 5. View RAB Summary
Scroll down to see:
- Items grouped by kategori
- Subtotal per kategori
- Total calculation:
  - Subtotal Pekerjaan
  - Overhead (10%)
  - Profit (10%)
  - PPN (11%)
  - **TOTAL RAB**

---

## Troubleshooting

### "Module not found: xlsx"
```bash
npm install xlsx
```

### "Cannot connect to database"
Check `.env.local`:
```
DATABASE_URL="postgresql://..."
```

### "Unauthorized" when accessing /admin/ahsp
Make sure your user has `role: 'ADMIN'` in database:
```sql
UPDATE users SET role = 'ADMIN' WHERE email = 'your@email.com';
```

### Search returns no results
1. Make sure you've imported the category first (via /admin/ahsp)
2. Check search query has 3+ characters
3. Try broader keywords: "beton", "atap", "dinding"

### Import fails with "file not found"
Verify Excel file exists:
```bash
ls -lh public/data/ahsp-2026.xlsx
```
Should show: `2.1M` file size

---

## File Structure

```
src/
├── app/
│   ├── api/
│   │   ├── admin/ahsp/import/route.ts          ← Import API
│   │   ├── master-analisa/search/route.ts      ← Search API
│   │   └── proyek/[id]/pekerjaan/
│   │       └── from-ahsp/route.ts              ← Add item API
│   └── (dashboard)/
│       ├── admin/ahsp/page.tsx                 ← Admin import UI
│       └── rab/page.tsx                        ← RAB central page
├── lib/
│   └── ahsp-parser.ts                          ← Excel parser
└── components/
    └── layout/
        ├── sidebar.tsx                         ← Updated nav
        └── mobile-nav.tsx                      ← Updated nav

public/
└── data/
    └── ahsp-2026.xlsx                          ← AHSP Excel data

prisma/
├── schema.prisma                               ← Updated schema
└── migrations/
    └── 20260725_add_ahsp_support.sql          ← Migration SQL
```

---

## API Endpoints Reference

### 1. Get Import Status
```bash
GET /api/admin/ahsp/import
Authorization: Required (ADMIN)

Response:
{
  "status": [
    {
      "sheetName": "Beton",
      "kategori": "beton",
      "imported": true,
      "count": 55
    },
    ...
  ]
}
```

### 2. Import Sheet
```bash
POST /api/admin/ahsp/import
Content-Type: application/json
Authorization: Required (ADMIN)

Body:
{
  "sheetName": "Beton",
  "forceReimport": false
}

Response:
{
  "success": true,
  "sheetName": "Beton",
  "kategori": "beton",
  "imported": 55,
  "total": 55
}
```

### 3. Search AHSP
```bash
GET /api/master-analisa/search?q=beton+k300&limit=10
Authorization: Required

Response:
{
  "results": [
    {
      "id": 123,
      "kode": "2.2.2.4.1",
      "nama": "1 m3 beton K-300 ready mix",
      "satuan": "m3",
      "hargaSatuan": 1205400,
      "kategori": "beton",
      "ahspKode": "2.2.2.4.1"
    }
  ],
  "total": 1,
  "query": "beton k300"
}
```

### 4. Add Item to RAB
```bash
POST /api/proyek/1/pekerjaan/from-ahsp
Content-Type: application/json
Authorization: Required

Body:
{
  "masterAnalisaId": 123,
  "volume": 15.5,
  "applyBreakdown": true
}

Response:
{
  "success": true,
  "pekerjaan": {
    "id": 456,
    "uraianPekerjaan": "1 m3 beton K-300 ready mix",
    "volume": 15.5,
    "hargaSatuan": 1205400,
    "totalBiaya": 18683700,
    "detailAnalisa": [...]
  }
}
```

---

## Next Steps After Setup

1. **Test the flow**:
   - Import → Search → Add → Calculate
   
2. **Import more categories**:
   - Go to /admin/ahsp
   - Import categories as needed
   
3. **Customize**:
   - Edit margin/overhead in /rekap
   - Add more projects
   - Generate invoices

4. **Production deployment**:
   - Run `npm run build`
   - Apply migration: `npx prisma migrate deploy`
   - Import AHSP via admin UI

---

## Support

Issues? Check:
1. `RAB_IMPLEMENTATION.md` - Full technical documentation
2. Console logs in browser DevTools
3. Server logs in terminal

---

Generated: 2026-07-25
