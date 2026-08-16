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

// formatWeight renders a work-item cost weight (% of project subtotal).
export function formatWeight(value: number | null | undefined): string {
  if (value == null) return "—";
  if (value <= 0) return "0%";
  if (value < 0.1) return "<0.1%";
  return `${value.toFixed(1)}%`;
}

// formatVolume renders a work-item volume with at most 2 decimals
// (id-ID grouping), trimming trailing zeros ("12.5", "1.3333" -> "1.33").
export function formatVolume(value: number | string | null | undefined): string {
  if (value == null || value === "") return "0";
  const n = Number(value);
  if (!Number.isFinite(n)) return "0";
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 2,
  }).format(n);
}

export function formatTimeline(months: number, days: number): string {
  const parts: string[] = [];
  if (months > 0) parts.push(`${months} bulan`);
  if (days > 0) parts.push(`${days} hari`);
  return parts.join(" ");
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

// formatDuration renders an hour count as "X hari Y jam" (1 hari = 24 jam).
// Sub-day values are shown as plain hours.
export function formatDuration(value: number | null | undefined): string {
  if (value == null || value === 0) return "—";
  const fmt = (v: number) =>
    new Intl.NumberFormat("id-ID", { maximumFractionDigits: 1 }).format(v);
  const days = Math.floor(value / 24);
  if (days > 0) {
    return `${fmt(days)} hari ${fmt(value % 24)} jam`;
  }
  return `${fmt(value)} jam`;
}

export function calculateAHS(
  volume: number,
  components: { coefficient: number; unitPrice: number; type: string }[]
) {
  const breakdown = { material: 0, upah: 0, alat: 0 };
  let unitPrice = 0;

  for (const c of components) {
    const total = c.coefficient * c.unitPrice;
    unitPrice += total;
    if (c.type in breakdown) {
      breakdown[c.type as keyof typeof breakdown] += total;
    }
  }

  return {
    volume,
    unitPrice,
    totalCost: volume * unitPrice,
    breakdown,
  };
}

export function generateId(): number {
  return Date.now() + Math.floor(Math.random() * 1000);
}
