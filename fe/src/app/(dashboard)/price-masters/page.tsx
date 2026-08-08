"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Pencil, Trash2, RotateCcw, Wand2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { SearchInput } from "@/components/shared/search-input";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { formatCurrency } from "@/lib/utils";
import { TIPE_KOMPONEN } from "@/lib/constants";
import { useDebouncedValue } from "@/hooks/use-debounce";
import toast from "react-hot-toast";
import type { PriceMaster, ComponentType } from "@/types";
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
  const [data, setData] = useState<PriceMaster[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search);
  const [filterKategori, setFilterKategori] = useState<string>("all");
  const [showForm, setShowForm] = useState(false);
  const [editItem, setEditItem] = useState<PriceMaster | null>(null);
  const [showDelete, setShowDelete] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  // Fetch data from API
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/price-masters');
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.priceMaster || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  const filtered = data.filter((item) => {
    const matchSearch = item.name
      .toLowerCase()
      .includes(debouncedSearch.toLowerCase());
    const matchKategori =
      filterKategori === "all" || item.type === filterKategori;
    return matchSearch && matchKategori;
  });

  const columns: ColumnDef<PriceMaster>[] = [
    {
      accessorKey: "name",
      header: "Nama",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.name}</span>
      ),
    },
    {
      accessorKey: "unit",
      header: "Satuan",
      cell: ({ row }) => (
        <span className="text-gray-600">{row.original.unit}</span>
      ),
    },
    {
      accessorKey: "price",
      header: "Harga (Rp)",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.price)}
        </span>
      ),
    },
    {
      accessorKey: "category",
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
            className={colors[row.original.type] || ""}
          >
            {row.original.type}
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
    // Auto-fill functionality removed - not applicable with real API
    toast.error("Fitur auto-fill tidak tersedia");
  };

  const handleReset = () => {
    // Reset functionality removed - not applicable with real API
    toast.error("Fitur reset tidak tersedia");
  };

  const handleDelete = async () => {
    if (!deleteId) return;
    
    try {
      const response = await fetch(`/api/price-masters/${deleteId}`, {
        method: 'DELETE',
      });
      
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Delete failed');
      }
      
      await fetchData(); // Refresh data
      toast.success("Data berhasil dihapus");
    } catch (error) {
      console.error('Delete error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal menghapus data');
    } finally {
      setShowDelete(false);
      setDeleteId(null);
    }
  };

  const handleSubmit = async (formData: Partial<PriceMaster>) => {
    try {
      if (editItem) {
        // Update
        const response = await fetch(`/api/price-masters/${editItem.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData),
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || 'Update failed');
        }
        
        toast.success("Data berhasil diupdate");
      } else {
        // Create
        const response = await fetch('/api/price-masters', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData),
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || 'Create failed');
        }
        
        toast.success("Data berhasil ditambahkan");
      }
      
      await fetchData(); // Refresh data
      setEditItem(null);
      setShowForm(false);
    } catch (error) {
      console.error('Submit error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal menyimpan data');
    }
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
              Tambah Price
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
            {TIPE_KOMPONEN.map((type) => (
              <SelectItem key={type.value} value={type.value}>
                {type.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        emptyTitle={loading ? "Memuat data..." : "Belum ada data harga"}
        emptyDescription={loading ? "" : "Tambahkan harga satuan material, upah, atau alat"}
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
  item: PriceMaster | null;
  onSubmit: (data: Partial<PriceMaster>) => void;
}) {
  const [form, setForm] = useState({
    name: item?.name || "",
    unit: item?.unit || "",
    price: item?.price || 0,
    type: item?.type || "material",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form as Partial<PriceMaster>);
    onOpenChange(false);
    setForm({ name: "", unit: "", price: 0, type: "material" });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{item ? "Edit Harga" : "Tambah Harga"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Nama *</Label>
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Nama material/upah/alat"
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Satuan</Label>
              <Input
                value={form.unit}
                onChange={(e) => setForm({ ...form, unit: e.target.value })}
                placeholder="kg, m3, OH"
              />
            </div>
            <div className="space-y-2">
              <Label>Kategori</Label>
              <Select
                value={form.type}
                onValueChange={(value) =>
                  setForm({ ...form, type: value as ComponentType })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TIPE_KOMPONEN.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <CurrencyInput
            label="Harga"
            value={form.price}
            onChange={(value) => setForm({ ...form, price: value })}
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
