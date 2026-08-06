import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";

export interface InvoicePdfData {
  id: number;
  nomor: string;
  tanggal: string;
  total: number;
  status: "draft" | "sent" | "paid";
  proyekId: number;
  namaProyek?: string;
  lokasi?: string | null;
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

function formatDate(date: string): string {
  const d = new Date(date);
  if (isNaN(d.getTime())) return date;
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(d);
}

const STATUS_LABEL: Record<string, string> = {
  draft: "Draft",
  sent: "Terkirim",
  paid: "Lunas",
};

export function exportInvoicePdf(inv: InvoicePdfData): void {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const pageWidth = doc.internal.pageSize.getWidth();
  const marginX = 14;
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
  doc.text("Invoice", marginX, 21);

  // Invoice info
  y = 40;
  doc.setTextColor(30, 30, 30);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(12);
  doc.text(inv.namaProyek || `Proyek #${inv.proyekId}`, marginX, y);
  y += 6;
  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  doc.setTextColor(...GRAY);
  if (inv.lokasi) {
    doc.text(String(inv.lokasi), marginX, y);
    y += 5;
  }

  // Detail block
  const rows = [
    ["No. Invoice", inv.nomor],
    ["Tanggal", formatDate(inv.tanggal)],
    ["Status", STATUS_LABEL[inv.status] || inv.status],
    ["Total", formatCurrency(Number(inv.total))],
  ];
  autoTable(doc, {
    startY: y + 4,
    head: [["Keterangan", "Nilai"]],
    body: rows,
    theme: "grid",
    headStyles: { fillColor: AMBER, fontSize: 10 },
    bodyStyles: { fontSize: 10 },
    margin: { left: marginX, right: marginX },
  });

  const finalY = (doc as unknown as { lastAutoTable: { finalY: number } })
    .lastAutoTable.finalY;
  doc.setDrawColor(...GRAY);
  doc.setLineWidth(0.3);
  doc.line(marginX, finalY + 10, pageWidth - marginX, finalY + 10);
  doc.setFontSize(8);
  doc.setTextColor(...GRAY);
  doc.text(
    "Dokumen ini dibuat secara otomatis oleh Halora Land.",
    marginX,
    finalY + 16
  );

  doc.save(`invoice-${inv.nomor}.pdf`);
}
