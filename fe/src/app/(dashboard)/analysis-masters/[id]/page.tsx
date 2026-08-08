"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Pencil, Copy as CopyIcon } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { formatCurrency, formatDuration } from "@/lib/utils";
import type { AnalysisMaster, AnalysisComponent } from "@/types";
import toast from "react-hot-toast";

const tipeLabel: Record<string, string> = {
  upah: "Tenaga Kerja",
  material: "Bahan",
  alat: "Peralatan",
};

export default function MasterAnalisaDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [item, setItem] = useState<AnalysisMaster | null>(null);
  const [component, setComponent] = useState<AnalysisComponent[]>([]);
  const [loading, setLoading] = useState(true);
  const [editMeta, setEditMeta] = useState(false);
  const [metaForm, setMetaForm] = useState({ name: "", unit: "" });
  const [editComp, setEditComp] = useState<AnalysisComponent | null>(null);
  const [compForm, setCompForm] = useState({ coefficient: "", unitPrice: "" });
  const [saving, setSaving] = useState(false);

  const load = async () => {
    if (!params?.id) return;
    setLoading(true);
    try {
      const [metaRes, rinRes] = await Promise.all([
        fetch(`/api/analysis-masters/${params.id}`),
        fetch(`/api/analysis-masters/${params.id}/components`),
      ]);
      if (!metaRes.ok) throw new Error("Gagal memuat analisa");
      setItem(await metaRes.json());
      if (rinRes.ok) setComponent(await rinRes.json());
    } catch (error) {
      console.error(error);
      toast.error("Gagal memuat detail analisa");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [params?.id]);

  const isEditable = !!item && !item.isSystem;

  const openMetaEdit = () => {
    if (!item) return;
    setMetaForm({ name: item.name, unit: item.unit || "" });
    setEditMeta(true);
  };

  const handleSaveMeta = async () => {
    if (!item) return;
    try {
      setSaving(true);
      const response = await fetch(`/api/analysis-masters/${item.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: metaForm.name, unit: metaForm.unit || null }),
      });
      const result = await response.json();
      if (!response.ok) throw new Error(result.error || "Gagal menyimpan");
      setItem({ ...item, name: result.name, unit: result.unit });
      setEditMeta(false);
      toast.success("Analisa diperbarui");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menyimpan");
    } finally {
      setSaving(false);
    }
  };

  const openCompEdit = (r: AnalysisComponent) => {
    setEditComp(r);
    setCompForm({ coefficient: String(r.coefficient), unitPrice: String(r.unitPrice || "") });
  };

  const handleSaveComp = async () => {
    if (!editComp || !item) return;
    try {
      setSaving(true);
      const body: Record<string, unknown> = {};
      if (compForm.coefficient.trim() !== "") {
        body.coefficient = Number(compForm.coefficient);
      }
      if (compForm.unitPrice.trim() !== "") {
        body.unitPrice = Number(compForm.unitPrice);
      }
      const response = await fetch(`/api/analysis-masters/${item.id}/components/${editComp.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const result = await response.json();
      if (!response.ok) throw new Error(result.error || "Gagal menyimpan komponen");
      setEditComp(null);
      toast.success("Komponen diperbarui");
      await load();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menyimpan komponen");
    } finally {
      setSaving(false);
    }
  };

  const handleCopy = async () => {
    if (!item) return;
    try {
      setSaving(true);
      const response = await fetch(`/api/analysis-masters/${item.id}/copy`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      });
      const result = await response.json();
      if (!response.ok) throw new Error(result.error || "Gagal menyalin analisa");
      toast.success("Analisa disalin");
      router.push(`/analysis-masters/${result.id}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menyalin analisa");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <p className="p-8 text-center text-gray-500">Memuat data...</p>;
  }
  if (!item) {
    return <p className="p-8 text-center text-gray-500">Analysis tidak ditemukan</p>;
  }

  const totalJumlah = component.reduce((sum, r) => sum + (r.totalPrice || 0), 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Detail Analisa"
        description="Rincian komponen analisa harga satuan pekerjaan"
      />

      <Card>
        <CardContent className="p-4 space-y-3">
          <button
            onClick={() => router.push("/analysis-masters")}
            className="inline-flex items-center gap-1 text-sm text-amber-600 hover:text-amber-700"
          >
            <ArrowLeft className="w-4 h-4" /> Kembali
          </button>
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-gray-400 font-mono">{item.code}</span>
            <h2 className="font-semibold">{item.name}</h2>
            {item.isSystem ? (
              <Badge variant="outline" className="text-gray-500">
                Sistem AHSP
              </Badge>
            ) : (
              <Badge variant="outline" className="text-emerald-600">
                Milik saya · dapat diedit
              </Badge>
            )}
          </div>
          <div className="flex flex-wrap gap-2 items-center">
            {item.unit && <Badge variant="outline">{item.unit}</Badge>}
            {item.unitPrice != null && item.unitPrice > 0 && (
              <Badge variant="outline" className="text-amber-700">
                {formatCurrency(item.unitPrice)}
              </Badge>
            )}
            <div className="ml-auto flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="text-blue-600"
                onClick={handleCopy}
                disabled={saving}
              >
                <CopyIcon className="w-3.5 h-3.5 mr-1" /> Salin
              </Button>
              {isEditable && (
                <Button variant="outline" size="sm" onClick={openMetaEdit}>
                  <Pencil className="w-3.5 h-3.5 mr-1" /> Ubah
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-xs uppercase text-gray-500">
                  <th className="px-4 py-3 w-10">No</th>
                  <th className="px-4 py-3">Description</th>
                  <th className="px-4 py-3">Code</th>
                  <th className="px-4 py-3">Unit</th>
                  <th className="px-4 py-3 text-right">Koefisien</th>
                  <th className="px-4 py-3 text-right">Price Unit</th>
                  <th className="px-4 py-3 text-right">Amount Price</th>
                  <th className="px-4 py-3 text-right">Duration (1/coefficient)</th>
                  {isEditable && <th className="px-4 py-3 w-10" />}
                </tr>
              </thead>
              <tbody>
                {component.length === 0 && (
                  <tr>
                    <td colSpan={isEditable ? 9 : 8} className="px-4 py-6 text-center text-gray-500">
                      Tidak ada component
                    </td>
                  </tr>
                )}
                {component.map((r, idx) => (
                  <tr key={r.id} className="border-b last:border-0">
                    <td className="px-4 py-2.5 text-gray-500">{idx + 1}</td>
                    <td className="px-4 py-2.5 max-w-[280px]">
                      <span
                        className={
                          r.type === "labor" || r.type === "material" || r.type === "equipment"
                            ? "font-medium block truncate"
                            : "block truncate"
                        }
                      >
                        {r.name || "-"}
                      </span>
                      {tipeLabel[r.type] && (
                        <span className="ml-2 text-xs text-gray-400">
                          {tipeLabel[r.type]}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-2.5 font-mono text-xs text-gray-500">
                      {r.referenceCode || "-"}
                    </td>
                    <td className="px-4 py-2.5">{r.unit || "-"}</td>
                    <td className="px-4 py-2.5 text-right font-mono">{r.coefficient}</td>
                    <td className="px-4 py-2.5 text-right">
                      {formatCurrency(r.unitPrice || 0)}
                    </td>
                    <td className="px-4 py-2.5 text-right font-medium">
                      {formatCurrency(r.totalPrice || 0)}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono">
                      {formatDuration(r.duration)}
                    </td>
                    {isEditable && (
                      <td className="px-4 py-2.5 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openCompEdit(r)}
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </Button>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
              {component.length > 0 && (
                <tfoot>
                  <tr className="border-t font-semibold">
                    <td colSpan={6} className="px-4 py-3 text-right">
                      Total (A+B+C)
                    </td>
                    <td className="px-4 py-3 text-right">{formatCurrency(totalJumlah)}</td>
                    <td />
                    {isEditable && <td />}
                  </tr>
                </tfoot>
              )}
            </table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={editMeta} onOpenChange={(open) => !open && setEditMeta(false)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Ubah Analisa</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="meta-name">Nama</Label>
              <Input
                id="meta-name"
                value={metaForm.name}
                onChange={(e) => setMetaForm({ ...metaForm, name: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="meta-unit">Satuan</Label>
              <Input
                id="meta-unit"
                value={metaForm.unit}
                onChange={(e) => setMetaForm({ ...metaForm, unit: e.target.value })}
                placeholder="m2, m1, buah..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setEditMeta(false)}>
              Batal
            </Button>
            <Button type="button" onClick={handleSaveMeta} disabled={saving || !metaForm.name.trim()}>
              {saving ? "Menyimpan..." : "Simpan"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={editComp !== null} onOpenChange={(open) => !open && setEditComp(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Ubah Komponen</DialogTitle>
          </DialogHeader>
          {editComp && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="comp-name">Nama</Label>
                <Input id="comp-name" value={editComp.name || ""} readOnly />
              </div>
              <div className="space-y-2">
                <Label htmlFor="comp-coef">Koefisien</Label>
                <Input
                  id="comp-coef"
                  type="number"
                  step="any"
                  value={compForm.coefficient}
                  onChange={(e) => setCompForm({ ...compForm, coefficient: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="comp-price">Harga Satuan</Label>
                <Input
                  id="comp-price"
                  type="number"
                  step="any"
                  value={compForm.unitPrice}
                  onChange={(e) => setCompForm({ ...compForm, unitPrice: e.target.value })}
                />
              </div>
              <p className="text-sm text-gray-500">
                Harga total (Amount Price) dan harga analisa dihitung ulang otomatis.
              </p>
            </div>
          )}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setEditComp(null)}>
              Batal
            </Button>
            <Button type="button" onClick={handleSaveComp} disabled={saving}>
              {saving ? "Menyimpan..." : "Simpan"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
