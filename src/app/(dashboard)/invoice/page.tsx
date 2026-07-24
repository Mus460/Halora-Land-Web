"use client";

import { ColumnDef } from "@tanstack/react-table";
import { Plus, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import { getInvoiceList } from "@/mock";
import { INVOICE_STATUS } from "@/lib/constants";
import type { Invoice } from "@/types";
import { useState } from "react";

export default function InvoicePage() {
  const [data] = useState<Invoice[]>(getInvoiceList(1));

  const columns: ColumnDef<Invoice>[] = [
    {
      accessorKey: "nomor",
      header: "No. Invoice",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.nomor}</span>
      ),
    },
    {
      accessorKey: "tanggal",
      header: "Tanggal",
      cell: ({ row }) => formatDateShort(row.original.tanggal),
    },
    {
      accessorKey: "total",
      header: "Total",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.total)}
        </span>
      ),
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = INVOICE_STATUS.find(
          (s) => s.value === row.original.status
        );
        return (
          <Badge
            variant={
              row.original.status === "paid"
                ? "default"
                : row.original.status === "sent"
                ? "secondary"
                : "outline"
            }
          >
            {status?.label || row.original.status}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "Aksi",
      cell: () => (
        <Button variant="ghost" size="sm">
          <FileText className="w-4 h-4 mr-1" />
          Cetak
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Invoice"
        description="Kelola invoice untuk proyek"
        actions={
          <Button className="bg-amber-500 hover:bg-amber-600">
            <Plus className="w-4 h-4 mr-2" />
            Buat Invoice
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data}
        emptyTitle="Belum ada invoice"
        emptyDescription="Buat invoice untuk tagihan proyek"
      />
    </div>
  );
}
