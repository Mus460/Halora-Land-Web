# Halora Land Web — Aplikasi RAB Konstruksi

Aplikasi web untuk menghitung Rencana Anggaran Biaya (RAB) konstruksi dengan metode AHSP (Analisa Harga Satuan Pekerjaan) standar PUPR atau manual.

## 🎯 Fitur Utama

- **Mode Gedung & Infrastruktur** — Switch mode sesuai jenis proyek
- **Perhitungan RAB AHSP** — Menggunakan standar PUPR 182/2026 (Bina Konstruksi)
- **Visualisasi 3D Atap** — Three.js rendering untuk preview struktur atap
- **Grafik Kurva S** — Chart.js untuk tracking progress keuangan proyek
- **Manajemen Proyek** — Tim, logistik, invoice, realisasi keuangan
- **Role Management** — Admin, Owner, User, Demo mode

## 🛠️ Tech Stack

- **Framework:** Next.js 16.2 (App Router)
- **Frontend:** React 19 + TypeScript
- **Database:** PostgreSQL + Prisma ORM 7.9
- **Styling:** Tailwind CSS v4 + shadcn/ui
- **3D Rendering:** Three.js + React Three Fiber
- **Charts:** Chart.js
- **Auth:** jose (JWT)
- **Forms:** React Hook Form + Zod validation
- **State:** Zustand

## 📦 Installation

```bash
# Clone repository
git clone https://github.com/Mus460/Halora-Land-Web.git
cd Halora-Land-Web

# Install dependencies
npm install

# Setup environment
cp .env.example .env
# Edit .env dengan database credentials kamu

# Setup database
npm run db:generate   # Generate Prisma Client
npm run db:push       # Push schema ke database
npm run db:seed       # Seed initial data

# Run development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) di browser.

## 📂 Project Structure

```
├── src/
│   ├── app/           # Next.js App Router (pages & layouts)
│   ├── components/    # React components
│   ├── lib/          # Utilities, configs, helpers
│   └── types/        # TypeScript type definitions
├── prisma/
│   ├── schema.prisma  # Database schema
│   └── seed.ts       # Seed data
├── public/           # Static assets (images, fonts)
└── README.md
```

## 🔐 Environment Variables

Create `.env` file di root:

```env
# Database
DATABASE_URL="postgresql://user:password@localhost:5432/halora_land"

# Auth
JWT_SECRET="your-secret-key-here"

# App
NODE_ENV="development"
```

**Security Note:** Never commit `.env` file!

## 🚀 Available Scripts

```bash
npm run dev          # Start development server
npm run build        # Build production bundle
npm run start        # Start production server
npm run db:generate  # Generate Prisma Client
npm run db:push      # Push schema to database
npm run db:migrate   # Create migration files
npm run db:seed      # Seed database
npm run db:studio    # Open Prisma Studio (DB GUI)
```

## 🏗️ Key Features

### Mode Gedung
- Pekerjaan Konstruksi (Struktur, Pondasi, dll)
- Arsitektur & MEP (Mekanikal, Elektrikal, Plumbing)

### Mode Infrastruktur
- Divisi 1–10 (Pekerjaan Umum, Drainase, Perkerasan, dll)

### Admin Center
- User management
- Import AHSP database
- Feedback handling

### Project Management
- Multi-project support
- Team collaboration
- Budget tracking
- Invoice generation
- Financial realization reports

## 📊 Database

PostgreSQL + Prisma ORM. Run `npm run db:studio` untuk GUI.

## 🔒 Security

- JWT-based authentication
- bcrypt password hashing
- Environment variables untuk secrets
- Role-based access control (RBAC)

## 🐛 Troubleshooting

### Database Connection Error
```bash
# Check PostgreSQL
sudo systemctl status postgresql

# Verify DATABASE_URL format
postgresql://USER:PASSWORD@HOST:PORT/DATABASE
```

### Build Errors
```bash
rm -rf .next node_modules
npm install
npm run build
```

## 📝 License

Private Project — All Rights Reserved

---

**Status:** Development phase — feature set subject to change.
