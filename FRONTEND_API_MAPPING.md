# Frontend-API Mapping Status

Last updated: 2026-07-25

## Already Using Real API ✅ (2/31)

1. `/proyek` → `GET /api/proyek`, `POST /api/proyek`, `PUT /api/proyek/:id`, `DELETE /api/proyek/:id`
2. `/proyek/[id]` → `GET /api/proyek/:id`, `PUT /api/proyek/:id`

## Using Mock Data - API Exists 🟡 (6/31)

3. `/master-harga` → `GET /api/master-harga`, `POST /api/master-harga`, `PUT /api/master-harga/:id`, `DELETE /api/master-harga/:id`
4. `/master-analisa` → `GET /api/master-analisa`, `POST /api/master-analisa`, `GET /api/master-analisa/:id/rincian`
5. `/invoice` → `GET /api/proyek/:id/invoice`, `POST /api/proyek/:id/invoice`
6. `/logistik` → `GET /api/proyek/:id/logistik`, `POST /api/proyek/:id/logistik`
7. `/realisasi` → `GET /api/proyek/:id/realisasi`, `POST /api/proyek/:id/realisasi`
8. `/rekap` → `GET /api/proyek/:id/rekap`

## Using Mock Data - API Exists (Pekerjaan Generic) 🟡 (15/31)

All these pages are pekerjaan types and share the same API pattern:
- `GET /api/pekerjaan?proyekId=X&tipe=Y`
- `POST /api/pekerjaan`
- `PUT /api/pekerjaan/:id`
- `DELETE /api/pekerjaan/:id`
- `GET /api/pekerjaan/:id/analisa`

9. `/persiapan` (tipe: PERSIAPAN)
10. `/pondasi` (tipe: PONDASI)
11. `/beton` (tipe: BETON)
12. `/kanopi` (tipe: KANOPI)
13. `/baja` (tipe: BAJA)
14. `/tangga` (tipe: TANGGA)
15. `/atap` (tipe: ATAP)
16. `/dinding` (tipe: DINDING)
17. `/plesteran` (tipe: PLESTERAN)
18. `/acian` (tipe: ACIAN)
19. `/keramik` (tipe: KERAMIK)
20. `/paving` (tipe: PAVING)
21. `/pengecatan` (tipe: PENGECATAN)
22. `/pintu` (tipe: PINTU_JENDELA)
23. `/interior` (tipe: INTERIOR)
24. `/toilet` (tipe: TOILET)
25. `/mep` (tipe: MEP)
26. `/pekerjaan-custom` (tipe: CUSTOM)

## Using Mock Data - No API Yet ❌ (5/31)

27. `/dashboard` - needs: GET /api/dashboard/stats or similar
28. `/kurva-s` - needs: GET /api/proyek/:id/kurva-s
29. `/monitoring` - needs: GET /api/monitoring or /api/proyek/:id/monitoring
30. `/feedback` - needs: GET /api/feedback, POST /api/feedback
31. `/profile` - can use: GET /api/auth/me (exists), PUT /api/auth/me (needs creation)

## Auth Pages (2/2)

- `/login` → `POST /api/auth/login` ✅
- `/register` → `POST /api/auth/register` ✅

## Summary

- **Total pages**: 31 main pages + 2 auth pages
- **Using real API**: 2/31 (6.5%)
- **Has API, needs wiring**: 21/31 (67.7%)
- **Missing API**: 5/31 (16.1%)
- **Auth pages**: 2/2 (100%)

## Priority Order for Integration

### Phase A: Quick Wins (API exists, just wire it)
1. master-harga (GET/POST/PUT/DELETE)
2. master-analisa (GET/POST/GET rincian)
3. All 15 pekerjaan pages (shared pattern)
4. invoice, logistik, realisasi, rekap (proyekId-scoped)

### Phase B: Missing APIs (need backend work)
1. dashboard stats
2. kurva-s
3. monitoring
4. feedback
5. profile update (auth/me PUT)
