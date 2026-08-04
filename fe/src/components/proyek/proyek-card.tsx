"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Building2,
  Copy,
  Handshake,
  MoreVertical,
  Pencil,
  Trash2,
  MapPin,
  Calendar,
  DollarSign,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { formatCurrency, formatDate } from "@/lib/utils";
import type { Proyek } from "@/types";

interface ProyekCardProps {
  proyek: Proyek;
  onDelete: (id: number) => void;
  onDuplicate: (id: number) => void;
  onSetActive: (id: number) => void;
  onSetPitching: (id: number, isPitching: boolean) => void;
  isActive?: boolean;
}

export function ProyekCard({
  proyek,
  onDelete,
  onDuplicate,
  onSetActive,
  onSetPitching,
  isActive,
}: ProyekCardProps) {
  const [showDelete, setShowDelete] = useState(false);

  return (
    <>
      <Card
        className={`hover:shadow-md transition-shadow ${
          isActive ? "ring-2 ring-amber-500" : ""
        }`}
      >
        <CardContent className="p-5">
          <div className="flex items-start justify-between mb-3">
            <Link href={`/proyek/${proyek.id}`} className="flex-1 min-w-0">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center shrink-0">
                  <Building2 className="w-5 h-5 text-amber-600" />
                </div>
                <div className="min-w-0">
                  <h3 className="font-semibold text-gray-900 truncate">
                    {proyek.namaProyek}
                  </h3>
                  <div className="flex items-center gap-1.5 mt-1">
                    <Badge variant="outline" className="text-xs">
                      {proyek.tipe === "gedung" ? "Gedung" : "Infrastruktur"}
                    </Badge>
                    {proyek.isPitching && (
                      <Badge variant="secondary" className="text-xs bg-amber-100 text-amber-700">
                        Pitching
                      </Badge>
                    )}
                  </div>
                </div>
              </div>
            </Link>
            <DropdownMenu>
              <DropdownMenuTrigger className="inline-flex items-center justify-center h-8 w-8 rounded-md hover:bg-gray-100 transition-colors">
                <MoreVertical className="w-4 h-4" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onSetActive(proyek.id)}>
                  <Building2 className="w-4 h-4 mr-2" />
                  {isActive ? "Proyek Aktif" : "Jadikan Aktif"}
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => onSetPitching(proyek.id, !proyek.isPitching)}
                >
                  <Handshake className="w-4 h-4 mr-2" />
                  {proyek.isPitching ? "Jadikan Aktif" : "Jadikan Pitching"}
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Pencil className="w-4 h-4 mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onDuplicate(proyek.id)}>
                  <Copy className="w-4 h-4 mr-2" />
                  Duplikat
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="text-red-600"
                  onClick={() => setShowDelete(true)}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Hapus
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <Link href={`/proyek/${proyek.id}`}>
            <div className="space-y-2 text-sm text-gray-600">
              {proyek.lokasi && (
                <div className="flex items-center gap-2">
                  <MapPin className="w-4 h-4 text-gray-400" />
                  <span className="truncate">{proyek.lokasi}</span>
                </div>
              )}
              {proyek.timeline && (
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-gray-400" />
                  <span>{proyek.timeline}</span>
                </div>
              )}
              {proyek.nilaiKontrak && (
                <div className="flex items-center gap-2">
                  <DollarSign className="w-4 h-4 text-gray-400" />
                  <span className="font-semibold text-gray-900">
                    {formatCurrency(proyek.nilaiKontrak)}
                  </span>
                </div>
              )}
            </div>
          </Link>

          <div className="mt-4 pt-3 border-t flex items-center justify-between text-xs text-gray-500">
            <span>{proyek._count?.pekerjaan || 0} item pekerjaan</span>
            <span>Update: {formatDate(proyek.updatedAt)}</span>
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        title="Hapus Proyek"
        description={`Apakah Anda yakin ingin menghapus "${proyek.namaProyek}"? Semua data terkait akan dihapus.`}
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={() => {
          onDelete(proyek.id);
          setShowDelete(false);
        }}
      />
    </>
  );
}
