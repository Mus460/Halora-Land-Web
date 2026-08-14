import { jsPDF } from "jspdf";
import type { Invoice } from "@/types";
import { terbilang } from "./terbilang";

const NAVY: [number, number, number] = [16, 42, 67];
const TEAL: [number, number, number] = [23, 126, 137];
const GOLD: [number, number, number] = [232, 163, 23];
const GREEN: [number, number, number] = [34, 139, 90];
const LIGHT_NAVY: [number, number, number] = [238, 243, 247];
const LIGHT_GOLD: [number, number, number] = [255, 245, 222];
const TEXT: [number, number, number] = [36, 55, 70];
const MUTED: [number, number, number] = [113, 128, 140];
const LINE: [number, number, number] = [215, 224, 230];
const WHITE: [number, number, number] = [255, 255, 255];
const WHITE75: [number, number, number] = [191, 191, 191];

const COMPANY = {
  name: "HALORA LAND",
  billing: "INVOICE / PROJECT BILLING",
  tagline: "Tagihan pekerjaan / material",
  subtitle: "Dokumen tagihan resmi HALORA LAND.",
  address:
    "Jl. Adam Malik No. 58, Ruko No.1, Cipadu Jaya, Larangan, Kota Tangerang, 15155 - Indonesia.",
  website: "land.halora.id",
  email: "halo@halora.id",
  wa: "+62 811 8622 225",
  quote: "Business is about Trust and Value",
};

const LOGO_GALONA = "/halora-galona.png";
const LOGO_LAND = "/halora-land.png";

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

function num(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 0,
  }).format(value || 0);
}

function numDec(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 2,
  }).format(value || 0);
}

function rupiah(value: number): string {
  return `Rp ${num(value)}`;
}

function formatDate(date: string | null | undefined): string {
  if (!date) return "—";
  const d = new Date(date);
  if (isNaN(d.getTime())) return date;
  return new Intl.DateTimeFormat("en-GB", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(d);
}

const MM = 13.5; // template geometry margin
const A4_W = 210;
const CR = A4_W - MM; // content right edge
const PAGE_BOTTOM = 262; // keep clear of the footer strip + foot rule

export async function exportInvoicePdf(
  inv: Invoice,
  project?: { name?: string; location?: string | null }
): Promise<void> {
  const doc = new jsPDF({ unit: "mm", format: "a4" });

  const [galona, land] = await Promise.all([
    loadLogo(LOGO_GALONA).catch(() => null),
    loadLogo(LOGO_LAND).catch(() => null),
  ]);

  const buyerName =
    inv.buyerName || project?.name || `Proyek #${inv.projectId}`;

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
  const paidSoFar = inv.status === "paid" ? grandTotal : 0;
  const remaining = Math.max(grandTotal - paidSoFar, 0);
  const statusLabel =
    inv.status === "paid" ? "LUNAS" : "REQUEST PEMBAYARAN 100%";

  // ============================================================
  // TOP ACCENT BANDS
  // ============================================================
  doc.setFillColor(...NAVY);
  doc.rect(0, 0, A4_W, 2.2, "F");
  doc.setFillColor(...GOLD);
  doc.rect(0, 2.2, A4_W, 0.8, "F");

  // ============================================================
  // HEADER
  // ============================================================
  if (galona) doc.addImage(galona, "PNG", MM, 4.2, 27, 15);
  if (land) doc.addImage(land, "PNG", CR - 30, 4.6, 30, 14.4);

  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text(COMPANY.billing, 46, 6.6);

  doc.setFontSize(18);
  doc.setTextColor(...NAVY);
  doc.text(COMPANY.name, 46, 12.8);

  doc.setFont("helvetica", "normal");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  const addrLines = doc.splitTextToSize(COMPANY.address, 116) as string[];
  addrLines.forEach((l, i) => doc.text(l, 46, 18 + i * 2.7));
  doc.text(
    `${COMPANY.website}  |  ${COMPANY.email}  |  ${COMPANY.wa}`,
    46,
    addrLines.length * 2.7 + 17.5
  );

  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.21);
  doc.line(MM, 26.2, CR, 26.2);

  // ============================================================
  // TITLE + STATUS CARD
  // ============================================================
  doc.setFont("helvetica", "bold");
  doc.setFontSize(28);
  doc.setTextColor(...NAVY);
  doc.text("INVOICE", MM, 38);

  doc.setFontSize(10);
  doc.setTextColor(...TEAL);
  doc.text(COMPANY.tagline, MM, 43.4);

  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(...MUTED);
  doc.text(COMPANY.subtitle, MM, 47.6);

  const cardX = 122;
  const cardW = CR - cardX;
  doc.setFillColor(...LIGHT_GOLD);
  doc.setDrawColor(245, 212, 151);
  doc.setLineWidth(0.3);
  doc.roundedRect(cardX, 29.5, cardW, 24, 2.2, 2.2, "FD");
  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("NO. INVOICE", cardX + 5, 34.5);
  doc.setFontSize(10);
  doc.setTextColor(...TEXT);
  const invNo = String(inv.number || `INV-${inv.id}`);
  const noLines = doc.splitTextToSize(invNo, cardW - 12) as string[];
  noLines.forEach((l, i) => doc.text(l, cardX + 5, 40 + i * 4.2));
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("STATUS", cardX + 5, 47);
  doc.setFontSize(8.5);
  doc.setTextColor(...GOLD);
  doc.text(statusLabel, cardX + 5, 51.2);

  // ============================================================
  // CUSTOMER + DATE CARDS
  // ============================================================
  const cardTop = 58.5;
  const cardH = 15;
  const cardW3 = (CR - MM - 5) / 3;
  const cardXs = [MM, MM + cardW3 + 2.5, MM + 2 * (cardW3 + 2.5)];
  const cardData: Array<{ label: string; value: string; size: number }> = [
    { label: "CUSTOMER", value: buyerName, size: 12 },
    { label: "INVOICE DATE", value: formatDate(inv.date), size: 8.5 },
    { label: "VALID DATE", value: formatDate(inv.dueDate), size: 8.5 },
  ];
  cardData.forEach((c, i) => {
    doc.setFillColor(...LIGHT_NAVY);
    doc.setDrawColor(...LINE);
    doc.setLineWidth(0.25);
    doc.roundedRect(cardXs[i], cardTop, cardW3, cardH, 2.2, 2.2, "FD");
    doc.setFont("helvetica", "bold");
    doc.setFontSize(6.5);
    doc.setTextColor(...MUTED);
    doc.text(c.label, cardXs[i] + 4.5, cardTop + 6);
    doc.setFontSize(c.size);
    doc.setTextColor(...NAVY);
    doc.text(
      doc.splitTextToSize(c.value, cardW3 - 9) as string[],
      cardXs[i] + 4.5,
      cardTop + 11.2
    );
  });

  // ============================================================
  // DETAIL SECTION HEADING
  // ============================================================
  doc.setFont("helvetica", "bold");
  doc.setFontSize(11);
  doc.setTextColor(...NAVY);
  doc.text("DETAIL TAGIHAN", MM, 80.5);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...MUTED);
  doc.text("Rincian pekerjaan dan material", MM + 0.2, 84.5);

  // ============================================================
  // ITEM TABLE
  // ============================================================
  const colX = [
    MM, // 1: no (13.5)
    MM + 5.5, // 2: uraian (19)
    MM + 71, // 3: qty (84.5)
    MM + 80, // 4: sat (93.5)
    MM + 88, // 5: harga satuan (101.5)
    MM + 112.5, // 6: jumlah (126)
  ];
  const colRight = [
    colX[0] + 5.5,
    colX[1] + 65.5,
    colX[2] + 9,
    colX[3] + 8,
    colX[4] + 24.5,
    colX[5] + 25.5,
  ];
  const tableRight = colRight[5];
  const headH = 7;
  const headY = 88;
  const headLabels = ["#", "URAIAN", "QTY", "SAT.", "HARGA SATUAN", "JUMLAH"];
  const headAlign: Array<"left" | "center" | "right"> = [
    "center",
    "left",
    "center",
    "center",
    "right",
    "right",
  ];
  const headX = [
    colX[0] + 2.75,
    colX[1] + 2,
    colX[2] + 4.5,
    colX[3] + 4,
    colRight[4] - 3,
    colRight[5] - 3,
  ];

  const drawTableHead = (top: number): void => {
    doc.setFillColor(...NAVY);
    doc.rect(MM, top, tableRight - MM, headH, "F");
    doc.setFont("helvetica", "bold");
    doc.setFontSize(7);
    doc.setTextColor(...WHITE);
    headLabels.forEach((label, i) =>
      doc.text(label, headX[i], top + 4.8, { align: headAlign[i] })
    );
  };

  drawTableHead(headY);
  let y = headY + headH;
  const lineH = 3.7;

  items.forEach((item, idx) => {
    const desc = String(item.description || buyerName);
    const descLines = desc.split("\n");
    const product = descLines[0];
    const notes = descLines
      .slice(1)
      .map((n) => n.trim())
      .filter(Boolean);
    const wrap = doc.splitTextToSize(product, 61) as string[];
    const noteLines: string[][] = notes.map(
      (n) => doc.splitTextToSize(n, 61) as string[]
    );
    const noteCount = noteLines.reduce((s, ls) => s + ls.length, 0);
    const rh = 5.6 + wrap.length * lineH + noteCount * 2.7;

    if (y + rh > PAGE_BOTTOM) {
      doc.addPage();
      drawTableHead(20);
      y = 27;
    }

    if (idx % 2 === 1) {
      doc.setFillColor(...LIGHT_NAVY);
      doc.rect(MM, y, tableRight - MM, rh, "F");
    }

    const base = y + 4.6;
    doc.setFont("helvetica", "bold");
    doc.setFontSize(6.5);
    doc.setTextColor(...MUTED);
    doc.text(String(idx + 1).padStart(2, "0"), headX[0], base, {
      align: "center",
    });

    let ty = base;
    doc.setFont("helvetica", "bold");
    doc.setFontSize(8.5);
    doc.setTextColor(...TEXT);
    wrap.forEach((l) => {
      doc.text(l, colX[1] + 2, ty);
      ty += lineH;
    });
    if (noteCount > 0) {
      doc.setFont("helvetica", "italic");
      doc.setFontSize(6.5);
      doc.setTextColor(...MUTED);
      noteLines.forEach((ls) =>
        ls.forEach((l) => {
          doc.text(l, colX[1] + 2, ty);
          ty += 2.7;
        })
      );
    }

    doc.setFont("helvetica", "normal");
    doc.setFontSize(8.5);
    doc.setTextColor(...TEXT);
    doc.text(numDec(Number(item.qty) || 0), headX[2], base, {
      align: "center",
    });
    doc.text(String(item.unit || ""), headX[3], base, { align: "center" });
    doc.text(rupiah(Number(item.unitPrice)), headX[4], base, {
      align: "right",
    });
    doc.setFont("helvetica", "bold");
    doc.text(
      rupiah((Number(item.qty) || 0) * (Number(item.unitPrice) || 0)),
      headX[5],
      base,
      { align: "right" }
    );
    y += rh;
  });

  // ============================================================
  // SUMMARY + TERBILANG
  // ============================================================
  let sumTop = y + 3;
  if (sumTop + 45 > PAGE_BOTTOM) {
    doc.addPage();
    sumTop = 22;
  }

  doc.setFont("helvetica", "bold");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text("TERBILANG", MM, sumTop);

  const terbilangText = terbilang(remaining)
    .split(" ")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
  const tWrap = doc.splitTextToSize(terbilangText, 92) as string[];
  const tH = 4 + tWrap.length * 3.8 + 4;
  doc.setFillColor(...LIGHT_GOLD);
  doc.setDrawColor(247, 214, 151);
  doc.setLineWidth(0.25);
  doc.roundedRect(MM, sumTop + 3, 100, tH, 2.2, 2.2, "FD");
  doc.setFont("helvetica", "italic");
  doc.setFontSize(9);
  doc.setTextColor(...TEXT);
  tWrap.forEach((l, i) => doc.text(l, MM + 5, sumTop + 7.6 + i * 3.8));

  const sumX = 126;
  const sumW = CR - sumX;
  const rows: Array<{
    label: string;
    value: string;
    teal?: boolean;
    bold?: boolean;
  }> = [
    { label: "Sub-Total", value: rupiah(subtotal), bold: true },
    { label: "Sudah Terbayar", value: rupiah(paidSoFar) },
    {
      label: "Request Pembayaran 100%",
      value: rupiah(remaining),
      teal: true,
      bold: true,
    },
    { label: "PPN", value: taxRate > 0 ? `${numDec(taxRate)}%` : "—" },
  ];
  let sy = sumTop + 4.5;
  rows.forEach((r) => {
    doc.setFont("helvetica", "normal");
    doc.setFontSize(8.5);
    doc.setTextColor(...(r.teal ? TEAL : MUTED));
    doc.text(r.label, sumX, sy);
    doc.setFont("helvetica", r.bold ? "bold" : "normal");
    doc.setTextColor(...(r.teal ? TEAL : TEXT));
    doc.text(r.value, CR, sy, { align: "right" });
    sy += 5.6;
  });

  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.25);
  doc.line(sumX, sy - 1.6, CR, sy - 1.6);

  doc.setFont("helvetica", "bold");
  doc.setFontSize(8.5);
  doc.setTextColor(...NAVY);
  doc.text("SISA PEMBAYARAN", sumX, sy + 4.2);
  doc.setFontSize(15);
  doc.setTextColor(...GREEN);
  doc.text(rupiah(remaining), CR, sy + 4.2, { align: "right" });

  // ============================================================
  // PAYMENT + ACCEPTED BY
  // ============================================================
  let payTop = Math.max(sumTop + tH + 8, sy + 12) + 3;
  const payH = 26;
  if (payTop + payH > PAGE_BOTTOM) {
    doc.addPage();
    payTop = 20;
  }

  doc.setFillColor(...LIGHT_NAVY);
  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.25);
  doc.roundedRect(MM, payTop, 118, payH, 2.4, 2.4, "FD");
  doc.setFont("helvetica", "bold");
  doc.setFontSize(9.5);
  doc.setTextColor(...NAVY);
  doc.text("INFORMASI PEMBAYARAN", MM + 5, payTop + 8);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...MUTED);
  doc.text(
    "Transfer pembayaran dapat dilakukan ke rekening berikut:",
    MM + 5,
    payTop + 12.8
  );

  const bank = inv.paymentBank || "BSI / BCA";
  const account = inv.paymentAccountNumber || "1101101009";
  const holderName = inv.paymentAccountName || "Rangga Donyta Putra";
  const payRows: Array<[string, string]> = [
    ["Bank", bank],
    ["No. Rekening", account],
    ["Atas Nama", holderName],
  ];
  payRows.forEach(([label, value], i) => {
    const py = payTop + 17.8 + i * 3.6;
    doc.setFont("helvetica", "normal");
    doc.setFontSize(8);
    doc.setTextColor(...MUTED);
    doc.text(label, MM + 5, py);
    doc.setFont("helvetica", "bold");
    doc.setTextColor(...TEXT);
    doc.text(
      doc.splitTextToSize(value, 88) as string[],
      MM + 27,
      py
    );
  });

  doc.setFillColor(...WHITE);
  doc.setDrawColor(...LINE);
  doc.roundedRect(137.5, payTop, CR - 137.5, payH, 2.4, 2.4, "FD");
  doc.setFont("helvetica", "bold");
  doc.setFontSize(9.5);
  doc.setTextColor(...NAVY);
  doc.text("ACCEPTED BY", 142.5, payTop + 8);
  doc.setFontSize(13);
  doc.setTextColor(...TEXT);
  doc.text(inv.financeName || "Rangga", 142.5, payTop + 14.5);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(6.5);
  doc.setTextColor(...MUTED);
  doc.text(COMPANY.name, 142.5, payTop + 19.5);
  doc.text(COMPANY.website, 142.5, payTop + 22.5);

  // ============================================================
  // FOOTER BRAND STRIP
  // ============================================================
  let stripTop = Math.max(payTop + payH + 8, 276);
  if (stripTop + 11.5 > 287.5) {
    doc.addPage();
    stripTop = 20;
  }
  doc.setFillColor(...NAVY);
  doc.roundedRect(MM, stripTop, CR - MM, 11, 2.2, 2.2, "F");
  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  doc.setTextColor(...WHITE);
  doc.text(COMPANY.name, MM + 5, stripTop + 5.6);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(6.5);
  doc.setTextColor(...WHITE75);
  doc.text(
    doc.splitTextToSize(COMPANY.address, 120) as string[],
    MM + 5,
    stripTop + 9
  );
  doc.text(
    `${COMPANY.website}  |  ${COMPANY.email}  |  ${COMPANY.wa}`,
    CR - 5,
    stripTop + 9,
    { align: "right" }
  );

  // ============================================================
  // PAGE FOOTER (quote + page number)
  // ============================================================
  const pages = doc.getNumberOfPages();
  for (let i = 1; i <= pages; i++) {
    doc.setPage(i);
    doc.setDrawColor(...LINE);
    doc.setLineWidth(0.12);
    doc.line(MM, 287.5, CR, 287.5);
    doc.setFont("helvetica", "italic");
    doc.setFontSize(7);
    doc.setTextColor(...MUTED);
    doc.text(COMPANY.quote, MM, 291.5);
    doc.text(`Halaman ${i} dari ${pages}`, CR, 291.5, { align: "right" });
  }
  doc.setPage(pages);

  const safeNumber = (inv.number || `INV-${inv.id}`).replace(/[/\\:]/g, "-");
  doc.save(`invoice-${safeNumber}.pdf`);
}