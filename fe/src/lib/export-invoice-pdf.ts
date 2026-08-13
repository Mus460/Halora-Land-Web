import { jsPDF } from "jspdf";
import { autoTable } from "jspdf-autotable";
import type { Invoice } from "@/types";
import { terbilang } from "./terbilang";

const DARK: [number, number, number] = [31, 41, 55];
const GRAY: [number, number, number] = [107, 114, 128];
const AMBER: [number, number, number] = [217, 119, 6];
const LIGHT: [number, number, number] = [249, 250, 251];

const COMPANY = {
  name: "HALORA LAND",
  addressLine1: "Jl. Adam Malik No. 58, Ruko No.1, Cipadu Jaya, Larangan,",
  addressLine2: "Kota Tangerang, 15155 - Indonesia.",
  contacts: "W: land.halora.id | e: halo@halora.id | wa: +62 811 8622 225",
  tagline: '"Business is about Trust and Value"',
};

const DEFAULT_ACCOUNTS: Array<{ bank: string; number: string; name: string }> = [
  { bank: "BSI", number: "1101101009", name: "Rangga Donyta Putra" },
  { bank: "BCA", number: "3450223963", name: "Rangga Donyta Putra" },
];

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

export function exportInvoicePdf(
  inv: Invoice,
  project?: { name?: string; location?: string | null }
): void {
  const doc = new jsPDF({ unit: "mm", format: "a4" });
  const W = doc.internal.pageSize.getWidth();
  const H = doc.internal.pageSize.getHeight();
  const X = 12;
  const XR = W - 12;

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

  // --- Header: INVOICE title + number ---
  doc.setFont("helvetica", "bold");
  doc.setFontSize(24);
  doc.setTextColor(...DARK);
  doc.text("INVOICE", XR, 22, { align: "right" });
  doc.setFont("helvetica", "normal");
  doc.setFontSize(9);
  doc.setTextColor(...GRAY);
  doc.text(`No. ${inv.number || ""}`, XR, 28, { align: "right" });

  // --- Meta: customer (left) + date/valid date (right) ---
  doc.setFont("helvetica", "bold");
  doc.setFontSize(8.5);
  doc.setTextColor(...GRAY);
  doc.text("Customer :", X, 40);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(11);
  doc.setTextColor(...DARK);
  doc.text(buyerName, X, 46);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(8.5);
  let metaY = 51;
  if (buyerAddress) {
    doc.setTextColor(...GRAY);
    doc.text(String(buyerAddress), X, metaY);
    metaY += 5;
  }
  if (inv.buyerContact) {
    doc.setTextColor(...GRAY);
    doc.text(`Contact : ${inv.buyerContact}`, X, metaY);
    metaY += 5;
  }

  const dateRows: Array<[string, string]> = [
    ["Date", formatDate(inv.date)],
    ["Valid Date", formatDate(inv.dueDate)],
  ];
  if (inv.poNumber) dateRows.push(["No. PO", inv.poNumber]);
  autoTable(doc, {
    startY: 36,
    margin: { left: XR - 78, right: 12 },
    theme: "plain",
    body: dateRows,
    styles: { fontSize: 9, cellPadding: { top: 1, bottom: 1 } },
    columnStyles: {
      0: { cellWidth: 28, fontStyle: "bold", textColor: GRAY },
      1: { cellWidth: 50, textColor: DARK },
    },
  });

  // --- Items table ---
  autoTable(doc, {
    startY: Math.max(metaY + 6, 60),
    margin: { left: X, right: X },
    head: [["No.", "Product", "Descriptions", "Quantity Unit", "Price per each", "Amount"]],
    body: items.map((item, idx) => {
      const descLines = String(item.description || buyerName).split("\n");
      const product = descLines[0];
      const rest = descLines.slice(1).join(" ");
      return [
        String(idx + 1),
        product,
        rest,
        `${formatNumber(Number(item.qty) || 0)} ${item.unit || ""}`.trim(),
        rupiah(Number(item.unitPrice)),
        rupiah((Number(item.qty) || 0) * (Number(item.unitPrice) || 0)),
      ];
    }),
    theme: "grid",
    headStyles: {
      fillColor: DARK,
      textColor: [255, 255, 255],
      fontSize: 9,
      fontStyle: "bold",
    },
    bodyStyles: { fontSize: 9, textColor: DARK },
    alternateRowStyles: { fillColor: LIGHT },
    columnStyles: {
      0: { cellWidth: 8, halign: "center" },
      1: { cellWidth: 45, halign: "left" },
      2: { cellWidth: 62, halign: "left" },
      3: { cellWidth: 26, halign: "right" },
      4: { cellWidth: 23, halign: "right" },
      5: { cellWidth: 24, halign: "right", fontStyle: "bold" },
    },
  });

  // --- Summary block (right aligned) ---
  const summaryRows: string[][] = [["Sub-Total (Rp)", rupiah(subtotal)]];
  if (discount > 0) summaryRows.push(["Diskon (Rp)", `- ${rupiah(discount)}`]);
  summaryRows.push(["Sudah Terbayar", isPaid ? rupiah(paidSoFar) : "Rp 0"]);
  summaryRows.push([
    `Request Pembayaran (${isPaid ? "0" : "100"}%)`,
    isPaid ? "-" : rupiah(remaining),
  ]);
  summaryRows.push([`PPN (${taxRate}%)`, tax > 0 ? rupiah(tax) : "-"]);
  summaryRows.push(["Sisa Pembayaran (Rp)", rupiah(remaining)]);

  autoTable(doc, {
    startY: lastTableY(doc) + 8,
    margin: { left: XR - 84, right: 12 },
    theme: "plain",
    body: summaryRows,
    styles: { fontSize: 9, cellPadding: { top: 1.5, bottom: 1.5 } },
    columnStyles: {
      0: { cellWidth: 52, fontStyle: "bold", textColor: GRAY },
      1: { cellWidth: 32, halign: "right", textColor: DARK },
    },
    didParseCell: (data) => {
      if (data.row.index === summaryRows.length - 1) {
        data.cell.styles.fillColor = [250, 240, 225];
        data.cell.styles.textColor = AMBER;
        data.cell.styles.fontStyle = "bold";
      }
    },
  });

  // --- Terbilang ---
  let y = lastTableY(doc) + 10;
  const terbilangText = `Terbilang : ${terbilang(isPaid ? grandTotal : remaining)}`;
  doc.setFont("helvetica", "normal");
  doc.setFontSize(8.5);
  doc.setTextColor(...GRAY);
  const terbilangLines = doc.splitTextToSize(terbilangText, XR - X);
  doc.text(terbilangLines, X, y);
  y += (terbilangLines as string[]).length * 4 + 8;

  // --- Rekening block ---
  const accounts = inv.paymentBank
    ? [
        {
          bank: inv.paymentBank,
          number: inv.paymentAccountNumber || "-",
          name: inv.paymentAccountName || "-",
        },
      ]
    : DEFAULT_ACCOUNTS;

  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  doc.setTextColor(...DARK);
  doc.text("Pembayaran dapat ditransfer ke :", X, y);
  y += 5;

  const accountWidth = (XR - X) / 2;
  accounts.forEach((acc, i) => {
    const ax = X + (i % 2) * accountWidth;
    const ay = y + Math.floor(i / 2) * 12;
    doc.setFont("helvetica", "bold");
    doc.setFontSize(8.5);
    doc.setTextColor(...DARK);
    doc.text(`${acc.bank}`, ax, ay);
    doc.setFont("helvetica", "normal");
    doc.setTextColor(...GRAY);
    doc.text(`No. Rekening : ${acc.number}`, ax, ay + 5);
    doc.text(`A/N : ${acc.name}`, ax, ay + 10);
  });
  y += Math.ceil(accounts.length / 2) * 12 + 6;

  // --- Accepted by ---
  const sigY = Math.max(y, H - 52);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(9);
  doc.setTextColor(...GRAY);
  doc.text("Accepted by :", XR, sigY, { align: "right" });
  doc.setDrawColor(170, 170, 170);
  doc.setLineWidth(0.3);
  doc.line(XR - 55, sigY + 14, XR, sigY + 14);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(10);
  doc.setTextColor(...DARK);
  doc.text(inv.financeName || "Rangga", XR, sigY + 21, { align: "right" });

  // --- Footer: company block + tagline ---
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

  doc.setFont("helvetica", "italic");
  doc.setFontSize(8);
  doc.setTextColor(...AMBER);
  doc.text(COMPANY.tagline, XR, footerY + 6, { align: "right" });

  doc.setFont("helvetica", "normal");
  doc.setFontSize(7);
  doc.setTextColor(...GRAY);
  doc.text(
    `Dibuat otomatis oleh Halora Land • ${new Date().toLocaleDateString("id-ID")}`,
    X,
    H - 8
  );

  const safeNumber = (inv.number || `INV-${inv.id}`).replace(/[/\\:]/g, "-");
  doc.save(`invoice-${safeNumber}.pdf`);
}