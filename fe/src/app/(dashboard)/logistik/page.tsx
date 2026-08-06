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
import type { Logistik, Realisasi } from "@/types";
import toast from "react-hot-toast";
import { useProject } from "@/contexts/ProjectContext";
import { EmptyProyekState } from "@/components/shared/empty-proyek-state";

export default function LogistikPage() {
  const { currentProyekId: proyekId, proyekList, loading: proyekLoading } = useProject();

  const [data, setData] = useState<Logistik[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    namaMaterial: "",
    satuan: "",
    volume: 0,
    hargaSatuan: 0,
    tanggal: "",
    keterangan: "",
    catatKeuangan: false,
  });

  useEffect(() => {
    if (proyekId) {
      fetchData();
    }
  }, [proyekId]);

  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (proyekList.length === 0) {
    return (
      <EmptyProyekState
        title="Belum Ada Data Logistik"
        description="Buat proyek untuk mencatat pengiriman material dan logistik"
      />
    );
  }

  const fetchData = async () => {
    if (!proyekId) return;

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

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!proyekId) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/proyek/${proyekId}/logistik`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          namaMaterial: form.namaMaterial,
          satuan: form.satuan,
          volume: form.volume,
          hargaSatuan: form.hargaSatuan,
          tanggal: form.tanggal || null,
          keterangan: form.keterangan || null,
          catatKeuangan: form.catatKeuangan,
        }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal menambah material");
      }
      setData((prev) => [result.logistik, ...prev]);
      setShowForm(false);
      setForm({
        namaMaterial: "",
        satuan: "",
        volume: 0,
        hargaSatuan: 0,
        tanggal: "",
        keterangan: "",
        catatKeuangan: false,
      });
      toast.success(
        form.catatKeuangan
          ? "Material ditambahkan, pengeluaran draft dibuat di Keuangan"
          : "Material berhasil ditambahkan"
      );
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setSaving(false);
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
          <p className="text-sm text-blue-700">Total Logistik</p>
          <p className="text-lg font-bold text-blue-900">
            {formatCurrency(totalBiaya)}
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
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Tambah Material</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="namaMaterial">Nama Material *</Label>
              <Input
                id="namaMaterial"
                value={form.namaMaterial}
                onChange={(e) =>
                  setForm({ ...form, namaMaterial: e.target.value })
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
                <Label htmlFor="satuan">Satuan *</Label>
                <Input
                  id="satuan"
                  value={form.satuan}
                  onChange={(e) =>
                    setForm({ ...form, satuan: e.target.value })
                  }
                  placeholder="sak / m3 / unit"
                  required
                />
              </div>
            </div>
            <CurrencyInput
              label="Harga Satuan *"
              value={form.hargaSatuan}
              onChange={(value) => setForm({ ...form, hargaSatuan: value })}
            />
            <div className="space-y-2">
              <Label htmlFor="tanggal">Tanggal</Label>
              <Input
                id="tanggal"
                type="date"
                value={form.tanggal}
                onChange={(e) => setForm({ ...form, tanggal: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="keterangan">Keterangan</Label>
              <Input
                id="keterangan"
                value={form.keterangan}
                onChange={(e) =>
                  setForm({ ...form, keterangan: e.target.value })
                }
                placeholder="Catatan tambahan"
              />
            </div>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={form.catatKeuangan}
                onChange={(e) =>
                  setForm({ ...form, catatKeuangan: e.target.checked })
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
