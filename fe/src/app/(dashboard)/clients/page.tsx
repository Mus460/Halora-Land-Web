"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { SearchInput } from "@/components/shared/search-input";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { useDebouncedValue } from "@/hooks/use-debounce";
import toast from "react-hot-toast";
import type { Client } from "@/types";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function MasterKlienPage() {
  const [data, setData] = useState<Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search);
  const [showForm, setShowForm] = useState(false);
  const [editItem, setEditItem] = useState<Client | null>(null);
  const [showDelete, setShowDelete] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/clients?search=${encodeURIComponent(debouncedSearch)}`);
      if (!response.ok) throw new Error("Failed to fetch");
      const result = await response.json();
      setData(result.clients || []);
    } catch (error) {
      console.error("Fetch error:", error);
      toast.error("Gagal memuat data klien");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [debouncedSearch]);

  const handleDelete = async () => {
    if (!deleteId) return;
    try {
      const response = await fetch(`/api/clients/${deleteId}`, { method: "DELETE" });
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Delete failed");
      }
      toast.success("Klien berhasil dihapus");
      await fetchData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menghapus klien");
    } finally {
      setShowDelete(false);
      setDeleteId(null);
    }
  };

  const handleSubmit = async (formData: Partial<Client>) => {
    try {
      if (editItem) {
        const response = await fetch(`/api/clients/${editItem.id}`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(formData),
        });
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || "Update failed");
        }
        toast.success("Klien berhasil diupdate");
      } else {
        const response = await fetch("/api/clients", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(formData),
        });
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error || "Create failed");
        }
        toast.success("Klien berhasil ditambahkan");
      }
      await fetchData();
      setEditItem(null);
      setShowForm(false);
    } catch (error) {
      console.error("Submit error:", error);
      toast.error(error instanceof Error ? error.message : "Gagal menyimpan klien");
    }
  };

  const columns: ColumnDef<Client>[] = [
    {
      accessorKey: "name",
      header: "Nama",
      cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
    },
    {
      accessorKey: "address",
      header: "Alamat",
      cell: ({ row }) => (
        <span className="text-gray-600">{row.original.address || "-"}</span>
      ),
    },
    {
      accessorKey: "contact",
      header: "Kontak",
      cell: ({ row }) => (
        <span className="text-gray-600">{row.original.contact || "-"}</span>
      ),
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

  return (
    <div className="space-y-6">
      <PageHeader
        title="Master Klien"
        description="Kelola data klien untuk mempercepat pembuatan invoice"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => {
              setEditItem(null);
              setShowForm(true);
            }}
          >
            <Plus className="w-4 h-4 mr-2" />
            Tambah Klien
          </Button>
        }
      />

      <SearchInput
        value={search}
        onChange={setSearch}
        placeholder="Cari klien..."
        className="max-w-sm"
      />

      <DataTable
        columns={columns}
        data={data}
        emptyTitle={loading ? "Memuat data..." : "Belum ada klien"}
        emptyDescription={loading ? "" : "Tambahkan klien agar data pembeli invoice terisi otomatis"}
      />

      <ClientFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        item={editItem}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        title="Hapus Klien"
        description="Apakah Anda yakin ingin menghapus klien ini?"
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}

function ClientFormDialog({
  open,
  onOpenChange,
  item,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  item: Client | null;
  onSubmit: (data: Partial<Client>) => void;
}) {
  const [form, setForm] = useState({
    name: "",
    address: "",
    contact: "",
  });

  useEffect(() => {
    if (!open) return;
    setForm({
      name: item?.name || "",
      address: item?.address || "",
      contact: item?.contact || "",
    });
  }, [open, item]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form);
    onOpenChange(false);
    setForm({ name: "", address: "", contact: "" });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{item ? "Edit Klien" : "Tambah Klien"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="client-name">Nama *</Label>
            <Input
              id="client-name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Nama klien / perusahaan"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="client-address">Alamat</Label>
            <Input
              id="client-address"
              value={form.address}
              onChange={(e) => setForm({ ...form, address: e.target.value })}
              placeholder="Alamat tagihan"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="client-contact">Kontak</Label>
            <Input
              id="client-contact"
              value={form.contact}
              onChange={(e) => setForm({ ...form, contact: e.target.value })}
              placeholder="PIC / No. telepon"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Batal
            </Button>
            <Button type="submit" className="bg-amber-500 hover:bg-amber-600">
              {item ? "Simpan" : "Tambah"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}