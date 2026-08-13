"use client";

import { useState, useEffect, useMemo } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Pencil, Trash2, Search, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { formatCurrency, formatDuration } from "@/lib/utils";
import { LEVEL_PEKERJAAN, SATUAN_OPTIONS } from "@/lib/constants";
import { useWorkItem } from "@/hooks/useWorkItem";
import { useProject } from "@/contexts/ProjectContext";
import { useDebouncedValue } from "@/hooks/use-debounce";
import type { WorkItem, WorkCategory, CalculationMethod } from "@/types";
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
  code: string;
  name: string;
  unit: string | null;
  unitPrice: number | null;
  ahspCode: string | null;
  ahspSheet: string | null;
}

interface WorkItemPageProps {
  category: WorkCategory;
  title: string;
  description: string;
  icon: LucideIcon;
  projectId?: number;
  formFields?: FormField[];
  showLevelPekerjaan?: boolean;
  showTipePekerjaan?: boolean;
  tipeOptions?: string[];
}

export function WorkItemPage({
  category,
  title,
  description,
  icon: Icon,
  projectId,
  formFields = [],
  showLevelPekerjaan = false,
  showTipePekerjaan = false,
  tipeOptions = [],
}: WorkItemPageProps) {
  const { currentProjectId } = useProject();
  const activeProjectId = projectId ?? currentProjectId ?? undefined;
  const { data, loading, createPekerjaan, updatePekerjaan, deletePekerjaan } = useWorkItem({ category, projectId: activeProjectId });
  const [showForm, setShowForm] = useState(false);
  const [editItem, setEditItem] = useState<WorkItem | null>(null);
  const [showDelete, setShowDelete] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const columns: ColumnDef<WorkItem>[] = [
    {
      accessorKey: "description",
      header: "Uraian WorkItem",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.description}</p>
          {row.original.level && (
            <p className="text-xs text-gray-500">
              {row.original.level}
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
          {row.original.volume} {row.original.unit}
        </span>
      ),
    },
    {
      accessorKey: "unitPrice",
      header: "Harga Satuan",
      cell: ({ row }) => formatCurrency(row.original.unitPrice),
    },
    {
      accessorKey: "totalCost",
      header: "Total Biaya",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.totalCost)}
        </span>
      ),
    },
    {
      accessorKey: "totalDuration",
      header: "Estimasi Waktu",
      cell: ({ row }) => (
        <span className="text-gray-600">{formatDuration(row.original.totalDuration)}</span>
      ),
    },
    {
      accessorKey: "calculationMethod",
      header: "Metode",
      cell: ({ row }) => (
        <Badge variant="outline" className="text-xs">
          {row.original.calculationMethod.toUpperCase()}
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

  const totalCost = data.reduce((sum, item) => sum + item.totalCost, 0);
  const totalDuration = data.reduce((sum, item) => sum + (item.totalDuration || 0), 0);

  const handleDelete = async () => {
    if (deleteId) {
      const success = await deletePekerjaan(deleteId);
      if (success) {
        setShowDelete(false);
        setDeleteId(null);
      }
    }
  };

  const handleSubmit = async (formData: Partial<WorkItem>) => {
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
            {formatCurrency(totalCost)}
          </p>
        </div>
        <div className="pl-4 border-l border-amber-200">
          <p className="text-sm text-amber-700">Estimasi Duration</p>
          <p className="text-lg font-bold text-amber-900">
            {formatDuration(totalDuration)}
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
        emptyDescription={`Tambahkan item ${title.toLowerCase()} untuk project ini`}
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
        title="Hapus WorkItem"
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
  item: WorkItem | null;
  formFields: FormField[];
  showLevelPekerjaan: boolean;
  showTipePekerjaan: boolean;
  tipeOptions: string[];
  onSubmit: (data: Partial<WorkItem>) => void;
}) {
  const [metode, setMetode] = useState<CalculationMethod>(
    item?.calculationMethod || "ahsp"
  );
  const [form, setForm] = useState({
    description: item?.description || "",
    volume: item?.volume || 0,
    unit: item?.unit || "m2",
    unitPrice: item?.unitPrice || 0,
    level: item?.level || "",
    type: item?.type || "",
    calculationMethod: item?.calculationMethod || "ahsp",
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
  const [formError, setFormError] = useState("");

  useEffect(() => {
    if (!open) {
      setAhspQuery("");
      setAhspResults([]);
      setSelectedAhsp(null);
      setAhspError("");
      setFormError("");
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
        `/api/analysis-masters/search?q=${encodeURIComponent(q)}&limit=8`
      );
      if (!response.ok) throw new Error("Search failed");
      const data = await response.json();
      const results = (data.results || []).filter(
        (r: AHSPResult) => r.unitPrice != null
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
      description: item.name,
      unit: item.unit || "m2",
      unitPrice: item.unitPrice || 0,
    });
  };

  const clearAhsp = () => {
    setSelectedAhsp(null);
    setAhspError("");
  };

  const volume = useMemo(() => {
    const u = (form.unit || "").toLowerCase().trim();
    if (u === "m3" && dimensi.panjang > 0 && dimensi.lebar > 0 && dimensi.tinggi > 0) {
      return dimensi.panjang * dimensi.lebar * dimensi.tinggi;
    }
    if (u === "m2" && dimensi.panjang > 0 && dimensi.lebar > 0) {
      return dimensi.panjang * dimensi.lebar;
    }
    if ((u === "m1" || u === "m" || u === "m'") && dimensi.panjang > 0) {
      return dimensi.panjang;
    }
    return form.volume;
  }, [form.unit, form.volume, dimensi]);

  const totalCost = volume * form.unitPrice;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setFormError("");
    for (const field of formFields) {
      if (field.required && !String((form as any)[field.name] || "").trim()) {
        setFormError(`${field.label} wajib diisi`);
        return;
      }
    }
    if (metode === "ahsp" && !item && !selectedAhsp) {
      setAhspError("Pilih item AHSP terlebih dahulu");
      return;
    }
    if (!(volume > 0)) {
      setAhspError("Volume/jumlah wajib diisi");
      return;
    }
    onSubmit({
      ...form,
      volume,
      totalCost,
      calculationMethod: metode,
      analysisMasterId: selectedAhsp?.id ?? null,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-5xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{item ? "Edit Pekerjaan" : "Tambah Pekerjaan"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Uraian Pekerjaan *</Label>
            <Input
              value={form.description}
              onChange={(e) =>
                setForm({ ...form, description: e.target.value })
              }
              placeholder="Deskripsi pekerjaan"
              required
            />
          </div>

          {/* Dynamic form fields */}
          {formFields.map((field) => (
            <div key={field.name} className="space-y-2">
              <Label>
                {field.label}
                {field.required ? " *" : ""}
              </Label>
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

          {/* Level & Type */}
          {showLevelPekerjaan && (
            <div className="space-y-2">
              <Label>Level Pekerjaan</Label>
              <Select
                value={form.level}
                onValueChange={(value) =>
                  setForm({ ...form, level: value || "" })
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
                value={form.type}
                onValueChange={(value) =>
                  setForm({ ...form, type: value || "" })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih tipe" />
                </SelectTrigger>
                <SelectContent>
                  {tipeOptions.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
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
              onValueChange={(v) => setMetode(v as CalculationMethod)}
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
            unit={form.unit}
            panjang={dimensi.panjang}
            lebar={dimensi.lebar}
            tinggi={dimensi.tinggi}
            onChange={setDimensi}
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
                value={form.unit}
                onValueChange={(value) => setForm({ ...form, unit: value || "m2" })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {SATUAN_OPTIONS.map((s) => (
                    <SelectItem key={s.value} value={s.value}>
                      {s.label}
                    </SelectItem>
                  ))}
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
                      <p className="text-sm font-medium line-clamp-2">
                        {selectedAhsp.ahspCode && (
                        <span className="text-amber-600 mr-1">
                          {selectedAhsp.ahspCode}
                        </span>
                      )}
                      {selectedAhsp.name}
                    </p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      {selectedAhsp.unit} · {formatCurrency(selectedAhsp.unitPrice || 0)}
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
                          <p className="text-sm font-medium line-clamp-2">
                            {r.ahspCode && (
                              <span className="text-amber-600 mr-1">
                                {r.ahspCode}
                              </span>
                            )}
                            {r.name}
                          </p>
                          <p className="text-xs text-gray-500 mt-0.5">
                            {r.unit} · {formatCurrency(r.unitPrice || 0)}
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
              value={form.unitPrice}
              onChange={(value) => setForm({ ...form, unitPrice: value })}
            />
          )}

          {/* Summary */}
          {volume > 0 && (
            <div className="p-3 bg-gray-50 rounded-lg space-y-1">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Volume</span>
                <span className="font-medium">
                  {volume.toFixed(2)} {form.unit}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Price Unit</span>
                <span className="font-medium">
                  {formatCurrency(form.unitPrice)}
                </span>
              </div>
              <div className="flex justify-between text-sm font-bold border-t pt-1">
                <span>Total Biaya</span>
                <span className="text-amber-600">
                  {formatCurrency(totalCost)}
                </span>
              </div>
            </div>
          )}

          {formError && (
            <p className="text-xs text-red-500">{formError}</p>
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
