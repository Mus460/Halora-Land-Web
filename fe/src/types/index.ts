export type Role = "ADMIN" | "OWNER" | "USER" | "DEMO";

export interface User {
  id: number;
  namaLengkap: string;
  email: string;
  role: Role;
  accountType: string;
  isDemo: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Proyek {
  id: number;
  userId: number;
  namaProyek: string;
  lokasi: string | null;
  tipe: "gedung" | "infra";
  isPitching: boolean;
  nilaiKontrak: number | null;
  timeline: string | null;
  createdAt: string;
  updatedAt: string;
  user?: User;
  timProyek?: TimProyek[];
  pekerjaan?: Pekerjaan[];
  _count?: {
    pekerjaan: number;
  };
}

export interface TimProyek {
  id: number;
  proyekId: number;
  userId: number;
  role: "owner" | "editor" | "viewer";
  createdAt: string;
  user?: User;
  proyek?: Proyek;
}

export type KategoriPekerjaan =
  | "persiapan"
  | "pondasi"
  | "beton"
  | "kanopi"
  | "baja"
  | "tangga"
  | "atap"
  | "dinding"
  | "plesteran"
  | "acian"
  | "keramik"
  | "paving"
  | "pengecatan"
  | "pintu"
  | "interior"
  | "toilet"
  | "mep"
  | "custom";

export type MetodeHitung =
  | "ahsp"
  | "manual"
  | "harga_borong"
  | "harga_manual"
  | "harga_custom";

export type TipeKomponen = "material" | "upah" | "alat";

export interface Pekerjaan {
  id: number;
  proyekId: number;
  kategori: KategoriPekerjaan;
  uraianPekerjaan: string;
  volume: number;
  satuan: string;
  hargaSatuan: number;
  totalBiaya: number;
  metodeHitung: MetodeHitung;
  levelPekerjaan: string | null;
  tipePekerjaan: string | null;
  masterAnalisaId: number | null;
  waktu: number | null;
  totalWaktu: number | null;
  createdAt: string;
  updatedAt: string;
  detailAnalisa?: DetailAnalisa[];
}

export interface DetailAnalisa {
  id: number;
  pekerjaanId: number;
  masterHargaId: number | null;
  nama: string;
  satuan: string;
  koef: number;
  hargaSatuan: number;
  totalBiaya: number;
  tipe: TipeKomponen;
}

export interface MasterAnalisa {
  id: number;
  kode: string;
  nama: string;
  level: number;
  parentId: number | null;
  satuan: string | null;
  hargaSatuan?: number | null;
  isGlobal: boolean;
  userId: number | null;
  createdAt: string;
  children?: MasterAnalisa[];
  rincianAnalisa?: RincianAnalisa[];
}

export interface RincianAnalisa {
  id: number;
  masterAnalisaId: number;
  komponenId: number | null;
  koef: number;
  tipe: TipeKomponen;
  nama: string | null;
  satuan: string | null;
  hargaSatuan: number | null;
  jumlahHarga: number | null;
  kodeReferensi: string | null;
  waktu: number | null;
  urutan: number;
  komponen?: MasterHarga;
}

export interface MasterHarga {
  id: number;
  nama: string;
  satuan: string;
  harga: number;
  kategori: TipeKomponen;
  isGlobal: boolean;
  userId: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface Rekap {
  id: number;
  proyekId: number;
  kategori: string;
  uraian: string;
  urutan: number;
  margin: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface Invoice {
  id: number;
  proyekId: number;
  nomor: string;
  tanggal: string;
  total: number;
  status: "draft" | "sent" | "paid";
  createdAt: string;
}

export interface Logistik {
  id: number;
  proyekId: number;
  namaMaterial: string;
  satuan: string;
  volume: number;
  hargaSatuan: number;
  totalBiaya: number;
  tanggal: string | null;
  keterangan: string | null;
  createdAt: string;
}

export interface Realisasi {
  id: number;
  proyekId: number;
  tanggal: string;
  kategori: string;
  jumlah: number;
  keterangan: string | null;
  jenis: "pengeluaran" | "pemasukan";
  status: "draft" | "approved" | "reverted";
  logistikId: number | null;
  invoiceId: number | null;
  createdAt: string;
}

export interface News {
  id: number;
  title: string;
  content: string;
  isActive: boolean;
  createdAt: string;
}

export interface CalculationResult {
  volume: number;
  hargaSatuan: number;
  totalBiaya: number;
  breakdown: {
    material: number;
    upah: number;
    alat: number;
  };
}

export interface RABResult {
  subtotal: number;
  overhead: number;
  profit: number;
  ppn: number;
  total: number;
}
