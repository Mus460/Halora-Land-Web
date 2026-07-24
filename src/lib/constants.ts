import type { KategoriPekerjaan, TipeKomponen } from "@/types";

export const KATEGORI_LIST: {
  value: KategoriPekerjaan;
  label: string;
  icon: string;
  section: string;
}[] = [
  { value: "persiapan", label: "Persiapan", icon: "ruler", section: "konstruksi" },
  { value: "pondasi", label: "Pondasi", icon: "box", section: "konstruksi" },
  { value: "beton", label: "Beton Struktur", icon: "pickaxe", section: "konstruksi" },
  { value: "kanopi", label: "Kanopi", icon: "umbrella", section: "konstruksi" },
  { value: "baja", label: "Baja Struktural", icon: "construction", section: "konstruksi" },
  { value: "tangga", label: "Tangga", icon: "move-up-right", section: "konstruksi" },
  { value: "atap", label: "Pek. Atap", icon: "house", section: "konstruksi" },
  { value: "dinding", label: "Pas. Dinding", icon: "grid-3x3", section: "arsitektur" },
  { value: "plesteran", label: "Plesteran", icon: "layers", section: "arsitektur" },
  { value: "acian", label: "Acian", icon: "droplet", section: "arsitektur" },
  { value: "keramik", label: "Lantai/Keramik", icon: "layout-grid", section: "arsitektur" },
  { value: "paving", label: "Paving & Halaman", icon: "map", section: "arsitektur" },
  { value: "pengecatan", label: "Cat & Plafon", icon: "paintbrush", section: "arsitektur" },
  { value: "pintu", label: "Kusen/Pintu", icon: "door-open", section: "arsitektur" },
  { value: "interior", label: "Interior", icon: "sofa", section: "arsitektur" },
  { value: "toilet", label: "Toilet/Sanitair", icon: "bath", section: "arsitektur" },
  { value: "mep", label: "Instalasi MEP", icon: "zap", section: "arsitektur" },
  { value: "custom", label: "Pekerjaan Custom", icon: "pen-tool", section: "tambahan" },
];

export const SATUAN_OPTIONS = [
  { value: "m2", label: "m²" },
  { value: "m3", label: "m³" },
  { value: "m1", label: "m¹" },
  { value: "m", label: "m" },
  { value: "unit", label: "Unit" },
  { value: "kg", label: "Kg" },
  { value: "ls", label: "LS" },
  { value: "set", label: "Set" },
  { value: "titik", label: "Titik" },
];

export const TIPE_KOMPONEN: { value: TipeKomponen; label: string }[] = [
  { value: "material", label: "Material" },
  { value: "upah", label: "Upah" },
  { value: "alat", label: "Alat" },
];

export const TIPE_BETON = [
  "kolom",
  "balok",
  "plat",
  "sloof",
  "ringbalk",
  "dak",
];

export const LEVEL_PEKERJAAN = [
  "Pekerjaan Utama",
  "Lantai 1",
  "Lantai 2",
  "Lantai 3",
];

export const JENIS_KANOPI = [
  "polycarbonate",
  "alderon",
  "grc",
  "spandek",
];

export const JENIS_DINDING = ["hebel", "bata merah", "batako"];

export const TIPE_PERSIAPAN = [
  "Pembersihan Lokasi",
  "Bouwplank",
  "Pagar Proyek",
  "Bedeng/Gudang",
];

export const TIPE_PONDASI = ["batu_kali", "footplate"];

export const MUTU_BETON = [
  "K-225 Standar",
  "K-250 Menengah",
  "K-300 Tinggi",
];

export const JENIS_KERAMIK = [
  "25x40",
  "25x50",
  "30x30",
  "30x60",
  "40x40",
  "60x60",
  "Granit",
];

export const JENIS_RANGKA_ATAP = ["baja ringan", "baja WF", "kayu"];

export const JENIS_ATAP = [
  "genteng beton",
  "genteng keramik",
  "metal",
  "UPVC",
];

export const BENTUK_ATAP = ["gable", "hip", "flat"];

export const INVOICE_STATUS = [
  { value: "draft", label: "Draft", color: "secondary" },
  { value: "sent", label: "Terkirim", color: "default" },
  { value: "paid", label: "Lunas", color: "success" },
] as const;

export const FEEDBACK_STATUS = [
  { value: "open", label: "Terbuka", color: "default" },
  { value: "in_progress", label: "Diproses", color: "warning" },
  { value: "resolved", label: "Selesai", color: "success" },
  { value: "closed", label: "Ditutup", color: "secondary" },
] as const;

export const SIDEBAR_WIDTH = "17.5rem"; // 280px
