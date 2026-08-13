import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";

const DARK: [number, number, number] = [31, 41, 55];
const GRAY: [number, number, number] = [107, 114, 128];
const AMBER: [number, number, number] = [217, 119, 6];
const LIGHT: [number, number, number] = [249, 250, 251];

const COMPANY = {
  name: "HALORA LAND",
  addressLine1: "Jl. Adam Malik No. 58, Ruko No.1, Cipadu Jaya, Larangan,",
  addressLine2: "Kota Tangerang, 15155 - Indonesia.",
  contacts: "W: land.halora.id | e: halo@halora.id | wa: +62 811 8622 225",
};

const LOGO_RAB = "/halora_land_gold.jpeg";

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

function formatDate(date: Date): string {
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(date);
}

function sanitizeFileName(name: string): string {
  return name.replace(/[^\w\- ]+/g, "").trim().replace(/\s+/g, "-") || "RAB";
}

function lastTableY(doc: jsPDF): number {
  return (doc as unknown as { lastAutoTable?: { finalY: number } }).lastAutoTable
    ?.finalY ?? 0;
}

export async function exportRekapPDF(
  data: RekapData,
  options: RekapPdfOptions = {}
): Promise<void> {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const W = doc.internal.pageSize.getWidth();
  const H = doc.internal.pageSize.getHeight();
  const X = 12;
  const XR = W - 12;
  const contentWidth = XR - X;
  const grouped = data.grouped || {};
  const project = data.project;

  const logo = await loadLogo(LOGO_RAB).catch(() => null);
  if (logo) {
    doc.addImage(logo, "JPEG", 134.5, 21.5, 38.5, 22.3);
  }

  // --- Title ---
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.setTextColor(...DARK);
  doc.text("RENCANA ANGGARAN BIAYA", W / 2, 18, { align: "center" });

  // --- Info block ---
  const infoRows: Array<[string, string]> = [
    ["PROYEK", project?.name || "-"],
    ["PEMILIK", options.ownerName || "-"],
    ["LOKASI PROYEK", project?.location || "-"],
    ["LUAS BANGUNAN", options.area || "-"],
  ];

  autoTable(doc, {
    startY: 46,
    margin: { left: X, right: X },
    theme: "grid",
    head: [["", ""]],
    body: infoRows,
    styles: { fontSize: 9, cellPadding: 2.5, textColor: DARK },
    headStyles: { fillColor: AMBER, textColor: [255, 255, 255], fontSize: 9 },
    columnStyles: {
      0: { cellWidth: 40, fontStyle: "bold", textColor: GRAY, halign: "left" },
    },
  });

  let y = lastTableY(doc) + 6;

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

  let sectionIndex = 0;
  for (const [category, title] of CATEGORY_TITLES) {
    const items = (grouped[category] || []) as RekapItem[];
    if (items.length === 0) continue;

    if (y + 14 > H - 34) {
      doc.addPage();
      y = 15;
    }

    // Section header
    doc.setFillColor(...LIGHT);
    doc.rect(X, y, contentWidth, 8, "F");
    doc.setTextColor(...AMBER);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(10);
    const roman = ROMAN[sectionIndex] || String(sectionIndex + 1);
    doc.text(`${roman}. ${title}`, X + 2, y + 5.5);
    sectionIndex += 1;
    y += 10;

    const sectionTotal = (items as RekapItem[]).reduce(
      (s, it) => s + Number(it.totalCost || 0),
      0
    );

    autoTable(doc, {
      startY: y,
      margin: { left: X, right: X },
      head:
        sectionIndex === 1
          ? [["No.", "Uraian Pekerjaan", "Spesifikasi", "Sat", "Vol", "Harga Sat", "Tot Harga"]]
          : [],
      body: [
        ...(items as RekapItem[]).map((item, idx) => [
          String(idx + 1),
          item.description,
          item.level || "-",
          item.unit || "-",
          formatNumber(Number(item.volume)),
          formatMoney(Number(item.unitPrice)),
          formatMoney(Number(item.totalCost)),
        ]),
        ["", "Sub Total :", "", "", "", "", formatMoney(sectionTotal)],
      ],
      theme: "grid",
      styles: { fontSize: 8.5, cellPadding: 2, textColor: DARK },
      headStyles: {
        fillColor: DARK,
        textColor: [255, 255, 255],
        fontStyle: "bold",
        fontSize: 8.5,
      },
      columnStyles: {
        0: { cellWidth: 8, halign: "center" },
        1: { cellWidth: 62, halign: "left" },
        2: { cellWidth: 34, halign: "left" },
        3: { cellWidth: 14, halign: "center" },
        4: { cellWidth: 18, halign: "right" },
        5: { cellWidth: 24, halign: "right" },
        6: { cellWidth: 26, halign: "right", fontStyle: "bold" },
      },
      didParseCell: (cell) => {
        if (cell.row.index === items.length) {
          cell.cell.styles.fillColor = LIGHT;
          cell.cell.styles.fontStyle = "bold";
          cell.cell.styles.textColor = DARK;
        }
      },
    });

    y = lastTableY(doc) + 8;
  }

  if (sectionIndex === 0) {
    doc.setTextColor(...GRAY);
    doc.setFontSize(10);
    doc.text("Belum ada item pekerjaan di RAB.", X, y);
    y += 10;
  }

  // --- Grand totals ---
  const totalFinal = Number(summary.totalFinal || 0);
  const totalRows: string[][] = [
    ["Total :", formatMoney(subtotal)],
    ["Total keseluruhan", `Rp ${formatMoney(totalFinal)}`],
  ];

  if (y + 20 > H - 34) {
    doc.addPage();
    y = 15;
  }

  autoTable(doc, {
    startY: y,
    margin: { left: XR - 100, right: X },
    theme: "grid",
    body: totalRows,
    styles: { fontSize: 10, cellPadding: 3, textColor: DARK },
    columnStyles: {
      0: { cellWidth: 60, fontStyle: "bold" },
      1: { cellWidth: 40, halign: "right", fontStyle: "bold" },
    },
    didParseCell: (cell) => {
      if (cell.row.index === totalRows.length - 1) {
        cell.cell.styles.fillColor = [250, 240, 225];
        cell.cell.styles.textColor = AMBER;
      }
    },
  });

  // --- Footer ---
  const footerY = H - 28;
  doc.setDrawColor(...GRAY);
  doc.setLineWidth(0.3);
  doc.line(X, footerY, XR, footerY);

  doc.setFont("helvetica", "bold");
  doc.setFontSize(9);
  doc.setTextColor(...DARK);
  doc.text(COMPANY.name, X, footerY + 6);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(...GRAY);
  doc.text(COMPANY.addressLine1, X, footerY + 11);
  doc.text(COMPANY.addressLine2, X, footerY + 16);
  doc.text(COMPANY.contacts, X, footerY + 21);

  doc.text(
    `Dicetak : ${formatDate(new Date())}`,
    XR,
    footerY + 6,
    { align: "right" }
  );

  const fileName = `Rekapitulasi-RAB-${sanitizeFileName(
    project?.name || "Project"
  )}.pdf`;
  doc.save(fileName);
}