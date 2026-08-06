"use client";

import { useEffect, useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { formatDate } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import toast from "react-hot-toast";
import type { User } from "@/types";

export default function AdminUsersPage() {
  const [data, setData] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch("/api/users");
      if (!response.ok) throw new Error("Failed to fetch");
      const result = await response.json();
      setData(result.users || []);
    } catch (error) {
      console.error("Fetch error:", error);
      toast.error("Gagal memuat data user");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleCreate = async (form: { namaLengkap: string; email: string; password: string }) => {
    try {
      const response = await fetch("/api/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Create failed");
      }
      toast.success("User berhasil ditambahkan");
      setShowForm(false);
      await fetchData();
    } catch (error) {
      console.error("Submit error:", error);
      toast.error(error instanceof Error ? error.message : "Gagal menambahkan user");
    }
  };

  const columns: ColumnDef<User>[] = [
    {
      accessorKey: "namaLengkap",
      header: "Nama",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.namaLengkap}</p>
          <p className="text-xs text-gray-500">{row.original.email}</p>
        </div>
      ),
    },
    {
      accessorKey: "role",
      header: "Role",
      cell: ({ row }) => {
        const colors: Record<string, string> = {
          ADMIN: "bg-red-100 text-red-700",
          OWNER: "bg-purple-100 text-purple-700",
          USER: "bg-blue-100 text-blue-700",
          DEMO: "bg-gray-100 text-gray-700",
        };
        return (
          <Badge variant="outline" className={colors[row.original.role]}>
            {row.original.role}
          </Badge>
        );
      },
    },
    {
      accessorKey: "accountType",
      header: "Tipe Akun",
      cell: ({ row }) => (
        <Badge variant="outline">
          {row.original.accountType === "pro" ? "Pro" : "Free"}
        </Badge>
      ),
    },
    {
      accessorKey: "createdAt",
      header: "Terdaftar",
      cell: ({ row }) => formatDate(row.original.createdAt),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Kelola User"
        description="Manajemen pengguna sistem"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => setShowForm(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            Tambah User
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data}
        emptyTitle={loading ? "Memuat data..." : "Belum ada user"}
        emptyDescription={loading ? "" : "Tambahkan user baru untuk mulai menggunakan sistem"}
      />

      <UserFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        onSubmit={handleCreate}
      />
    </div>
  );
}

function UserFormDialog({
  open,
  onOpenChange,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { namaLengkap: string; email: string; password: string }) => void;
}) {
  const [form, setForm] = useState({ namaLengkap: "", email: "", password: "" });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form);
    setForm({ namaLengkap: "", email: "", password: "" });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Tambah User</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Nama Lengkap (opsional)</Label>
            <Input
              value={form.namaLengkap}
              onChange={(e) => setForm({ ...form, namaLengkap: e.target.value })}
              placeholder="Nama user"
            />
          </div>
          <div className="space-y-2">
            <Label>Email *</Label>
            <Input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="user@email.com"
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Password *</Label>
            <Input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder="Minimal 6 karakter"
              minLength={6}
              required
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
              Tambah
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
