"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CurrencyInput } from "@/components/shared/currency-input";
import type { Proyek } from "@/types";

interface ProyekFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  proyek?: Proyek | null;
  onSubmit: (data: Partial<Proyek>) => void;
}

export function ProyekForm({
  open,
  onOpenChange,
  proyek,
  onSubmit,
}: ProyekFormProps) {
  const [form, setForm] = useState({
    namaProyek: proyek?.namaProyek || "",
    lokasi: proyek?.lokasi || "",
    tipe: proyek?.tipe || "gedung",
    isPitching: proyek?.isPitching || false,
    nilaiKontrak: proyek?.nilaiKontrak || 0,
    timeline: proyek?.timeline || "",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form);
    onOpenChange(false);
    setForm({
      namaProyek: "",
      lokasi: "",
      tipe: "gedung",
      isPitching: false,
      nilaiKontrak: 0,
      timeline: "",
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {proyek ? "Edit Proyek" : "Buat Proyek Baru"}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="nama">Nama Proyek *</Label>
            <Input
              id="nama"
              value={form.namaProyek}
              onChange={(e) =>
                setForm({ ...form, namaProyek: e.target.value })
              }
              placeholder="Masukkan nama proyek"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="lokasi">Lokasi</Label>
            <Input
              id="lokasi"
              value={form.lokasi}
              onChange={(e) => setForm({ ...form, lokasi: e.target.value })}
              placeholder="Alamat proyek"
            />
          </div>
          <div className="space-y-2">
            <Label>Tipe Proyek</Label>
            <Select
              value={form.tipe}
              onValueChange={(value) =>
                setForm({ ...form, tipe: value as "gedung" | "infra" })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="gedung">Gedung</SelectItem>
                <SelectItem value="infra">Infrastruktur</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Status Proyek</Label>
            <Select
              value={form.isPitching ? "pitching" : "aktif"}
              onValueChange={(value) =>
                setForm({ ...form, isPitching: value === "pitching" })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="aktif">Aktif</SelectItem>
                <SelectItem value="pitching">Pitching</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <CurrencyInput
            label="Nilai Kontrak"
            value={form.nilaiKontrak}
            onChange={(value) => setForm({ ...form, nilaiKontrak: value })}
          />
          <div className="space-y-2">
            <Label htmlFor="timeline">Timeline</Label>
            <Input
              id="timeline"
              value={form.timeline}
              onChange={(e) =>
                setForm({ ...form, timeline: e.target.value })
              }
              placeholder="Contoh: 6 bulan"
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
            <Button
              type="submit"
              className="bg-amber-500 hover:bg-amber-600"
            >
              {proyek ? "Simpan" : "Buat Proyek"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
