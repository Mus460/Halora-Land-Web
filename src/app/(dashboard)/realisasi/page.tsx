"use client";

import { useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Wallet, TrendingDown, TrendingUp, DollarSign } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { StatCard } from "@/components/shared/stat-card";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import { getRealisasi, getAllPekerjaan } from "@/mock";
import type { Realisasi } from "@/types";

export default function RealisasiPage() {
  const [data] = useState<Realisasi[]>(getRealisasi(1));
  const pekerjaan = getAllPekerjaan(1);

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

  const totalRAB = pekerjaan.reduce((sum, p) => sum + p.totalBiaya, 0);
  const totalPengeluaran = data.reduce((sum, r) => sum + r.jumlah, 0);
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
