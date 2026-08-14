import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";
import type { RowInput } from "jspdf-autotable";

const NAVY: [number, number, number] = [16, 42, 67];
const TEAL: [number, number, number] = [23, 126, 137];
const GOLD: [number, number, number] = [232, 163, 23];
const LIGHT_GOLD: [number, number, number] = [255, 245, 222];
const GOLD_BORDER: [number, number, number] = [244, 214, 151];
const TEXT: [number, number, number] = [36, 55, 70];
const MUTED: [number, number, number] = [113, 128, 140];
const LINE: [number, number, number] = [215, 224, 230];
const WHITE: [number, number, number] = [255, 255, 255];

const COMPANY = {
  short: "RAB - HALORA LAND",
  name: "HALORA LAND",
  type: "Construction",
  addressLine1: "Jl. Adam Malik No. 58, Ruko No.1, Cipadu Jaya,",
  addressLine2: "Larangan, Kota Tangerang, 15155 - Indonesia.",
  website: "land.halora.id",
  email: "halo@halora.id",
  wa: "+62 811 8622 225",
  siteLine: "HALORA LAND | land.halora.id",
};

const LOGO_RAB = "/halora-land-logo.png";

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

const CATEGORY_TITLES: Array<[string, string]> = [
  ["preparation", "PEKERJAAN PERSIAPAN"],
  ["foundation", "PEKERJAAN PONDASI"],
  ["concrete", "PEKERJAAN BETON STRUKTUR"],
  ["canopy", "PEKERJAAN KANOPI"],
  ["steel", "PEKERJAAN BAJA STRUKTURAL"],
  ["stairs", "PEKERJAAN TANGGA"],
  ["roof", "PEKERJAAN ATAP"],
  ["wall", "PEKERJAAN DINDING"],
  ["plastering", "PEKERJAAN PLESTERAN"],
  ["finishing", "PEKERJAAN ACIAN"],
  ["tiles", "PEKERJAAN LANTAI/KERAMIK"],
  ["paving", "PEKERJAAN PAVING & HALAMAN"],
  ["painting", "PEKERJAAN CAT & PLAFON"],
  ["doors", "PEKERJAAN KUSEN, PINTU & JENDELA"],
  ["interior", "PEKERJAAN INTERIOR"],
  ["toilet", "PEKERJAAN TOILET/SANITAIR"],
  ["mep", "PEKERJAAN INSTALASI MEP"],
  ["custom-work", "PEKERJAAN CUSTOM"],
];

const ROMAN: string[] = [
  "I",
  "II",
  "III",
  "IV",
  "V",
  "VI",
  "VII",
  "VIII",
  "IX",
  "X",
  "XI",
  "XII",
  "XIII",
  "XIV",
  "XV",
  "XVI",
  "XVII",
  "XVIII",
  "XIX",
  "XX",
];

export interface RekapItem {
  id: number;
  category: string;
  description: string;
  volume: number;
  unit: string;
  unitPrice: number;
  totalCost: number;
  totalDuration: number | null;
  level?: string | null;
}

export interface RekapData {
  project?: {
    id: number;
    name: string;
    location?: string | null;
    contractValue?: number | null;
  };
  projects?: {
    id: number;
    name: string;
    location?: string | null;
    contractValue?: number | null;
  };
  grouped?: Record<string, RekapItem[]>;
  subtotals?: Record<string, number>;
  summary?: {
    subtotal?: number;
    margin?: number;
    subtotalWithMargin?: number;
    overhead?: number;
    profit?: number;
    ppn?: number;
    totalPPN?: number;
    totalFinal?: number;
    totalDuration?: number;
  };
}

export interface RekapPdfOptions {
  ownerName?: string;
  area?: string;
}

function formatMoney(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 0,
  }).format(value || 0);
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 2,
  }).format(value || 0);
}

function sanitizeFileName(name: string): string {
  return name.replace(/[^\w\- ]+/g, "").trim().replace(/\s+/g, "-") || "RAB";
}

function lastTableY(doc: jsPDF): number {
  return (doc as unknown as { lastAutoTable?: { finalY: number } })
    .lastAutoTable?.finalY ?? 0;
}

export async function exportRekapPDF(
  data: RekapData,
  options: RekapPdfOptions = {}
): Promise<void> {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const W = doc.internal.pageSize.getWidth();
  const H = doc.internal.pageSize.getHeight();
  const M = 9.5;
  const XR = W - M;
  const CW = XR - M;
  const grouped = data.grouped || {};
  const project = data.project || data.projects;

  const logo = await loadLogo(LOGO_RAB).catch(() => null);

  // --- Top accent bands ---
  doc.setFillColor(...NAVY);
  doc.rect(0, 0, W, 1.5, "F");
  doc.setFillColor(...GOLD);
  doc.rect(0, 1.5, W, 0.6, "F");

  // --- Header: logo | title | company contact ---
  if (logo) {
    doc.addImage(logo, "PNG", M + 0.8, M, 28, 13.5);
  }

  doc.setFont("helvetica", "bold");
  doc.setFontSize(17);
  doc.setTextColor(...NAVY);
  doc.text("RENCANA ANGGARAN BIAYA", 52.5, 30.5);
  doc.setFontSize(8);
  doc.setTextColor(...TEAL);
  doc.text(COMPANY.type, 52.5, 35.2);

  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...MUTED);
  doc.text(COMPANY.addressLine1, XR, 31.5, { align: "right" });
  doc.text(COMPANY.addressLine2, XR, 35, { align: "right" });

  const rightLine = (
    parts: Array<[string, boolean]>, // [text, bold]
    y: number
  ): void => {
    const widths: number[] = [];
    parts.forEach(([text, bold]) => {
      doc.setFont("helvetica", bold ? "bold" : "normal");
      widths.push(doc.getTextWidth(text));
    });
    const total = widths.reduce((s, w) => s + w, 0);
    let x = XR - total;
    parts.forEach(([text, bold], i) => {
      doc.setFont("helvetica", bold ? "bold" : "normal");
      doc.text(text, x, y);
      x += widths[i];
    });
  };

  rightLine(
    [
      ["W:", true],
      [COMPANY.website, false],
    ],
    38.4
  );
  rightLine(
    [
      ["e:", true],
      [" halo@halora.id ", false],
      ["|", false],
      ["  wa:", true],
      ["  +62 811 8622 225", false],
    ],
    41.8
  );

  // --- Divider ---
  doc.setDrawColor(...LINE);
  doc.setLineWidth(0.19);
  doc.line(M, 43.5, XR, 43.5);

  // --- Info card: label (teal) + value (text) on the same row ---
  const cardY = 47;
  const padTop = 5.5;
  const padBottom = 5.5;
  const lineH = 3.6;

  const info: Array<{ label: string; value: string }> = [
    { label: "PROYEK", value: project?.name || "-" },
    { label: "PEMILIK", value: options.ownerName || "-" },
    { label: "LOKASI PROYEK", value: project?.location || "-" },
    { label: "LUAS BANGUNAN", value: options.area || "-" },
  ];
  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  const valueLines = info.map(
    (it) => doc.splitTextToSize(String(it.value), 70) as string[]
  );
  const rowHs = [
    Math.max(
      valueLines[0].length * lineH,
      valueLines[1].length * lineH,
      lineH
    ) + 2.2,
    Math.max(
      valueLines[2].length * lineH,
      valueLines[3].length * lineH,
      lineH
    ) + 2.2,
  ];
  const cardH = padTop + rowHs[0] + rowHs[1] + padBottom;

  doc.setFillColor(...LIGHT_GOLD);
  doc.setDrawColor(...GOLD_BORDER);
  doc.setLineWidth(0.3);
  doc.roundedRect(M, cardY, CW, cardH, 2.1, 2.1, "FD");

  let vy = cardY + padTop;
  info.forEach(({ label, value }, i) => {
    const col = i % 2;
    const row = Math.floor(i / 2);
    const lx = 12 + col * 92.4;
    const vx = lx + 21;
    doc.setFont("helvetica", "bold");
    doc.setFontSize(8);
    doc.setTextColor(...TEAL);
    doc.text(label, lx, vy + 3.4);
    doc.setFont("helvetica", "normal");
    doc.setFontSize(9.5);
    doc.setTextColor(...TEXT);
    valueLines[i].forEach((l, j) => doc.text(l, vx, vy + 3.4 + j * lineH));
    if (i % 2 === 1) vy += rowHs[row];
  });

  // --- Sectioned items ---
  const summary = data.summary || {};
  const subtotal =
    Number(summary.subtotal) ||
    Object.values(grouped).reduce(
      (sum, items) =>
        sum +
        (items || []).reduce(
          (s, it) => s + Number((it as RekapItem).totalCost || 0),
          0
        ),
      0
    );

  const body: RowInput[] = [];
  let sectionIndex = 0;
  let hasItems = false;
  for (const [category, title] of CATEGORY_TITLES) {
    const items = (grouped[category] || []) as RekapItem[];
    if (items.length === 0) continue;
    hasItems = true;

    body.push([
      {
        content: `${ROMAN[sectionIndex] || String(sectionIndex + 1)} ${title}`,
        colSpan: 7,
        styles: {
          fillColor: NAVY,
          textColor: WHITE,
          fontStyle: "bold",
          fontSize: 8,
          halign: "left",
          cellPadding: { top: 1.6, bottom: 1.6, left: 2.2, right: 2.2 },
        },
      },
    ]);
    sectionIndex += 1;

    items.forEach((item, idx) => {
      body.push([
        String(idx + 1),
        item.description,
        item.level || "",
        item.unit || "",
        formatNumber(Number(item.volume)),
        formatMoney(Number(item.unitPrice)),
        formatMoney(Number(item.totalCost)),
      ]);
    });

    const sectionTotal = (items as RekapItem[]).reduce(
      (s, it) => s + Number(it.totalCost || 0),
      0
    );
    body.push([
      {
        content: "Sub Total :",
        colSpan: 5,
        styles: {
          halign: "right",
          fontStyle: "bold",
          textColor: TEAL,
          cellPadding: { top: 1.4, bottom: 1.4, left: 2.2, right: 2.2 },
        },
      },
      {
        content: `Rp ${formatMoney(sectionTotal)}`,
        colSpan: 2,
        styles: {
          halign: "right",
          fontStyle: "bold",
          textColor: NAVY,
          cellPadding: { top: 1.4, bottom: 1.4, left: 2.2, right: 2.2 },
        },
      },
    ]);
  }

  autoTable(doc, {
    startY: cardY + cardH + 6,
    margin: { left: M, right: M },
    head: [
      [
        "No.",
        "Uraian Pekerjaan",
        "Spesifikasi",
        "Sat",
        "Vol",
        "Harga Sat",
        "Tot Harga",
      ],
    ],
    body,
    theme: "plain",
    headStyles: {
      fillColor: NAVY,
      textColor: WHITE,
      fontStyle: "bold",
      fontSize: 8,
      halign: "left",
      cellPadding: { top: 1.6, bottom: 1.6, left: 2.2, right: 2.2 },
    },
    styles: {
      fontSize: 8,
      textColor: TEXT,
      cellPadding: { top: 1.4, bottom: 1.4, left: 2.2, right: 2.2 },
    },
    columnStyles: {
      0: { cellWidth: 9, halign: "center", cellPadding: { top: 1.4, bottom: 1.4, left: 1, right: 1 } },
      1: { halign: "left" },
      2: { cellWidth: 31, halign: "left" },
      3: { cellWidth: 8.5, halign: "center" },
      4: { cellWidth: 9.5, halign: "right" },
      5: { cellWidth: 19.5, halign: "right" },
      6: { cellWidth: 20, halign: "right", fontStyle: "bold" },
    },
  });

  if (!hasItems) {
    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.setTextColor(...MUTED);
    doc.text(
      "Belum ada item pekerjaan di RAB.",
      M,
      Math.max(lastTableY(doc) + 4, cardY + cardH + 12)
    );
  }

  // --- Total bar ---
  const totalFinal = Number(summary.totalFinal || 0) || subtotal;
  let ty = lastTableY(doc) + 5;
  if (ty + 16 > H - 13) {
    doc.addPage();
    ty = 15;
  }
  doc.setFillColor(...NAVY);
  doc.roundedRect(M, ty, CW, 13, 2.5, 2.5, "F");
  doc.setFont("helvetica", "bold");
  doc.setFontSize(12);
  doc.setTextColor(...WHITE);
  doc.text("TOTAL KESELURUHAN", M + 5.5, ty + 8.6);
  doc.setFontSize(18);
  doc.text(`Rp ${formatMoney(totalFinal)}`, XR - 5.5, ty + 8.8, {
    align: "right",
  });

  // --- Footer note ---
  const noteY = Math.min(ty + 17, H - 15);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...MUTED);
  doc.text(
    "Catatan: angka dan subtotal mengikuti dokumen RAB sumber.",
    M,
    noteY,
    { maxWidth: 120 }
  );
  doc.text(COMPANY.siteLine, XR, noteY, { align: "right" });

  // --- Fancy footer on every page ---
  const pages = doc.getNumberOfPages();
  for (let i = 1; i <= pages; i++) {
    doc.setPage(i);
    doc.setDrawColor(...LINE);
    doc.setLineWidth(0.12);
    doc.line(M, H - 13, XR, H - 13);
    doc.setFont("helvetica", "normal");
    doc.setFontSize(7);
    doc.setTextColor(...MUTED);
    doc.text(COMPANY.short, M, H - 8);
    doc.text(`Halaman ${i} dari ${pages}`, XR, H - 8, { align: "right" });
  }
  doc.setPage(pages);

  const fileName = `Rekapitulasi-RAB-${sanitizeFileName(
    project?.name || "Project"
  )}.pdf`;
  doc.save(fileName);
}