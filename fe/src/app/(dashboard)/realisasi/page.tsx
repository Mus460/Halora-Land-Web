"use client";
import { useProject } from "@/contexts/ProjectContext";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Wallet, TrendingDown, TrendingUp, DollarSign, CheckCircle2, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { StatCard } from "@/components/shared/stat-card";
import { CurrencyInput } from "@/components/shared/currency-input";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import type { Realisasi } from "@/types";
import toast from "react-hot-toast";
import { EmptyProyekState } from "@/components/shared/empty-proyek-state";

const REALISASI_KATEGORI = [
  "Material",
  "Upah",
  "Alat",
  "Transport",
  "Lainnya",
];

const JENIS_LABEL: Record<string, string> = {
  pengeluaran: "Pengeluaran",
  pemasukan: "Pemasukan",
};

const STATUS_LABEL: Record<string, string> = {
  draft: "Draft",
  approved: "Disetujui",
  reverted: "Dibatalkan",
};

export default function RealisasiPage() {
  const { currentProyekId: proyekId, proyekList, loading: proyekLoading } = useProject();

  const [data, setData] = useState<Realisasi[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalRAB, setTotalRAB] = useState(0);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [updatingId, setUpdatingId] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Realisasi | null>(null);
  const [form, setForm] = useState({
    tanggal: "",
    kategori: REALISASI_KATEGORI[0],
    jumlah: 0,
    keterangan: "",
    jenis: "pengeluaran" as "pengeluaran" | "pemasukan",
  });

  const fetchData = async () => {
    if (!proyekId) return;
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

  const fetchTotalRAB = async () => {
    if (!proyekId) return;
    try {
      const response = await fetch(`/api/proyek/${proyekId}/rekap`);
      if (!response.ok) return;
      const result = await response.json();
      setTotalRAB(Number(result?.summary?.totalAkhir) || 0);
    } catch (error) {
      console.error('Fetch total RAB error:', error);
    }
  };

  useEffect(() => {
    if (proyekId) {
      fetchData();
      fetchTotalRAB();
    }
  }, [proyekId]);

  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (proyekList.length === 0) {
    return (
      <EmptyProyekState
        title="Belum Ada Data Keuangan"
        description="Buat proyek untuk mencatat realisasi keuangan dan pengeluaran"
      />
    );
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!proyekId) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/proyek/${proyekId}/realisasi`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          tanggal: form.tanggal,
          kategori: form.kategori,
          jumlah: form.jumlah,
          keterangan: form.keterangan || null,
          jenis: form.jenis,
        }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal menambah transaksi");
      }
      setData((prev) => [result.realisasi, ...prev]);
      setShowForm(false);
      setForm({
        tanggal: "",
        kategori: REALISASI_KATEGORI[0],
        jumlah: 0,
        keterangan: "",
        jenis: "pengeluaran",
      });
      toast.success("Transaksi berhasil ditambahkan");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setSaving(false);
    }
  };

  const handleApprove = async (id: number) => {
    if (!proyekId) return;
    try {
      setUpdatingId(id);
      const response = await fetch(`/api/proyek/${proyekId}/realisasi/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: "approved" }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal menyetujui transaksi");
      }
      setData((prev) =>
        prev.map((r) => (r.id === id ? result.realisasi : r))
      );
      toast.success("Transaksi disetujui");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setUpdatingId(null);
    }
  };

  const handleDelete = async (id: number) => {
    if (!proyekId) return;
    try {
      setUpdatingId(id);
      const response = await fetch(`/api/proyek/${proyekId}/realisasi/${id}`, {
        method: "DELETE",
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal menghapus transaksi");
      }
      setData((prev) => prev.filter((r) => r.id !== id));
      setDeleteTarget(null);
      toast.success("Transaksi dihapus");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setUpdatingId(null);
    }
  };
  
  const columns: ColumnDef<Realisasi>[] = [
    {
      accessorKey: "tanggal",
      header: "Tanggal",
      cell: ({ row }) => formatDateShort(row.original.tanggal),
    },
    {
      accessorKey: "jenis",
      header: "Jenis",
      cell: ({ row }) => (
        <Badge
          variant={row.original.jenis === "pemasukan" ? "default" : "outline"}
          className={
            row.original.jenis === "pemasukan"
              ? "bg-green-100 text-green-700"
              : "bg-red-50 text-red-700"
          }
        >
          {JENIS_LABEL[row.original.jenis] || row.original.jenis}
        </Badge>
      ),
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
        <span
          className={`font-semibold ${
            row.original.jenis === "pemasukan"
              ? "text-green-600"
              : "text-red-600"
          }`}
        >
          {row.original.jenis === "pemasukan" ? "+ " : "- "}
          {formatCurrency(row.original.jumlah)}
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
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = row.original.status;
        const variant =
          status === "approved"
            ? "default"
            : status === "reverted"
            ? "outline"
            : "secondary";
        return (
          <Badge variant={variant}>
            {STATUS_LABEL[status] || status}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "Aksi",
      cell: ({ row }) => {
        if (row.original.status !== "draft") return null;
        return (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              disabled={updatingId === row.original.id}
              onClick={() => handleApprove(row.original.id)}
            >
              <CheckCircle2 className="w-4 h-4 mr-1" />
              Setujui
            </Button>
            <Button
              variant="ghost"
              size="sm"
              disabled={updatingId === row.original.id}
              onClick={() => setDeleteTarget(row.original)}
            >
              <Trash2 className="w-4 h-4 mr-1 text-red-600" />
              Hapus
            </Button>
          </div>
        );
      },
    },
  ];

  const approved = data.filter((r) => r.status === "approved");
  const totalPemasukan = approved
    .filter((r) => r.jenis === "pemasukan")
    .reduce((sum, r) => sum + r.jumlah, 0);
  const totalPengeluaran = approved
    .filter((r) => r.jenis === "pengeluaran")
    .reduce((sum, r) => sum + r.jumlah, 0);
  const selisih = totalRAB + totalPemasukan - totalPengeluaran;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Keuangan"
        description="Tracking realisasi pengeluaran proyek"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => setShowForm(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            Tambah Transaksi
          </Button>
        }
      />

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total RAB"
          value={formatCurrency(totalRAB)}
          icon={<DollarSign className="w-6 h-6" />}
        />
        <StatCard
          title="Total Pemasukan"
          value={formatCurrency(totalPemasukan)}
          icon={<TrendingUp className="w-6 h-6" />}
        />
        <StatCard
          title="Total Pengeluaran"
          value={formatCurrency(totalPengeluaran)}
          icon={<TrendingDown className="w-6 h-6" />}
        />
        <StatCard
          title="Selisih"
          value={formatCurrency(selisih)}
          icon={<Wallet className="w-6 h-6" />}
          trend={{
            value: totalRAB > 0 ? Math.round((selisih / totalRAB) * 100) : 0,
            isPositive: selisih >= 0,
          }}
        />
      </div>

      <DataTable
        columns={columns}
        data={data}
        loading={loading}
        emptyTitle="Belum ada transaksi"
        emptyDescription="Catat pengeluaran proyek untuk tracking keuangan"
      />

      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Tambah Transaksi</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-2">
              <Label>Jenis</Label>
              <Select
                value={form.jenis}
                onValueChange={(value) =>
                  setForm({
                    ...form,
                    jenis: (value ?? "pengeluaran") as "pengeluaran" | "pemasukan",
                  })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pengeluaran">Pengeluaran</SelectItem>
                  <SelectItem value="pemasukan">Pemasukan</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="tanggal">Tanggal *</Label>
              <Input
                id="tanggal"
                type="date"
                value={form.tanggal}
                onChange={(e) => setForm({ ...form, tanggal: e.target.value })}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>Kategori</Label>
              <Select
                value={form.kategori}
                onValueChange={(value) =>
                  setForm({ ...form, kategori: value ?? form.kategori })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {REALISASI_KATEGORI.map((k) => (
                    <SelectItem key={k} value={k}>
                      {k}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <CurrencyInput
              label="Jumlah *"
              value={form.jumlah}
              onChange={(value) => setForm({ ...form, jumlah: value })}
            />
            <div className="space-y-2">
              <Label htmlFor="keterangan">Keterangan</Label>
              <Input
                id="keterangan"
                value={form.keterangan}
                onChange={(e) =>
                  setForm({ ...form, keterangan: e.target.value })
                }
                placeholder="Deskripsi pengeluaran"
              />
            </div>
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

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => !o && setDeleteTarget(null)}
        title="Hapus Transaksi"
        description={`Hapus transaksi ${deleteTarget?.keterangan || deleteTarget?.kategori || ""} sebesar ${formatCurrency(
          deleteTarget?.jumlah || 0
        )}?`}
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={() => deleteTarget && handleDelete(deleteTarget.id)}
      />
    </div>
  );
}
