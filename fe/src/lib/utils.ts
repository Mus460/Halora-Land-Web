import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value);
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat("id-ID").format(value);
}

export function parseCurrency(value: string): number {
  return Number(value.replace(/[^0-9-]/g, "")) || 0;
}

export function formatDate(date: string | Date): string {
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(date));
}

export function formatDateShort(date: string | Date): string {
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(date));
}

export function calculateAHS(
  volume: number,
  components: { koef: number; hargaSatuan: number; tipe: string }[]
) {
  const breakdown = { material: 0, upah: 0, alat: 0 };
  let hargaSatuan = 0;

  for (const c of components) {
    const total = c.koef * c.hargaSatuan;
    hargaSatuan += total;
    if (c.tipe in breakdown) {
      breakdown[c.tipe as keyof typeof breakdown] += total;
    }
  }

  return {
    volume,
    hargaSatuan,
    totalBiaya: volume * hargaSatuan,
    breakdown,
  };
}

export function calculateRAB(
  subtotal: number,
  marginPercent: number = 0
) {
  const overhead = subtotal * 0.1;
  const profit = (subtotal + overhead) * (marginPercent / 100);
  const ppn = (subtotal + overhead + profit) * 0.11;
  const total = subtotal + overhead + profit + ppn;

  return { subtotal, overhead, profit, ppn, total };
}

export function generateId(): number {
  return Date.now() + Math.floor(Math.random() * 1000);
}
