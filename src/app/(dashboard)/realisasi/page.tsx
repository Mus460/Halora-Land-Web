"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Wallet, TrendingDown, TrendingUp, DollarSign } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { StatCard } from "@/components/shared/stat-card";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import type { Realisasi } from "@/types";
import toast from "react-hot-toast";

export default function RealisasiPage() {
  const [data, setData] = useState<Realisasi[]>([]);
  const [loading, setLoading] = useState(true);
  const proyekId = 1; // TODO: get from context/URL

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/proyek/${proyekId}/realisasi`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.realisasi || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnDef<Realisasi>[] = [
    {
      accessorKey: "tanggal",
      header: "Tanggal",
      cell: ({ row }) => formatDateShort(row.original.tanggal),
    },
    {
      accessorKey: "kategori",
      header: "Kategori",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.kategori}</span>
      ),
    },
    {
      accessorKey: "jumlah",
      header: "Jumlah",
      cell: ({ row }) => (
        <span className="font-semibold text-red-600">
          - {formatCurrency(row.original.jumlah)}
        </span>
      ),
    },
    {
      accessorKey: "keterangan",
      header: "Keterangan",
      cell: ({ row }) => (
        <span className="text-gray-500 text-sm">
          {row.original.keterangan || "-"}
        </span>
      ),
    },
  ];

  const totalPengeluaran = data.reduce((sum, r) => sum + r.jumlah, 0);
  // TODO: fetch totalRAB from API or pass via props
  const totalRAB = 0; // Placeholder
  const selisih = totalRAB - totalPengeluaran;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Keuangan"
        description="Tracking realisasi pengeluaran proyek"
        actions={
          <Button className="bg-amber-500 hover:bg-amber-600">
            <Plus className="w-4 h-4 mr-2" />
            Tambah Transaksi
          </Button>
        }
      />

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          title="Total RAB"
          value={formatCurrency(totalRAB)}
          icon={<DollarSign className="w-6 h-6" />}
        />
        <StatCard
          title="Total Pengeluaran"
          value={formatCurrency(totalPengeluaran)}
          icon={<TrendingDown className="w-6 h-6" />}
        />
        <StatCard
          title="Selisih"
          value={formatCurrency(selisih)}
          icon={<TrendingUp className="w-6 h-6" />}
          trend={{
            value: totalRAB > 0 ? Math.round((selisih / totalRAB) * 100) : 0,
            isPositive: selisih >= 0,
          }}
        />
      </div>

      <DataTable
        columns={columns}
        data={data}
        emptyTitle="Belum ada transaksi"
        emptyDescription="Catat pengeluaran proyek untuk tracking keuangan"
      />
    </div>
  );
}
