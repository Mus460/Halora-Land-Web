"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { PageHeader } from "@/components/shared/page-header";
import { SearchInput } from "@/components/shared/search-input";
import { DataTable } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useDebouncedValue } from "@/hooks/use-debounce";
import Link from "next/link";
import toast from "react-hot-toast";
import type { MasterAnalisa } from "@/types";

export default function MasterAnalisaPage() {
  const [data, setData] = useState<MasterAnalisa[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/master-analisa');
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.data || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  const filtered = data.filter((item) => {
    const q = debouncedSearch.toLowerCase();
    return (
      item.nama.toLowerCase().includes(q) ||
      item.kode.toLowerCase().includes(q)
    );
  });

  const columns: ColumnDef<MasterAnalisa>[] = [
    {
      accessorKey: "kode",
      header: "Kode",
      cell: ({ row }) => (
        <span className="text-xs text-gray-400 font-mono">{row.original.kode}</span>
      ),
    },
    {
      accessorKey: "nama",
      header: "Nama",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.nama}</span>
      ),
    },
    {
      accessorKey: "level",
      header: "Level",
      cell: ({ row }) => (
        <Badge variant="outline">Level {row.original.level}</Badge>
      ),
    },
    {
      accessorKey: "satuan",
      header: "Satuan",
      cell: ({ row }) => (
        <span className="text-gray-600">{row.original.satuan || "—"}</span>
      ),
    },
    {
      id: "actions",
      header: "Aksi",
      cell: ({ row }) => (
        <Link href={`/master-analisa/${row.original.id}`}>
          <Button variant="ghost" size="sm" className="text-amber-600 hover:text-amber-700">
            Detail
          </Button>
        </Link>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Master Analisa (AHSP)"
        description="Database Analisa Harga Satuan Pekerjaan PUPR 2026"
      />

      <div className="flex items-center gap-4">
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Cari analisa..."
          className="max-w-sm"
        />
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        emptyTitle={loading ? "Memuat data..." : "Belum ada data analisa"}
        emptyDescription={loading ? "" : "Tidak ditemukan data analisa"}
      />
    </div>
  );
}
