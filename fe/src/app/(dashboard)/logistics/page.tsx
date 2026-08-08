"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { CurrencyInput } from "@/components/shared/currency-input";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import type { Logistics, Transaction } from "@/types";
import toast from "react-hot-toast";
import { useProject } from "@/contexts/ProjectContext";
import { EmptyProjectState } from "@/components/shared/empty-project-state";

export default function LogistikPage() {
  const { currentProjectId: projectId, projectList, loading: proyekLoading } = useProject();

  const [data, setData] = useState<Logistics[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    materialName: "",
    unit: "",
    volume: 0,
    unitPrice: 0,
    date: "",
    description: "",
    recordExpense: false,
  });

  const fetchData = async () => {
    if (!projectId) return;

    try {
      setLoading(true);
      const response = await fetch(`/api/projects/${projectId}/logistics`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.logistics || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (projectId) {
      fetchData();
    }
  }, [projectId]);

  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (projectList.length === 0) {
    return (
      <EmptyProjectState
        title="Belum Ada Data Logistics"
        description="Buat proyek untuk mencatat pengiriman material dan logistik"
      />
    );
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!projectId) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/projects/${projectId}/logistics`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          materialName: form.materialName,
          unit: form.unit,
          volume: form.volume,
          unitPrice: form.unitPrice,
          date: form.date || null,
          description: form.description || null,
          recordExpense: form.recordExpense,
        }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal menambah material");
      }
      setData((prev) => [result.logistics, ...prev]);
      setShowForm(false);
      setForm({
        materialName: "",
        unit: "",
        volume: 0,
        unitPrice: 0,
        date: "",
        description: "",
        recordExpense: false,
      });
      toast.success(
        form.recordExpense
          ? "Material ditambahkan, pengeluaran draft dibuat di Keuangan"
          : "Material berhasil ditambahkan"
      );
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setSaving(false);
    }
  };

  const columns: ColumnDef<Logistics>[] = [
    {
      accessorKey: "materialName",
      header: "Material",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.materialName}</span>
      ),
    },
    {
      accessorKey: "volume",
      header: "Volume",
      cell: ({ row }) => (
        <span>
          {row.original.volume} {row.original.unit}
        </span>
      ),
    },
    {
      accessorKey: "unitPrice",
      header: "Harga Satuan",
      cell: ({ row }) => formatCurrency(row.original.unitPrice),
    },
    {
      accessorKey: "totalCost",
      header: "Total",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.totalCost)}
        </span>
      ),
    },
    {
      accessorKey: "date",
      header: "Tanggal",
      cell: ({ row }) =>
        row.original.date
          ? formatDateShort(row.original.date)
          : "-",
    },
    {
      accessorKey: "description",
      header: "Keterangan",
      cell: ({ row }) => (
        <span className="text-gray-500 text-sm truncate max-w-[200px] block">
          {row.original.description || "-"}
        </span>
      ),
    },
  ];

  const totalCost = data.reduce((sum, item) => sum + item.totalCost, 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Logistics"
        description="Tracking material dan pengiriman"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => setShowForm(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            Tambah Material
          </Button>
        }
      />

      <div className="flex items-center gap-4 p-4 bg-blue-50 rounded-lg border border-blue-200">
        <Package className="w-5 h-5 text-blue-600" />
        <div>
          <p className="text-sm text-blue-700">Total Logistics</p>
          <p className="text-lg font-bold text-blue-900">
            {formatCurrency(totalCost)}
          </p>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={data}
        loading={loading}
        emptyTitle="Belum ada data logistik"
        emptyDescription="Tambahkan material yang sudah dikirim ke lokasi"
      />

      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Tambah Material</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="materialName">Nama Material *</Label>
              <Input
                id="materialName"
                value={form.materialName}
                onChange={(e) =>
                  setForm({ ...form, materialName: e.target.value })
                }
                placeholder="Contoh: Semen 50kg"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="volume">Volume *</Label>
                <Input
                  id="volume"
                  type="number"
                  step="any"
                  min="0"
                  value={form.volume || ""}
                  onChange={(e) =>
                    setForm({ ...form, volume: Number(e.target.value) || 0 })
                  }
                  placeholder="0"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="unit">Satuan *</Label>
                <Input
                  id="unit"
                  value={form.unit}
                  onChange={(e) =>
                    setForm({ ...form, unit: e.target.value })
                  }
                  placeholder="sak / m3 / unit"
                  required
                />
              </div>
            </div>
            <CurrencyInput
              label="Harga Satuan *"
              value={form.unitPrice}
              onChange={(value) => setForm({ ...form, unitPrice: value })}
            />
            <div className="space-y-2">
              <Label htmlFor="date">Tanggal</Label>
              <Input
                id="date"
                type="date"
                value={form.date}
                onChange={(e) => setForm({ ...form, date: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Keterangan</Label>
              <Input
                id="description"
                value={form.description}
                onChange={(e) =>
                  setForm({ ...form, description: e.target.value })
                }
                placeholder="Catatan tambahan"
              />
            </div>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={form.recordExpense}
                onChange={(e) =>
                  setForm({ ...form, recordExpense: e.target.checked })
                }
                className="h-4 w-4 rounded border-gray-300 accent-amber-500"
              />
              <span>
                Catat ke Keuangan{" "}
                <span className="text-xs text-gray-500">
                  (membuat entri pengeluaran draft)
                </span>
              </span>
            </label>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowForm(false)}
              >
                Batal
              </Button>
              <Button
                type="submit"
                disabled={saving}
                className="bg-amber-500 hover:bg-amber-600"
              >
                {saving ? "Menyimpan..." : "Simpan"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
