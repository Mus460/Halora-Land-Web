"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import type { Logistik } from "@/types";
import toast from "react-hot-toast";

export default function LogistikPage() {
  const [data, setData] = useState<Logistik[]>([]);
  const [loading, setLoading] = useState(true);
  const proyekId = 1; // TODO: get from context/URL

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/proyek/${proyekId}/logistik`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.logistik || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnDef<Logistik>[] = [
    {
      accessorKey: "namaMaterial",
      header: "Material",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.namaMaterial}</span>
      ),
    },
    {
      accessorKey: "volume",
      header: "Volume",
      cell: ({ row }) => (
        <span>
          {row.original.volume} {row.original.satuan}
        </span>
      ),
    },
    {
      accessorKey: "hargaSatuan",
      header: "Harga Satuan",
      cell: ({ row }) => formatCurrency(row.original.hargaSatuan),
    },
    {
      accessorKey: "totalBiaya",
      header: "Total",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.totalBiaya)}
        </span>
      ),
    },
    {
      accessorKey: "tanggal",
      header: "Tanggal",
      cell: ({ row }) =>
        row.original.tanggal
          ? formatDateShort(row.original.tanggal)
          : "-",
    },
    {
      accessorKey: "keterangan",
      header: "Keterangan",
      cell: ({ row }) => (
        <span className="text-gray-500 text-sm truncate max-w-[200px] block">
          {row.original.keterangan || "-"}
        </span>
      ),
    },
  ];

  const totalBiaya = data.reduce((sum, item) => sum + item.totalBiaya, 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Logistik"
        description="Tracking material dan pengiriman"
        actions={
          <Button className="bg-amber-500 hover:bg-amber-600">
            <Plus className="w-4 h-4 mr-2" />
            Tambah Material
          </Button>
        }
      />

      <div className="flex items-center gap-4 p-4 bg-blue-50 rounded-lg border border-blue-200">
        <Package className="w-5 h-5 text-blue-600" />
        <div>
          <p className="text-sm text-blue-700">Total Logistik</p>
          <p className="text-lg font-bold text-blue-900">
            {formatCurrency(totalBiaya)}
          </p>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={data}
        emptyTitle="Belum ada data logistik"
        emptyDescription="Tambahkan material yang sudah dikirim ke lokasi"
      />
    </div>
  );
}
