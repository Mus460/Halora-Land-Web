"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Pencil, Trash2, Search, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { formatCurrency, formatWaktu } from "@/lib/utils";
import { LEVEL_PEKERJAAN } from "@/lib/constants";
import { usePekerjaan } from "@/hooks/usePekerjaan";
import { useDebouncedValue } from "@/hooks/use-debounce";
import type { Pekerjaan, KategoriPekerjaan, MetodeHitung } from "@/types";
import type { LucideIcon } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { VolumeInput } from "@/components/shared/volume-input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface FormField {
  name: string;
  label: string;
  type: "select" | "text" | "number" | "checkbox";
  options?: { value: string; label: string }[];
  required?: boolean;
}

interface AHSPResult {
  id: number;
  kode: string;
  nama: string;
  satuan: string | null;
  hargaSatuan: number | null;
  ahspKode: string | null;
  ahspSheet: string | null;
}

interface PekerjaanPageProps {
  kategori: KategoriPekerjaan;
  title: string;
  description: string;
  icon: LucideIcon;
  proyekId?: number; // Changed from initialData to proyekId
  formFields?: FormField[];
  showLevelPekerjaan?: boolean;
  showTipePekerjaan?: boolean;
  tipeOptions?: string[];
}

export function PekerjaanPage({
  kategori,
  title,
  description,
  icon: Icon,
  proyekId = 1, // Default to proyekId 1 for now
  formFields = [],
  showLevelPekerjaan = false,
  showTipePekerjaan = false,
  tipeOptions = [],
}: PekerjaanPageProps) {
  const { data, loading, createPekerjaan, updatePekerjaan, deletePekerjaan } = usePekerjaan({ kategori, proyekId });
  const [showForm, setShowForm] = useState(false);
  const [editItem, setEditItem] = useState<Pekerjaan | null>(null);
  const [showDelete, setShowDelete] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const columns: ColumnDef<Pekerjaan>[] = [
    {
      accessorKey: "uraianPekerjaan",
      header: "Uraian Pekerjaan",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.uraianPekerjaan}</p>
          {row.original.levelPekerjaan && (
            <p className="text-xs text-gray-500">
              {row.original.levelPekerjaan}
            </p>
          )}
        </div>
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
      header: "Total Biaya",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.totalBiaya)}
        </span>
      ),
    },
    {
      accessorKey: "totalWaktu",
      header: "Estimasi Waktu",
      cell: ({ row }) => (
        <span className="text-gray-600">{formatWaktu(row.original.totalWaktu)}</span>
      ),
    },
    {
      accessorKey: "metodeHitung",
      header: "Metode",
      cell: ({ row }) => (
        <Badge variant="outline" className="text-xs">
          {row.original.metodeHitung.toUpperCase()}
        </Badge>
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

  const totalBiaya = data.reduce((sum, item) => sum + item.totalBiaya, 0);
  const totalWaktu = data.reduce((sum, item) => sum + (item.totalWaktu || 0), 0);

  const handleDelete = async () => {
    if (deleteId) {
      const success = await deletePekerjaan(deleteId);
      if (success) {
        setShowDelete(false);
        setDeleteId(null);
      }
    }
  };

  const handleSubmit = async (formData: Partial<Pekerjaan>) => {
    let success = false;
    
    if (editItem) {
      success = await updatePekerjaan(editItem.id, formData);
    } else {
      success = await createPekerjaan(formData);
    }
    
    if (success) {
      setEditItem(null);
      setShowForm(false);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={title}
        description={description}
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => {
              setEditItem(null);
              setShowForm(true);
            }}
          >
            <Plus className="w-4 h-4 mr-2" />
            Tambah {title}
          </Button>
        }
      />

      {/* Summary */}
      <div className="flex items-center gap-4 p-4 bg-amber-50 rounded-lg border border-amber-200">
        <Icon className="w-5 h-5 text-amber-600" />
        <div>
          <p className="text-sm text-amber-700">Total {title}</p>
          <p className="text-lg font-bold text-amber-900">
            {formatCurrency(totalBiaya)}
          </p>
        </div>
        <div className="pl-4 border-l border-amber-200">
          <p className="text-sm text-amber-700">Estimasi Waktu</p>
          <p className="text-lg font-bold text-amber-900">
            {formatWaktu(totalWaktu)}
          </p>
        </div>
        <Badge variant="outline" className="ml-auto">
          {data.length} item
        </Badge>
      </div>

      <DataTable
        columns={columns}
        data={data}
        emptyTitle={`Belum ada data ${title.toLowerCase()}`}
        emptyDescription={`Tambahkan item ${title.toLowerCase()} untuk proyek ini`}
      />

      {/* Form Dialog */}
      <PekerjaanFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        item={editItem}
        formFields={formFields}
        showLevelPekerjaan={showLevelPekerjaan}
        showTipePekerjaan={showTipePekerjaan}
        tipeOptions={tipeOptions}
        onSubmit={handleSubmit}
      />

      {/* Delete Dialog */}
      <ConfirmDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        title="Hapus Pekerjaan"
        description="Apakah Anda yakin ingin menghapus pekerjaan ini?"
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}

function PekerjaanFormDialog({
  open,
  onOpenChange,
  item,
  formFields,
  showLevelPekerjaan,
  showTipePekerjaan,
  tipeOptions,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  item: Pekerjaan | null;
  formFields: FormField[];
  showLevelPekerjaan: boolean;
  showTipePekerjaan: boolean;
  tipeOptions: string[];
  onSubmit: (data: Partial<Pekerjaan>) => void;
}) {
  const [metode, setMetode] = useState<MetodeHitung>(
    item?.metodeHitung || "ahsp"
  );
  const [form, setForm] = useState({
    uraianPekerjaan: item?.uraianPekerjaan || "",
    volume: item?.volume || 0,
    satuan: item?.satuan || "m2",
    hargaSatuan: item?.hargaSatuan || 0,
    levelPekerjaan: item?.levelPekerjaan || "",
    tipePekerjaan: item?.tipePekerjaan || "",
    metodeHitung: item?.metodeHitung || "ahsp",
  });
  const [dimensi, setDimensi] = useState({
    panjang: 0,
    lebar: 0,
    tinggi: 0,
  });
  const [ahspQuery, setAhspQuery] = useState("");
  const [ahspResults, setAhspResults] = useState<AHSPResult[]>([]);
  const [ahspSearching, setAhspSearching] = useState(false);
  const [selectedAhsp, setSelectedAhsp] = useState<AHSPResult | null>(null);
  const [ahspError, setAhspError] = useState("");

  useEffect(() => {
    if (!open) {
      setAhspQuery("");
      setAhspResults([]);
      setSelectedAhsp(null);
      setAhspError("");
    }
  }, [open]);

  const debouncedAhspQuery = useDebouncedValue(ahspQuery);

  useEffect(() => {
    if (debouncedAhspQuery.trim().length >= 3) {
      searchAhsp(debouncedAhspQuery.trim());
    } else {
      setAhspResults([]);
    }
  }, [debouncedAhspQuery]);

  const searchAhsp = async (q: string) => {
    setAhspSearching(true);
    try {
      const response = await fetch(
        `/api/master-analisa/search?q=${encodeURIComponent(q)}&limit=8`
      );
      if (!response.ok) throw new Error("Search failed");
      const data = await response.json();
      const results = (data.results || []).filter(
        (r: AHSPResult) => r.hargaSatuan != null
      );
      setAhspResults(results);
    } catch (error) {
      console.error("AHSP search error:", error);
      setAhspResults([]);
    } finally {
      setAhspSearching(false);
    }
  };

  const pickAhsp = (item: AHSPResult) => {
    setSelectedAhsp(item);
    setAhspResults([]);
    setAhspQuery("");
    setAhspError("");
    setForm({
      ...form,
      uraianPekerjaan: item.nama,
      satuan: item.satuan || "m2",
      hargaSatuan: item.hargaSatuan || 0,
    });
  };

  const clearAhsp = () => {
    setSelectedAhsp(null);
    setAhspError("");
  };

  const volume =
    dimensi.panjang > 0 && dimensi.lebar > 0
      ? dimensi.tinggi > 0
        ? dimensi.panjang * dimensi.lebar * dimensi.tinggi
        : dimensi.panjang * dimensi.lebar
      : form.volume;

  const totalBiaya = volume * form.hargaSatuan;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (metode === "ahsp" && !item && !selectedAhsp) {
      setAhspError("Pilih item AHSP terlebih dahulu");
      return;
    }
    onSubmit({
      ...form,
      volume,
      totalBiaya,
      metodeHitung: metode,
      masterAnalisaId: selectedAhsp?.id ?? null,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{item ? "Edit Pekerjaan" : "Tambah Pekerjaan"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Uraian Pekerjaan *</Label>
            <Input
              value={form.uraianPekerjaan}
              onChange={(e) =>
                setForm({ ...form, uraianPekerjaan: e.target.value })
              }
              placeholder="Deskripsi pekerjaan"
              required
            />
          </div>

          {/* Dynamic form fields */}
          {formFields.map((field) => (
            <div key={field.name} className="space-y-2">
              <Label>{field.label}</Label>
              {field.type === "select" && field.options && (
                <Select
                  value={(form as any)[field.name] || ""}
                  onValueChange={(value) =>
                    setForm({ ...form, [field.name]: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder={`Pilih ${field.label}`} />
                  </SelectTrigger>
                  <SelectContent>
                    {field.options.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
              {field.type === "text" && (
                <Input
                  value={(form as any)[field.name] || ""}
                  onChange={(e) =>
                    setForm({ ...form, [field.name]: e.target.value })
                  }
                  placeholder={field.label}
                />
              )}
            </div>
          ))}

          {/* Level & Tipe */}
          {showLevelPekerjaan && (
            <div className="space-y-2">
              <Label>Level Pekerjaan</Label>
              <Select
                value={form.levelPekerjaan}
                onValueChange={(value) =>
                  setForm({ ...form, levelPekerjaan: value || "" })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih level" />
                </SelectTrigger>
                <SelectContent>
                  {LEVEL_PEKERJAAN.map((level) => (
                    <SelectItem key={level} value={level}>
                      {level}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {showTipePekerjaan && tipeOptions.length > 0 && (
            <div className="space-y-2">
              <Label>Tipe Pekerjaan</Label>
              <Select
                value={form.tipePekerjaan}
                onValueChange={(value) =>
                  setForm({ ...form, tipePekerjaan: value || "" })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih tipe" />
                </SelectTrigger>
                <SelectContent>
                  {tipeOptions.map((tipe) => (
                    <SelectItem key={tipe} value={tipe}>
                      {tipe}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Metode Hitung */}
          <div className="space-y-2">
            <Label>Metode Hitung</Label>
            <Tabs
              value={metode}
              onValueChange={(v) => setMetode(v as MetodeHitung)}
            >
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="ahsp">AHSP</TabsTrigger>
                <TabsTrigger value="manual">Manual</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          {/* Volume */}
          <VolumeInput
            label="Dimensi"
            panjang={dimensi.panjang}
            lebar={dimensi.lebar}
            tinggi={dimensi.tinggi}
            onChange={setDimensi}
            showTinggi={true}
          />

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Volume</Label>
              <Input
                type="number"
                value={volume || ""}
                onChange={(e) =>
                  setForm({ ...form, volume: Number(e.target.value) })
                }
                placeholder="0"
                step="0.01"
              />
            </div>
            <div className="space-y-2">
              <Label>Satuan</Label>
              <Select
                value={form.satuan}
                onValueChange={(value) => setForm({ ...form, satuan: value || "m2" })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {["m2", "m3", "m1", "m", "unit", "kg", "ls", "set", "titik"].map(
                    (s) => (
                      <SelectItem key={s} value={s}>
                        {s}
                      </SelectItem>
                    )
                  )}
                </SelectContent>
              </Select>
            </div>
          </div>

          {metode === "ahsp" && (
            <div className="space-y-2">
              <Label>Item AHSP</Label>
              {selectedAhsp ? (
                <div className="flex items-start justify-between gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg">
                  <div className="min-w-0">
                    <p className="text-sm font-medium truncate">
                      {selectedAhsp.ahspKode && (
                        <span className="text-amber-600 mr-1">
                          {selectedAhsp.ahspKode}
                        </span>
                      )}
                      {selectedAhsp.nama}
                    </p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      {selectedAhsp.satuan} · {formatCurrency(selectedAhsp.hargaSatuan || 0)}
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 shrink-0"
                    onClick={clearAhsp}
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              ) : (
                <div>
                  <div className="relative">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
                    <Input
                      value={ahspQuery}
                      onChange={(e) => setAhspQuery(e.target.value)}
                      placeholder="Cari uraian / kode AHSP (min. 3 huruf)..."
                      className="pl-8"
                    />
                  </div>
                  {(ahspResults.length > 0 || ahspSearching) && (
                    <div className="mt-1 w-full bg-white border rounded-lg shadow-lg max-h-64 overflow-y-auto">
                      {ahspSearching && (
                        <p className="px-3 py-2 text-xs text-gray-500">
                          Mencari...
                        </p>
                      )}
                      {ahspResults.map((r) => (
                        <button
                          key={r.id}
                          type="button"
                          onClick={() => pickAhsp(r)}
                          className="w-full text-left px-3 py-2 hover:bg-amber-50 border-b last:border-0"
                        >
                          <p className="text-sm font-medium truncate">
                            {r.ahspKode && (
                              <span className="text-amber-600 mr-1">
                                {r.ahspKode}
                              </span>
                            )}
                            {r.nama}
                          </p>
                          <p className="text-xs text-gray-500 mt-0.5">
                            {r.satuan} · {formatCurrency(r.hargaSatuan || 0)}
                          </p>
                        </button>
                      ))}
                      {!ahspSearching && ahspResults.length === 0 && ahspQuery.trim().length >= 3 && (
                        <p className="px-3 py-2 text-xs text-gray-500">
                          Tidak ada hasil
                        </p>
                      )}
                    </div>
                  )}
                </div>
              )}
              {ahspError && (
                <p className="text-xs text-red-500">{ahspError}</p>
              )}
            </div>
          )}

          {metode === "manual" && (
            <CurrencyInput
              label="Harga Satuan"
              value={form.hargaSatuan}
              onChange={(value) => setForm({ ...form, hargaSatuan: value })}
            />
          )}

          {/* Summary */}
          {volume > 0 && (
            <div className="p-3 bg-gray-50 rounded-lg space-y-1">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Volume</span>
                <span className="font-medium">
                  {volume.toFixed(2)} {form.satuan}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Harga Satuan</span>
                <span className="font-medium">
                  {formatCurrency(form.hargaSatuan)}
                </span>
              </div>
              <div className="flex justify-between text-sm font-bold border-t pt-1">
                <span>Total Biaya</span>
                <span className="text-amber-600">
                  {formatCurrency(totalBiaya)}
                </span>
              </div>
            </div>
          )}

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
