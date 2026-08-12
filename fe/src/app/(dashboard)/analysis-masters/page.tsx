"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { PageHeader } from "@/components/shared/page-header";
import { SearchInput } from "@/components/shared/search-input";
import { DataTable } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Copy as CopyIcon } from "lucide-react";
import { useDebouncedValue } from "@/hooks/use-debounce";
import { useRouter } from "next/navigation";
import Link from "next/link";
import toast from "react-hot-toast";
import type { AnalysisMaster } from "@/types";

export default function MasterAnalisaPage() {
  const router = useRouter();
  const [data, setData] = useState<AnalysisMaster[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [copyTarget, setCopyTarget] = useState<AnalysisMaster | null>(null);
  const [copyName, setCopyName] = useState("");
  const [copying, setCopying] = useState(false);
  const debouncedSearch = useDebouncedValue(search);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/analysis-masters');
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

  const openCopyDialog = (item: AnalysisMaster) => {
    setCopyTarget(item);
    setCopyName(`Salin - ${item.name}`);
  };

  const handleCopy = async () => {
    if (!copyTarget) return;
    try {
      setCopying(true);
      const response = await fetch(`/api/analysis-masters/${copyTarget.id}/copy`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: copyName.trim() || undefined }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal menyalin analisa");
      }
      toast.success("Analisa disalin");
      setCopyTarget(null);
      router.push(`/analysis-masters/${result.id}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menyalin analisa");
    } finally {
      setCopying(false);
    }
  };

  const filtered = data.filter((item) => {
    const q = debouncedSearch.toLowerCase();
    return (
      item.name.toLowerCase().includes(q) ||
      item.code.toLowerCase().includes(q)
    );
  });

  const columns: ColumnDef<AnalysisMaster>[] = [
    {
      accessorKey: "code",
      header: "Kode",
      cell: ({ row }) => (
        <span className="text-xs text-gray-400 font-mono">{row.original.code}</span>
      ),
    },
    {
      accessorKey: "name",
      header: "Nama",
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <span className="font-medium">{row.original.name}</span>
          {!row.original.isSystem && (
            <Badge variant="outline" className="text-emerald-600">
              Milik saya
            </Badge>
          )}
        </div>
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
      accessorKey: "unit",
      header: "Satuan",
      cell: ({ row }) => (
        <span className="text-gray-600">{row.original.unit || "—"}</span>
      ),
    },
    {
      id: "actions",
      header: "Aksi",
      cell: ({ row }) => (
        <div className="flex items-center gap-1">
          <Link href={`/analysis-masters/${row.original.id}`}>
            <Button variant="ghost" size="sm" className="text-amber-600 hover:text-amber-700">
              Detail
            </Button>
          </Link>
          <Button
            variant="ghost"
            size="sm"
            className="text-blue-600 hover:text-blue-700"
            onClick={() => openCopyDialog(row.original)}
          >
            <CopyIcon className="w-3.5 h-3.5 mr-1" /> Salin
          </Button>
        </div>
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

      <Dialog open={copyTarget !== null} onOpenChange={(open) => !open && setCopyTarget(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Salin Analisa</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="copy-name">Nama baru</Label>
              <Input
                id="copy-name"
                value={copyName}
                onChange={(e) => setCopyName(e.target.value)}
                placeholder="Nama analisa salinan"
              />
            </div>
            <p className="text-sm text-gray-500">
              Salinan adalah milik Anda dan dapat dimodifikasi, terpisah dari data AHSP sistem.
            </p>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setCopyTarget(null)}
            >
              Batal
            </Button>
            <Button type="button" onClick={handleCopy} disabled={copying || !copyName.trim()}>
              {copying ? "Menyalin..." : "Salin"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
