"use client";

import { ColumnDef } from "@tanstack/react-table";
import { Users, Pencil } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatDate } from "@/lib/utils";
import { getUsers } from "@/mock";
import type { User } from "@/types";
import { useState } from "react";

export default function AdminUsersPage() {
  const [data] = useState<User[]>(getUsers());

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
    {
      id: "actions",
      header: "Aksi",
      cell: () => (
        <Button variant="ghost" size="sm">
          <Pencil className="w-4 h-4 mr-1" />
          Edit
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Kelola User"
        description="Manajemen pengguna sistem"
      />

      <DataTable
        columns={columns}
        data={data}
        emptyTitle="Belum ada user"
      />
    </div>
  );
}
