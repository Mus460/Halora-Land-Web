import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";
import type { Invoice } from "@/types";
import { terbilang } from "./terbilang";

const NAVY: [number, number, number] = [16, 42, 67];
const TEAL: [number, number, number] = [23, 126, 137];
const GOLD: [number, number, number] = [232, 163, 23];
const LIGHT_NAVY: [number, number, number] = [238, 243, 247];
const LIGHT_GOLD: [number, number, number] = [255, 245, 222];
const GOLD_BORDER: [number, number, number] = [241, 200, 117];
const GREEN: [number, number, number] = [34, 139, 90];
const TEXT: [number, number, number] = [36, 55, 70];
const MUTED: [number, number, number] = [113, 128, 140];
const LINE: [number, number, number] = [215, 224, 230];
const WHITE: [number, number, number] = [255, 255, 255];
const WHITE75: [number, number, number] = [191, 191, 191];

const COMPANY = {
  name: "HALORA LAND",
  tagline: '"Business is about Trust and Value"',
  address:
    "Jl. Adam Malik No. 58, Ruko No.1, Cipadu Jaya, Larangan, Kota Tangerang, 15155 - Indonesia.",
  contacts: "land.halora.id | halo@halora.id | +62 811 8622 225",
  website: "land.halora.id",
};

const LOGO_LEFT = "/halora_land_supply_nicer.jpeg";
const LOGO_RIGHT = "/halora_land_blue.jpeg";

const logoCache = new Map<string, Promise<string>>();

function loadLogo(src: string): Promise<string> {
  if (!logoCache.has(src)) {
    logoCache.set(
      src,
      fetch(src)
        .then((res) => res.blob())
        .then(
          (blob) =>
            new Promise<string>((resolve, reject) => {
              const reader = new FileReader();
              reader.onload = () => resolve(reader.result as string);
              reader.onerror = () => reject(reader.error);
              reader.readAsDataURL(blob);
            })
        )
    );
  }
  return logoCache.get(src)!;
}

function rupiah(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value || 0);
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 2,
  }).format(value || 0);
}

function formatDate(date: string | null | undefined): string {
  if (!date) return "—";
  const d = new Date(date);
  if (isNaN(d.getTime())) return date;
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(d);
}

function lastTableY(doc: jsPDF): number {
  return (doc as unknown as { lastAutoTable?: { finalY: number } }).lastAutoTable
    ?.finalY ?? 0;
}

function drawTopAccent(doc: jsPDF, navyH: number, goldH: number): void {
  doc.setFillColor(...NAVY);
  doc.rect(0, 0, doc.internal.pageSize.getWidth(), navyH, "F");
  doc.setFillColor(...GOLD);
  doc.rect(0, navyH, doc.internal.pageSize.getWidth(), goldH, "F");
}

function drawFooter(
  doc: jsPDF,
  left: string,
  right: string,
  ruleY: number
): void {
  const W = doc.internal.pageSize.getWidth();
  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.12);
  doc.line(13.5, ruleY, W - 13.5, ruleY);
  doc.setFont("helvetica", "italic");
  doc.setFontSize(7);
  doc.setTextColor(...MUTED);
  doc.text(left, 13.5, ruleY + 5);
  doc.setFont("helvetica", "normal");
  doc.text(right, W - 13.5, ruleY + 5, { align: "right" });
}

function drawCard(
  doc: jsPDF,
  x: number,
  y: number,
  w: number,
  h: number,
  fill: [number, number, number],
  border: [number, number, number]
): void {
  doc.setFillColor(...fill);
  doc.setDrawColor(...border);
  doc.setLineWidth(0.3);
  doc.roundedRect(x, y, w, h, 1.8, 1.8, "FD");
}

export async function exportInvoicePdf(
  inv: Invoice,
  project?: { name?: string; location?: string | null }
): Promise<void> {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const W = doc.internal.pageSize.getWidth();
  const H = doc.internal.pageSize.getHeight();
  const M = 13.5;
  const XR = W - M;
  const CW = XR - M;

  const [logoLeft, logoRight] = await Promise.all([
    loadLogo(LOGO_LEFT).catch(() => null),
    loadLogo(LOGO_RIGHT).catch(() => null),
  ]);

  // --- Top accent bands ---
  drawTopAccent(doc, 2.2, 0.8);

  // --- Header: logos + company block ---
  if (logoLeft) {
    doc.addImage(logoLeft, "JPEG", M, 8.5, 27, 12.5);
  }
  if (logoRight) {
    doc.addImage(logoRight, "JPEG", XR - 30, 8.5, 30, 13.5);
  }

  const cx = 50;
  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("INVOICE / PROJECT BILLING", cx, 11.5);
  doc.setFontSize(17);
  doc.setTextColor(...NAVY);
  doc.text(COMPANY.name, cx, 16.5);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  const addressLines = doc.splitTextToSize(COMPANY.address, 92) as string[];
  addressLines.forEach((line, i) => {
    doc.text(line, cx, 21.5 + i * 3.2);
  });
  doc.text(COMPANY.contacts, cx, 21.5 + addressLines.length * 3.2 + 0.8);

  // --- Divider ---
  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.21);
  doc.line(M, 32.5, XR, 32.5);

  const buyerName = inv.buyerName || project?.name || `Proyek #${inv.projectId}`;
  const buyerAddress = inv.buyerAddress || project?.location || null;

  const items =
    inv.items && inv.items.length > 0
      ? inv.items
      : [
          {
            description: buyerName,
            qty: 1,
            unit: "ls",
            unitPrice: Number(inv.total) || 0,
          },
        ];
  const subtotal = items.reduce(
    (sum, i) => sum + (Number(i.qty) || 0) * (Number(i.unitPrice) || 0),
    0
  );
  const discount = Number(inv.discount) || 0;
  const taxRate = Number(inv.taxRate) || 0;
  const tax = (Math.max(subtotal - discount, 0) * taxRate) / 100;
  const grandTotal =
    Number(inv.total) || Math.max(subtotal - discount + tax, 0);
  const isPaid = inv.status === "paid";
  const paidSoFar = isPaid ? grandTotal : 0;
  const remaining = Math.max(grandTotal - paidSoFar, 0);
  const statusText = isPaid ? "LUNAS" : "REQUEST PEMBAYARAN 100%";

  // --- Title + status card ---
  doc.setFont("helvetica", "bold");
  doc.setFontSize(28);
  doc.setTextColor(...NAVY);
  doc.text("INVOICE", M, 47);
  doc.setFontSize(9);
  doc.setTextColor(...TEAL);
  doc.text("Tagihan pekerjaan / material", M, 52.5);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(...MUTED);
  doc.text("Dokumen tagihan resmi HALORA LAND.", M, 57);

  const cardX = XR - 57.5;
  drawCard(doc, cardX, 39, 57.5, 22, LIGHT_GOLD, GOLD_BORDER);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("NO. INVOICE", cardX + 6, 45);
  doc.setFontSize(10.5);
  doc.setTextColor(...TEXT);
  doc.text(String(inv.number || ""), cardX + 6, 50.5);
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("STATUS", cardX + 6, 55);
  doc.setFontSize(9);
  doc.setTextColor(...GOLD);
  doc.text(statusText, cardX + 6, 59.5);

  // --- Customer / date cards ---
  const cardsY = 65;
  const cardsH = 24;
  const cardW = 58;
  const gaps = (CW - 3 * cardW) / 2;
  drawCard(doc, M, cardsY, cardW, cardsH, LIGHT_NAVY, LINE);
  drawCard(doc, M + cardW + gaps, cardsY, cardW, cardsH, LIGHT_NAVY, LINE);
  drawCard(doc, M + 2 * (cardW + gaps), cardsY, cardW, cardsH, LIGHT_NAVY, LINE);

  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("CUSTOMER", M + 6, cardsY + 7.5);
  doc.setFontSize(13);
  doc.setTextColor(...NAVY);
  doc.text(buyerName, M + 6, cardsY + 13.8);
  if (buyerAddress) {
    doc.setFont("helvetica", "normal");
    doc.setFontSize(6.5);
    doc.setTextColor(...MUTED);
    const addrLines = doc.splitTextToSize(String(buyerAddress), cardW - 12) as string[];
    addrLines.slice(0, 2).forEach((line, i) => {
      doc.text(line, M + 6, cardsY + 18.6 + i * 3.2);
    });
  }

  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("INVOICE DATE", M + cardW + gaps + 6, cardsY + 7.5);
  doc.setFontSize(10);
  doc.setTextColor(...TEXT);
  doc.text(formatDate(inv.date), M + cardW + gaps + 6, cardsY + 14.2);

  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("VALID DATE", M + 2 * (cardW + gaps) + 6, cardsY + 7.5);
  doc.setFontSize(10);
  doc.setTextColor(...TEXT);
  doc.text(formatDate(inv.dueDate), M + 2 * (cardW + gaps) + 6, cardsY + 14.2);

  // --- Detail table ---
  const detailY = cardsY + cardsH + 7;
  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  doc.setTextColor(...NAVY);
  doc.text("DETAIL TAGIHAN", M, detailY);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...MUTED);
  doc.text("Rincian pekerjaan dan material", M, detailY + 4.5);

  autoTable(doc, {
    startY: detailY + 8,
    margin: { left: M, right: M },
    head: [["#", "Uraian", "Qty", "Sat.", "Harga Satuan", "Jumlah"]],
    body: items.map((item, idx) => {
      const descLines = String(item.description || buyerName).split("\n");
      const product = descLines[0];
      const notes = descLines.slice(1).map((n) => n.trim()).filter(Boolean);
      return [
        String(idx + 1),
        [product, ...notes].join("\n"),
        formatNumber(Number(item.qty) || 0),
        String(item.unit || ""),
        rupiah(Number(item.unitPrice)),
        rupiah((Number(item.qty) || 0) * (Number(item.unitPrice) || 0)),
      ];
    }),
    theme: "plain",
    headStyles: {
      fillColor: NAVY,
      textColor: WHITE,
      fontStyle: "bold",
      fontSize: 7.5,
      halign: "center",
      cellPadding: { top: 2, bottom: 2, left: 2, right: 2 },
    },
    styles: {
      fontSize: 8,
      textColor: TEXT,
      cellPadding: { top: 1.8, bottom: 1.8, left: 2, right: 2 },
    },
    alternateRowStyles: { fillColor: LIGHT_NAVY },
    columnStyles: {
      0: { cellWidth: 7, halign: "center" },
      1: { cellWidth: 87, halign: "left" },
      2: { cellWidth: 12, halign: "center" },
      3: { cellWidth: 10, halign: "center" },
      4: { cellWidth: 33, halign: "right" },
      5: { cellWidth: 34, halign: "right", fontStyle: "bold" },
    },
    didDrawCell: (data) => {
      if (data.section !== "body" || data.column.index !== 1) return;
      const raw = data.row.raw as string[];
      const parts = String(raw[1] || "").split("\n");
      if (parts.length < 2) return;
      const cell = data.cell;
      const padL = cell.padding("left");
      const padT = cell.padding("top");
      const padR = cell.padding("right");
      const padB = cell.padding("bottom");
      const innerW = cell.width - padL - padR;
      const fs = 8;
      const pt2mm = 25.4 / 72;
      const lh = fs * pt2mm * (doc.getLineHeightFactor ? doc.getLineHeightFactor() : 1.15);
      const prodLines = (doc.splitTextToSize(parts[0], innerW) as string[]).length;
      const total = prodLines + parts.length - 1;
      let ty =
        cell.y +
        (cell.height - padT - padB) / 2 +
        padT +
        fs * pt2mm * 0.85 -
        (total / 2) * lh;
      ty += prodLines * lh;
      doc.setFont("helvetica", "italic");
      doc.setFontSize(7);
      doc.setTextColor(...MUTED);
      parts.slice(1).forEach((note) => {
        const noteLines = doc.splitTextToSize(note, innerW) as string[];
        noteLines.forEach((line) => {
          doc.text(line, cell.x + padL, ty);
          ty += lh;
        });
      });
    },
  });

  // --- Terbilang + summary ---
  let sy = lastTableY(doc) + 9;
  if (sy + 82 > H - 28) {
    doc.addPage();
    sy = 18;
  }

  const terbilangLines = doc.splitTextToSize(
    terbilang(isPaid ? grandTotal : remaining),
    100
  ) as string[];
  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("TERBILANG", M, sy);
  const cardH = 9 + terbilangLines.length * 4.6;
  drawCard(doc, M, sy + 4.5, 116, cardH, LIGHT_GOLD, GOLD_BORDER);
  doc.setFont("helvetica", "italic");
  doc.setFontSize(8);
  doc.setTextColor(...TEXT);
  terbilangLines.forEach((line, i) => {
    doc.text(line, M + 6, sy + 4.5 + 7 + i * 4.6);
  });

  const sumX = M + 116 + 5;
  const sumW = XR - sumX;
  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(...MUTED);
  doc.text("Sub-Total", sumX, sy);
  doc.setFont("helvetica", "bold");
  doc.setTextColor(...TEXT);
  doc.text(rupiah(subtotal), XR, sy, { align: "right" });

  doc.setFont("helvetica", "normal");
  doc.setTextColor(...MUTED);
  doc.text("Sudah Terbayar", sumX, sy + 6);
  doc.setFont("helvetica", "bold");
  doc.setTextColor(...TEXT);
  doc.text(rupiah(paidSoFar), XR, sy + 6, { align: "right" });

  doc.setFont("helvetica", "bold");
  doc.setTextColor(...TEAL);
  doc.text("Request Pembayaran 100%", sumX, sy + 12);
  doc.text(isPaid ? "-" : rupiah(remaining), XR, sy + 12, { align: "right" });

  doc.setFont("helvetica", "normal");
  doc.setTextColor(...MUTED);
  doc.text(`PPN (${formatNumber(taxRate)}%)`, sumX, sy + 18);
  doc.setFont("helvetica", "bold");
  doc.setTextColor(...TEXT);
  doc.text(tax > 0 ? rupiah(tax) : "-", XR, sy + 18, { align: "right" });

  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.3);
  doc.line(sumX, sy + 23.5, XR, sy + 23.5);

  doc.setFont("helvetica", "bold");
  doc.setFontSize(9);
  doc.setTextColor(...NAVY);
  doc.text("SISA PEMBAYARAN", sumX, sy + 27.5);
  doc.setFontSize(15);
  doc.setTextColor(...GREEN);
  doc.text(rupiah(remaining), XR, sy + 28.5, { align: "right" });

  // --- Payment + accepted cards ---
  let py = Math.max(sy + 36, sy + 4.5 + cardH) + 7;
  if (py + 32 > H - 30) {
    doc.addPage();
    py = 18;
  }

  drawCard(doc, M, py, 90, 31, LIGHT_NAVY, LINE);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(9);
  doc.setTextColor(...NAVY);
  doc.text("INFORMASI PEMBAYARAN", M + 6, py + 6.5);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text(
    "Transfer pembayaran dapat dilakukan ke rekening berikut:",
    M + 6,
    py + 11.5
  );
  const paymentRows: Array<[string, string]> = [
    ["Bank", inv.paymentBank || "-"],
    ["No. Rekening", inv.paymentAccountNumber || "-"],
    ["A/N", inv.paymentAccountName || "-"],
  ];
  paymentRows.forEach(([label, value], i) => {
    const y = py + 17 + i * 5;
    doc.setFont("helvetica", "bold");
    doc.setFontSize(7.5);
    doc.setTextColor(...MUTED);
    doc.text(label, M + 6, y);
    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.setTextColor(...TEXT);
    doc.text(String(value), M + 36, y);
  });

  drawCard(doc, M + 93, py, 90, 31, WHITE, LINE);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(9);
  doc.setTextColor(...NAVY);
  doc.text("ACCEPTED BY", M + 93 + 6, py + 6.5);
  doc.setFontSize(14);
  doc.setTextColor(...TEXT);
  doc.text(inv.financeName || "-", M + 93 + 6, py + 13);
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text(COMPANY.name, M + 93 + 6, py + 19.5);
  doc.text(COMPANY.website, M + 93 + 6, py + 23.5);

  // --- Brand strip (last page) ---
  const stripY = H - 24;
  const stripH = 12;
  doc.setFillColor(...NAVY);
  doc.roundedRect(M, stripY, CW, stripH, 1.8, 1.8, "F");
  doc.setFont("helvetica", "bold");
  doc.setFontSize(9);
  doc.setTextColor(...WHITE);
  doc.text(COMPANY.name, M + 6, stripY + 5.2);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(6.5);
  doc.setTextColor(...WHITE75);
  doc.text(COMPANY.address, M + 6, stripY + 9.6);
  doc.text(COMPANY.contacts, XR - 6, stripY + 8, { align: "right" });

  // --- Fancy footer on every page ---
  const pages = doc.getNumberOfPages();
  for (let i = 1; i <= pages; i++) {
    doc.setPage(i);
    drawFooter(doc, COMPANY.tagline, `Halaman ${i} dari ${pages}`, H - 10);
  }
  doc.setPage(pages);

  const safeNumber = (inv.number || `INV-${inv.id}`).replace(/[/\\:]/g, "-");
  doc.save(`invoice-${safeNumber}.pdf`);
}
