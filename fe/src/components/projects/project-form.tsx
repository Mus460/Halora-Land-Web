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
import type { Project } from "@/types";

interface ProjectFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  project?: Project | null;
  onSubmit: (data: Partial<Project>, boqFile?: File | null) => void;
}

export function ProjectForm({
  open,
  onOpenChange,
  project,
  onSubmit,
}: ProjectFormProps) {
  const [form, setForm] = useState({
    name: project?.name || "",
    location: project?.location || "",
    type: project?.type || "building",
    isPitching: project?.isPitching || false,
    isDone: project?.isDone || false,
    contractValue: project?.contractValue || 0,
    timelineMonths: project?.timelineMonths || 0,
    timelineDays: project?.timelineDays || 0,
  });
  const [boqFile, setBoqFile] = useState<File | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form, boqFile);
    onOpenChange(false);
    setForm({
      name: "",
      location: "",
      type: "building",
      isPitching: false,
      isDone: false,
      contractValue: 0,
      timelineMonths: 0,
      timelineDays: 0,
    });
    setBoqFile(null);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {project ? "Edit Project" : "Buat Proyek Baru"}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!project && (
            <div className="space-y-2">
              <Label htmlFor="boq">File BOQ / RAB (xlsx)</Label>
              <Input
                id="boq"
                type="file"
                accept=".xlsx,.xls"
                onChange={(e) =>
                  setBoqFile(e.target.files?.[0] || null)
                }
              />
              <p className="text-xs text-gray-500">
                Pilih file BOQ untuk otomatis mengisi item pekerjaan, rekap
                per divisi, dan nilai kontrak dari total RAB. Nama proyek bisa
                dikosongkan agar diambil dari judul BOQ.
              </p>
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="name">Nama Proyek {!boqFile ? "*" : ""}</Label>
            <Input
              id="name"
              value={form.name}
              onChange={(e) =>
                setForm({ ...form, name: e.target.value })
              }
              placeholder={boqFile ? "Dari judul BOQ" : "Masukkan nama proyek"}
              required={!boqFile}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="location">Lokasi</Label>
            <Input
              id="location"
              value={form.location}
              onChange={(e) => setForm({ ...form, location: e.target.value })}
              placeholder="Alamat proyek"
            />
          </div>
          <div className="space-y-2">
            <Label>Tipe Proyek</Label>
            <Select
              value={form.type}
              onValueChange={(value) =>
                setForm({ ...form, type: value as "building" | "infrastructure" })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="building">Gedung</SelectItem>
                <SelectItem value="infrastructure">Infrastruktur</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Status Proyek</Label>
            <Select
              value={form.isDone ? "selesai" : form.isPitching ? "pitching" : "aktif"}
              onValueChange={(value) =>
                setForm({
                  ...form,
                  isDone: value === "selesai",
                  isPitching: value === "pitching",
                })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="aktif">Aktif</SelectItem>
                <SelectItem value="pitching">Pitching</SelectItem>
                <SelectItem value="selesai">Selesai</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <CurrencyInput
            label="Nilai Kontrak"
            value={form.contractValue}
            onChange={(value) => setForm({ ...form, contractValue: value })}
          />
          <div className="space-y-2">
            <Label>Timeline</Label>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Input
                  type="number"
                  min={0}
                  value={form.timelineMonths || ""}
                  onChange={(e) =>
                    setForm({ ...form, timelineMonths: Math.max(0, Number(e.target.value) || 0) })
                  }
                  placeholder="0"
                />
                <p className="text-xs text-gray-500">Bulan</p>
              </div>
              <div className="space-y-2">
                <Input
                  type="number"
                  min={0}
                  max={30}
                  value={form.timelineDays || ""}
                  onChange={(e) =>
                    setForm({ ...form, timelineDays: Math.max(0, Number(e.target.value) || 0) })
                  }
                  placeholder="0"
                />
                <p className="text-xs text-gray-500">Hari</p>
              </div>
            </div>
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
              {project ? "Simpan" : "Buat Project"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
