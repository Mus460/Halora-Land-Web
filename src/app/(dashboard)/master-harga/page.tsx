"use client";

import { useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Pencil, Trash2, RotateCcw, Wand2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { SearchInput } from "@/components/shared/search-input";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { formatCurrency } from "@/lib/utils";
import { getMasterHarga } from "@/mock";
import { TIPE_KOMPONEN } from "@/lib/constants";
import toast from "react-hot-toast";
import type { MasterHarga, TipeKomponen } from "@/types";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CurrencyInput } from "@/components/shared/currency-input";

export default function MasterHargaPage() {
  const [data, setData] = useState<MasterHarga[]>(getMasterHarga());
  const [search, setSearch] = useState("");
  const [filterKategori, setFilterKategori] = useState<string>("all");
  const [showForm, setShowForm] = useState(false);
  const [editItem, setEditItem] = useState<MasterHarga | null>(null);
  const [showDelete, setShowDelete] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const filtered = data.filter((item) => {
    const matchSearch = item.nama
      .toLowerCase()
      .includes(search.toLowerCase());
    const matchKategori =
      filterKategori === "all" || item.kategori === filterKategori;
    return matchSearch && matchKategori;
  });

  const columns: ColumnDef<MasterHarga>[] = [
    {
      accessorKey: "nama",
      header: "Nama",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.nama}</span>
      ),
    },
    {
      accessorKey: "satuan",
      header: "Satuan",
      cell: ({ row }) => (
        <span className="text-gray-600">{row.original.satuan}</span>
      ),
    },
    {
      accessorKey: "harga",
      header: "Harga (Rp)",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.harga)}
        </span>
      ),
    },
    {
      accessorKey: "kategori",
      header: "Kategori",
      cell: ({ row }) => {
        const colors: Record<string, string> = {
          material: "bg-blue-100 text-blue-700",
          upah: "bg-green-100 text-green-700",
          alat: "bg-purple-100 text-purple-700",
        };
        return (
          <Badge
            variant="outline"
            className={colors[row.original.kategori] || ""}
          >
            {row.original.kategori}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "Aksi",
      cell: ({ row }) => (
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => {
              setEditItem(row.original);
              setShowForm(true);
            }}
          >
            <Pencil className="w-4 h-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-red-500 hover:text-red-600"
            onClick={() => {
              setDeleteId(row.original.id);
              setShowDelete(true);
            }}
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      ),
    },
  ];

  const handleAutoFill = () => {
    setData((prev) =>
      prev.map((item) =>
        item.harga === 0
          ? { ...item, harga: Math.floor(Math.random() * 500000) + 10000 }
          : item
      )
    );
    toast.success("Harga berhasil diisi otomatis");
  };

  const handleReset = () => {
    setData((prev) => prev.map((item) => ({ ...item, harga: 0 })));
    toast.success("Semua harga direset ke 0");
  };

  const handleDelete = () => {
    if (deleteId) {
      setData((prev) => prev.filter((item) => item.id !== deleteId));
      toast.success("Data berhasil dihapus");
    }
    setShowDelete(false);
    setDeleteId(null);
  };

  const handleSubmit = (formData: Partial<MasterHarga>) => {
    if (editItem) {
      setData((prev) =>
        prev.map((item) =>
          item.id === editItem.id
            ? { ...item, ...formData, updatedAt: new Date().toISOString() }
            : item
        )
      );
      toast.success("Data berhasil diupdate");
    } else {
      const newItem: MasterHarga = {
        id: Date.now(),
        nama: formData.nama || "",
        satuan: formData.satuan || "",
        harga: formData.harga || 0,
        kategori: formData.kategori || "material",
        isGlobal: false,
        userId: 2,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      setData((prev) => [newItem, ...prev]);
      toast.success("Data berhasil ditambahkan");
    }
    setEditItem(null);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Master Harga"
        description="Kelola harga satuan material, upah, dan alat"
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleAutoFill}>
              <Wand2 className="w-4 h-4 mr-2" />
              Auto Fill
            </Button>
            <Button variant="outline" onClick={handleReset}>
              <RotateCcw className="w-4 h-4 mr-2" />
              Reset
            </Button>
            <Button
              className="bg-amber-500 hover:bg-amber-600"
              onClick={() => {
                setEditItem(null);
                setShowForm(true);
              }}
            >
              <Plus className="w-4 h-4 mr-2" />
              Tambah Harga
            </Button>
          </div>
        }
      />

      <div className="flex items-center gap-4">
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Cari harga..."
          className="max-w-sm"
        />
        <Select value={filterKategori} onValueChange={(v) => setFilterKategori(v || "all")}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="Kategori" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua</SelectItem>
            {TIPE_KOMPONEN.map((tipe) => (
              <SelectItem key={tipe.value} value={tipe.value}>
                {tipe.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        emptyTitle="Belum ada data harga"
        emptyDescription="Tambahkan harga satuan material, upah, atau alat"
      />

      {/* Form Dialog */}
      <HargaFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        item={editItem}
        onSubmit={handleSubmit}
      />

      {/* Delete Dialog */}
      <ConfirmDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        title="Hapus Harga"
        description="Apakah Anda yakin ingin menghapus data harga ini?"
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}

function HargaFormDialog({
  open,
  onOpenChange,
  item,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  item: MasterHarga | null;
  onSubmit: (data: Partial<MasterHarga>) => void;
}) {
  const [form, setForm] = useState({
    nama: item?.nama || "",
    satuan: item?.satuan || "",
    harga: item?.harga || 0,
    kategori: item?.kategori || "material",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form as Partial<MasterHarga>);
    onOpenChange(false);
    setForm({ nama: "", satuan: "", harga: 0, kategori: "material" });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{item ? "Edit Harga" : "Tambah Harga"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Nama *</Label>
            <Input
              value={form.nama}
              onChange={(e) => setForm({ ...form, nama: e.target.value })}
              placeholder="Nama material/upah/alat"
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Satuan</Label>
              <Input
                value={form.satuan}
                onChange={(e) => setForm({ ...form, satuan: e.target.value })}
                placeholder="kg, m3, OH"
              />
            </div>
            <div className="space-y-2">
              <Label>Kategori</Label>
              <Select
                value={form.kategori}
                onValueChange={(value) =>
                  setForm({ ...form, kategori: value as TipeKomponen })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TIPE_KOMPONEN.map((tipe) => (
                    <SelectItem key={tipe.value} value={tipe.value}>
                      {tipe.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <CurrencyInput
            label="Harga"
            value={form.harga}
            onChange={(value) => setForm({ ...form, harga: value })}
          />
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Batal
            </Button>
            <Button
              type="submit"
              className="bg-amber-500 hover:bg-amber-600"
            >
              {item ? "Simpan" : "Tambah"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
