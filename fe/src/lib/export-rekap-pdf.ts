import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";
import { formatDuration } from "./utils";

export interface RekapItem {
  id: number;
  category: string;
  description: string;
  volume: number;
  unit: string;
  unitPrice: number;
  totalCost: number;
  totalDuration: number | null;
}

export interface RekapData {
  project?: {
    id: number;
    name: string;
    location?: string | null;
    contractValue?: number | null;
  };
  grouped?: Record<string, RekapItem[]>;
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

const AMBER: [number, number, number] = [217, 119, 6];
const GRAY: [number, number, number] = [107, 114, 128];

function formatCurrency(value: number): string {
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

export function exportRekapPDF(data: RekapData): void {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const pageWidth = doc.internal.pageSize.getWidth();
  const marginX = 14;
  const contentWidth = pageWidth - marginX * 2;
  let y = 0;

  // Header band
  doc.setFillColor(...AMBER);
  doc.rect(0, 0, pageWidth, 30, "F");
  doc.setTextColor(255, 255, 255);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(18);
  doc.text("Halora Land", marginX, 14);
  doc.setFontSize(11);
  doc.setFont("helvetica", "normal");
  doc.text("Rekapitulasi Rencana Anggaran Biaya (RAB)", marginX, 21);

  // Project info
  y = 40;
  doc.setTextColor(30, 30, 30);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(12);
  const project = data.project;
  doc.text(project?.name || "Project", marginX, y);
  y += 6;

  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  const location = project?.location ? String(project.location) : "-";
  const contractValue = project?.contractValue
    ? formatCurrency(Number(project.contractValue))
    : "-";
  doc.setTextColor(...GRAY);
  doc.text(`Location: ${location}`, marginX, y);
  y += 5;
  doc.text(`Nilai Kontrak: ${contractValue}`, marginX, y);
  y += 5;
  doc.text(`Dicetak: ${formatDate(new Date())}`, marginX, y);
  y += 8;

  const grouped = data.grouped || {};
  const kategoriKeys = Object.keys(grouped);

  for (const category of kategoriKeys) {
    const items = grouped[category] || [];
    if (items.length === 0) continue;

    // Section header
    doc.setFillColor(250, 250, 250);
    doc.rect(marginX, y, contentWidth, 8, "F");
    doc.setTextColor(...AMBER);
    doc.setFont("helvetica", "bold");
    doc.setFontSize(10);
    doc.text(category.toUpperCase().replace("_", " "), marginX + 2, y + 5.5);

    y += 11;

    autoTable(doc, {
      startY: y,
      margin: { left: marginX, right: marginX },
      head: [["Uraian", "Volume", "Satuan", "Harga Satuan", "Jumlah Harga"]],
      body: items.map((item) => [
        item.description,
        formatNumber(Number(item.volume)),
        item.unit || "-",
        formatCurrency(Number(item.unitPrice)),
        formatCurrency(Number(item.totalCost)),
      ]),
      theme: "grid",
      styles: { fontSize: 9, cellPadding: 2.2, textColor: [50, 50, 50] },
      headStyles: {
        fillColor: AMBER,
        textColor: [255, 255, 255],
        fontStyle: "bold",
        fontSize: 9,
      },
      columnStyles: {
        1: { halign: "right" },
        2: { halign: "center" },
        3: { halign: "right" },
        4: { halign: "right", fontStyle: "bold" },
      },
      didParseCell: (cell) => {
        if (cell.column.index > 0 && cell.section === "body") {
          cell.cell.styles.halign = "right";
        }
      },
    });

    y = (doc as any).lastAutoTable.finalY + 8;
  }

  if (kategoriKeys.length === 0) {
    doc.setTextColor(...GRAY);
    doc.setFontSize(10);
    doc.text("Belum ada item pekerjaan di RAB.", marginX, y);
    y += 10;
  }

  // Summary section
  const summary = data.summary || {};
  const subtotal = Number(summary.subtotal || 0);
  const margin = Number(summary.margin || 0);
  const totalDuration = summary.totalDuration;

  if (y + 40 > doc.internal.pageSize.getHeight()) {
    doc.addPage();
    y = 20;
  }

  doc.setFillColor(250, 250, 250);
  doc.rect(marginX, y, contentWidth, 8, "F");
  doc.setTextColor(...AMBER);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  doc.text("RINGKASAN RAB", marginX + 2, y + 5.5);
  y += 11;

  const rows: [string, string, boolean?][] = [
    ["Subtotal WorkItem", formatCurrency(subtotal)],
    ["Estimasi Waktu", formatDuration(totalDuration)],
    [
      `Subtotal + Margin (${formatNumber(margin)}%)`,
      formatCurrency(Number(summary.subtotalWithMargin || 0)),
    ],
    ["Overhead (10%)", formatCurrency(Number(summary.overhead || 0))],
    ["Profit", formatCurrency(Number(summary.profit || 0))],
    [
      `PPN (${formatNumber(Number(summary.ppn || 0))}%)`,
      formatCurrency(Number(summary.totalPPN || 0)),
    ],
  ];

  autoTable(doc, {
    startY: y,
    margin: { left: marginX, right: marginX },
    body: rows.map(([label, value]) => [label, value]),
    theme: "plain",
    styles: { fontSize: 10, cellPadding: 3, textColor: [50, 50, 50] },
    columnStyles: { 1: { halign: "right", fontStyle: "bold" } },
    didParseCell: (cell) => {
      if (cell.column.index === 0) {
        cell.cell.styles.textColor = [70, 70, 70];
      }
    },
  });

  y = (doc as any).lastAutoTable.finalY + 2;

  autoTable(doc, {
    startY: y,
    margin: { left: marginX, right: marginX },
    body: [["GRAND TOTAL", formatCurrency(Number(summary.totalFinal || 0))]],
    theme: "grid",
    styles: {
      fontSize: 12,
      cellPadding: 4,
      fontStyle: "bold",
      fillColor: [250, 240, 225],
      textColor: AMBER,
    },
    columnStyles: { 1: { halign: "right" } },
  });

  const fileName = `Rekapitulasi-RAB-${sanitizeFileName(
    project?.name || "Project"
  )}.pdf`;
  doc.save(fileName);
}
