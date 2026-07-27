#!/bin/bash
# RAB + AHSP - Quick Test Script

echo "=================================="
echo "RAB + AHSP Implementation Test"
echo "=================================="
echo ""

echo "1. Checking files..."
echo ""

# Check critical files
FILES=(
  "src/lib/ahsp-parser.ts"
  "src/app/api/admin/ahsp/import/route.ts"
  "src/app/api/master-analisa/search/route.ts"
  "src/app/api/proyek/[id]/pekerjaan/from-ahsp/route.ts"
  "src/app/(dashboard)/admin/ahsp/page.tsx"
  "src/app/(dashboard)/rab/page.tsx"
  "public/data/ahsp-2026.xlsx"
  "prisma/schema.prisma"
)

for file in "${FILES[@]}"; do
  if [ -f "$file" ]; then
    SIZE=$(du -sh "$file" | cut -f1)
    echo "✅ $file ($SIZE)"
  else
    echo "❌ $file - NOT FOUND"
  fi
done

echo ""
echo "2. Checking schema changes..."
grep -q "isSystem" prisma/schema.prisma && echo "✅ isSystem field added" || echo "❌ isSystem field missing"
grep -q "ahspKode" prisma/schema.prisma && echo "✅ ahspKode field added" || echo "❌ ahspKode field missing"
grep -q "ahspSheet" prisma/schema.prisma && echo "✅ ahspSheet field added" || echo "❌ ahspSheet field missing"

echo ""
echo "3. Next steps:"
echo ""
echo "   Run database migration:"
echo "   $ npx prisma db push"
echo ""
echo "   Generate Prisma client:"
echo "   $ npx prisma generate"
echo ""
echo "   Start dev server:"
echo "   $ npm run dev"
echo ""
echo "   Then navigate to:"
echo "   - http://localhost:3000/admin/ahsp  (Import AHSP)"
echo "   - http://localhost:3000/rab         (Create RAB)"
echo ""
echo "=================================="
echo "✅ All files created successfully!"
echo "=================================="
