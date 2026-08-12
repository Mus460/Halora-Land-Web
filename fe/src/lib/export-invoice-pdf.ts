import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";
import type { Invoice } from "@/types";

const AMBER: [number, number, number] = [217, 119, 6];
const DARK: [number, number, number] = [30, 30, 30];
const GRAY: [number, number, number] = [107, 114, 128];
const LIGHT: [number, number, number] = [249, 250, 251];

const STATUS_LABEL: Record<string, string> = {
  draft: "Draft",
  sent: "Terkirim",
  paid: "Lunas",
};

function rupiah(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
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

export function exportInvoicePdf(
  inv: Invoice,
  project?: { name?: string; location?: string | null }
): void {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const W = doc.internal.pageSize.getWidth();
  const H = doc.internal.pageSize.getHeight();
  const X = 14;
  const XR = W - 14;

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
  const subtotal = items.reduce((sum, i) => sum + (Number(i.qty) || 0) * (Number(i.unitPrice) || 0), 0);
  const discount = Number(inv.discount) || 0;
  const taxRate = Number(inv.taxRate) || 0;
  const tax = (Math.max(subtotal - discount, 0) * taxRate) / 100;
  const grandTotal = Number(inv.total) || Math.max(subtotal - discount + tax, 0);

  doc.setFillColor(...AMBER);
  doc.rect(0, 0, W, 30, "F");
  doc.setTextColor(255, 255, 255);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.text("Halora Land", X, 13);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(255, 235, 214);
  doc.text("Dokumen tagihan resmi", X, 18);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(20);
  doc.text("INVOICE", XR, 13, { align: "right" });
  doc.setFontSize(9);
  doc.setTextColor(255, 235, 214);
  doc.text(inv.number, XR, 19, { align: "right" });
  doc.setTextColor(...AMBER);
  doc.setFillColor(255, 255, 255);
  doc.roundedRect(XR - 26, 21, 26, 6, 1, 1, "F");
  doc.setFontSize(7);
  doc.setFont("helvetica", "bold");
  doc.text(STATUS_LABEL[inv.status] || inv.status, XR - 13, 25.5, {
    align: "center",
  });

  const metaRows: Array<[string, string]> = [
    ["No. Invoice", inv.number],
    ["Tanggal Penerbitan", formatDate(inv.date)],
  ];
  if (inv.dueDate) metaRows.push(["Jatuh Tempo", formatDate(inv.dueDate)]);
  if (inv.poNumber) metaRows.push(["No. PO", inv.poNumber]);

  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  doc.setTextColor(...DARK);
  doc.text("Ditagihkan Kepada (Bill To)", X, 40);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(9);
  doc.setTextColor(...DARK);
  doc.text(buyerName, X, 46);
  doc.setTextColor(...GRAY);
  let buyerY = 51;
  if (buyerAddress) {
    doc.text(String(buyerAddress), X, buyerY);
    buyerY += 5;
  }
  if (inv.buyerContact) {
    doc.text(`PIC: ${inv.buyerContact}`, X, buyerY);
  }

  autoTable(doc, {
    startY: 38,
    margin: { left: XR - 76, right: 14 },
    theme: "plain",
    body: metaRows,
    styles: { fontSize: 9, cellPadding: { top: 1, bottom: 1 } },
    columnStyles: {
      0: { cellWidth: 34, fontStyle: "bold", textColor: GRAY },
      1: { cellWidth: 42, textColor: DARK },
    },
  });

  autoTable(doc, {
    startY: Math.max(lastTableY(doc), 64) + 6,
    head: [["No.", "Deskripsi", "Qty", "Satuan", "Harga Satuan", "Total"]],
    body: items.map((item, idx) => [
      String(idx + 1),
      item.description,
      String(item.qty),
      item.unit,
      rupiah(Number(item.unitPrice)),
      rupiah((Number(item.qty) || 0) * (Number(item.unitPrice) || 0)),
    ]),
    theme: "grid",
    headStyles: { fillColor: AMBER, textColor: 255, fontSize: 9 },
    bodyStyles: { fontSize: 9, textColor: DARK },
    alternateRowStyles: { fillColor: LIGHT },
    columnStyles: {
      0: { cellWidth: 10, halign: "center" },
      2: { cellWidth: 14, halign: "center" },
      3: { cellWidth: 18, halign: "center" },
      4: { cellWidth: 32, halign: "right" },
      5: { cellWidth: 32, halign: "right" },
    },
    margin: { left: X, right: X },
  });

  const summaryRows: string[][] = [["Subtotal", rupiah(subtotal)]];
  if (discount > 0) summaryRows.push(["Diskon", `- ${rupiah(discount)}`]);
  if (taxRate > 0) summaryRows.push([`PPN (${taxRate}%)`, rupiah(tax)]);
  summaryRows.push(["Grand Total", rupiah(grandTotal)]);

  autoTable(doc, {
    startY: lastTableY(doc) + 6,
    margin: { left: XR - 66, right: 14 },
    theme: "plain",
    body: summaryRows,
    styles: { fontSize: 9, cellPadding: { top: 1, bottom: 1 } },
    columnStyles: {
      0: { cellWidth: 34, fontStyle: "bold", textColor: GRAY },
      1: { cellWidth: 32, halign: "right", textColor: DARK },
    },
    didParseCell: (data) => {
      if (data.row.index === summaryRows.length - 1) {
        data.cell.styles.fillColor = AMBER;
        data.cell.styles.textColor = 255;
        data.cell.styles.fontStyle = "bold";
      }
    },
  });

  let y = lastTableY(doc) + 16;
  const hasPayment =
    inv.paymentBank || inv.paymentAccountNumber || inv.paymentAccountName;
  const notes = inv.notes ? inv.notes.split("\n").filter((n) => n.trim()) : [];

  if (hasPayment || notes.length > 0) {
    doc.setFont("helvetica", "bold");
    doc.setFontSize(10);
    doc.setTextColor(...DARK);
    if (hasPayment) {
      doc.text("Instruksi Pembayaran", X, y);
      doc.setFont("helvetica", "normal");
      doc.setFontSize(9);
      doc.setTextColor(...GRAY);
      const payRows: Array<[string, string]> = [];
      if (inv.paymentBank) payRows.push(["Bank", inv.paymentBank]);
      if (inv.paymentAccountNumber) payRows.push(["No. Rekening", inv.paymentAccountNumber]);
      if (inv.paymentAccountName) payRows.push(["Atas Nama", inv.paymentAccountName]);
      autoTable(doc, {
        startY: y + 3,
        margin: { left: X, right: W / 2 + 6 },
        theme: "plain",
        body: payRows,
        styles: { fontSize: 9, cellPadding: { top: 0.8, bottom: 0.8 } },
        columnStyles: {
          0: { cellWidth: 28, fontStyle: "bold", textColor: GRAY },
          1: { textColor: DARK },
        },
      });
      y = lastTableY(doc) + 8;
    }
    if (notes.length > 0) {
      doc.setFont("helvetica", "bold");
      doc.setFontSize(10);
      doc.setTextColor(...DARK);
      doc.text("Catatan", X, y);
      doc.setFont("helvetica", "normal");
      doc.setFontSize(9);
      doc.setTextColor(...GRAY);
      notes.forEach((note, i) => {
        doc.text(`• ${note}`, X, y + 5 + i * 5);
      });
    }
  }

  const sigY = H - 40;
  doc.setTextColor(...GRAY);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(9);
  doc.text("Hormat kami,", XR, sigY, { align: "right" });
  doc.setDrawColor(170, 170, 170);
  doc.setLineWidth(0.3);
  doc.line(XR - 50, sigY + 12, XR, sigY + 12);
  doc.setFont("helvetica", "bold");
  doc.setTextColor(...DARK);
  doc.text(inv.financeName || "Bagian Keuangan", XR, sigY + 19, {
    align: "right",
  });

  doc.setDrawColor(...GRAY);
  doc.setLineWidth(0.3);
  doc.line(X, H - 18, XR, H - 18);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...GRAY);
  doc.text(
    `Dibuat otomatis oleh Halora Land • ${new Date().toLocaleDateString("id-ID")}`,
    X,
    H - 13
  );

  const safeNumber = (inv.number || `INV-${inv.id}`).replace(/[/\\:]/g, "-");
  doc.save(`invoice-${safeNumber}.pdf`);
}
