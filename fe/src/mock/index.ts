import type {
  Proyek,
  Pekerjaan,
  MasterHarga,
  MasterAnalisa,
  User,
  Invoice,
  Logistik,
  Realisasi,
  News,
  KategoriPekerjaan,
} from "@/types";

import usersData from "./users.json";
import proyekData from "./proyek.json";
import masterHargaData from "./master-harga.json";
import rekapData from "./rekap.json";
import invoiceData from "./invoice.json";
import logistikData from "./logistik.json";
import realisasiData from "./realisasi.json";
import newsData from "./news.json";
import dashboardData from "./dashboard.json";
import kurvaSData from "./kurva-s.json";
import monitoringData from "./monitoring.json";

import persiapanData from "./pekerjaan/persiapan.json";
import pondasiData from "./pekerjaan/pondasi.json";
import betonData from "./pekerjaan/beton.json";
import kanopiData from "./pekerjaan/kanopi.json";
import bajaData from "./pekerjaan/baja.json";
import tanggaData from "./pekerjaan/tangga.json";
import atapData from "./pekerjaan/atap.json";
import dindingData from "./pekerjaan/dinding.json";
import plesteranData from "./pekerjaan/plesteran.json";
import acianData from "./pekerjaan/acian.json";
import keramikData from "./pekerjaan/keramik.json";
import pavingData from "./pekerjaan/paving.json";
import pengecatanData from "./pekerjaan/pengecatan.json";
import pintuData from "./pekerjaan/pintu.json";
import interiorData from "./pekerjaan/interior.json";
import toiletData from "./pekerjaan/toilet.json";
import mepData from "./pekerjaan/mep.json";
import customData from "./pekerjaan/custom.json";

const pekerjaanMap: Record<KategoriPekerjaan, Pekerjaan[]> = {
  persiapan: persiapanData as Pekerjaan[],
  pondasi: pondasiData as Pekerjaan[],
  beton: betonData as Pekerjaan[],
  kanopi: kanopiData as Pekerjaan[],
  baja: bajaData as Pekerjaan[],
  tangga: tanggaData as Pekerjaan[],
  atap: atapData as Pekerjaan[],
  dinding: dindingData as Pekerjaan[],
  plesteran: plesteranData as Pekerjaan[],
  acian: acianData as Pekerjaan[],
  keramik: keramikData as Pekerjaan[],
  paving: pavingData as Pekerjaan[],
  pengecatan: pengecatanData as Pekerjaan[],
  pintu: pintuData as Pekerjaan[],
  interior: interiorData as Pekerjaan[],
  toilet: toiletData as Pekerjaan[],
  mep: mepData as Pekerjaan[],
  custom: customData as Pekerjaan[],
};

export function getUsers(): User[] {
  return usersData as User[];
}

export function getUserById(id: number): User | undefined {
  return (usersData as User[]).find((u) => u.id === id);
}

export function getProyekList(): Proyek[] {
  return proyekData as Proyek[];
}

export function getProyekById(id: number): Proyek | undefined {
  return (proyekData as Proyek[]).find((p) => p.id === id);
}

export function getProyekByUser(userId: number): Proyek[] {
  return (proyekData as Proyek[]).filter((p) => p.userId === userId);
}

export function getMasterHarga(): MasterHarga[] {
  return masterHargaData as MasterHarga[];
}

export function getMasterHargaByKategori(kategori: string): MasterHarga[] {
  return (masterHargaData as MasterHarga[]).filter(
    (h) => h.kategori === kategori
  );
}

export function getPekerjaanByKategori(
  kategori: KategoriPekerjaan,
  proyekId?: number
): Pekerjaan[] {
  const items = pekerjaanMap[kategori] || [];
  if (proyekId) {
    return items.filter((p) => p.proyekId === proyekId);
  }
  return items;
}

export function getAllPekerjaan(proyekId?: number): Pekerjaan[] {
  const all = Object.values(pekerjaanMap).flat();
  if (proyekId) {
    return all.filter((p) => p.proyekId === proyekId);
  }
  return all;
}

export function getRekap(proyekId?: number) {
  const items = rekapData as any[];
  if (proyekId) {
    return items.filter((r) => r.proyekId === proyekId);
  }
  return items;
}

export function getInvoiceList(proyekId?: number): Invoice[] {
  const items = invoiceData as Invoice[];
  if (proyekId) {
    return items.filter((i) => i.proyekId === proyekId);
  }
  return items;
}

export function getLogistik(proyekId?: number): Logistik[] {
  const items = logistikData as Logistik[];
  if (proyekId) {
    return items.filter((l) => l.proyekId === proyekId);
  }
  return items;
}

export function getRealisasi(proyekId?: number): Realisasi[] {
  const items = realisasiData as Realisasi[];
  if (proyekId) {
    return items.filter((r) => r.proyekId === proyekId);
  }
  return items;
}

export function getNewsList(): News[] {
  return newsData as News[];
}

export function getDashboardData() {
  return dashboardData;
}

export function getKurvaSData() {
  return kurvaSData;
}

export function getMonitoringData() {
  return monitoringData;
}

export function getMasterAnalisaTree(): MasterAnalisa[] {
  return [
    {
      id: 1,
      kode: "A",
      nama: "PEKERJAAN PERSIAPAN",
      level: 0,
      parentId: null,
      satuan: null,
      isGlobal: true,
      userId: null,
      createdAt: "2026-01-01T00:00:00Z",
      children: [
        {
          id: 2,
          kode: "A.1",
          nama: "Pembersihan Lokasi",
          level: 1,
          parentId: 1,
          satuan: "m2",
          isGlobal: true,
          userId: null,
          createdAt: "2026-01-01T00:00:00Z",
        },
        {
          id: 3,
          kode: "A.2",
          nama: "Bouwplank",
          level: 1,
          parentId: 1,
          satuan: "m1",
          isGlobal: true,
          userId: null,
          createdAt: "2026-01-01T00:00:00Z",
        },
      ],
    },
    {
      id: 10,
      kode: "B",
      nama: "PEKERJAAN TANAH",
      level: 0,
      parentId: null,
      satuan: null,
      isGlobal: true,
      userId: null,
      createdAt: "2026-01-01T00:00:00Z",
      children: [
        {
          id: 11,
          kode: "B.1",
          nama: "Galian Tanah",
          level: 1,
          parentId: 10,
          satuan: "m3",
          isGlobal: true,
          userId: null,
          createdAt: "2026-01-01T00:00:00Z",
        },
      ],
    },
    {
      id: 20,
      kode: "C",
      nama: "PEKERJAAN PONDASI",
      level: 0,
      parentId: null,
      satuan: null,
      isGlobal: true,
      userId: null,
      createdAt: "2026-01-01T00:00:00Z",
      children: [
        {
          id: 21,
          kode: "C.1",
          nama: "Pondasi Batu Kali",
          level: 1,
          parentId: 20,
          satuan: "m3",
          isGlobal: true,
          userId: null,
          createdAt: "2026-01-01T00:00:00Z",
        },
      ],
    },
  ];
}
