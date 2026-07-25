"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import { INVOICE_STATUS } from "@/lib/constants";
import type { Invoice } from "@/types";
import toast from "react-hot-toast";
import { useProject } from "@/contexts/ProjectContext";

export default function InvoicePage() {
  const { currentProyekId: proyekId } = useProject();
  const [data, setData] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (proyekId) {
      fetchData();
    }
  }, [proyekId]);

  const fetchData = async () => {
    if (!proyekId) return;
    
    try {
      setLoading(true);
      const response = await fetch(`/api/proyek/${proyekId}/invoice`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.invoices || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

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
