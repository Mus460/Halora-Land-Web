"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, FileText, Printer, CheckCircle2, Undo2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { CurrencyInput } from "@/components/shared/currency-input";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import { INVOICE_STATUS } from "@/lib/constants";
import { exportInvoicePdf } from "@/lib/export-invoice-pdf";
import type { Invoice } from "@/types";
import toast from "react-hot-toast";
import { useProject } from "@/contexts/ProjectContext";
import { EmptyProyekState } from "@/components/shared/empty-proyek-state";

export default function InvoicePage() {
  const { currentProyekId: proyekId, proyekList, loading: proyekLoading } = useProject();

  const [data, setData] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    tanggal: "",
    total: 0,
    status: "draft" as "draft" | "sent" | "paid",
  });
  const [markPaid, setMarkPaid] = useState<Invoice | null>(null);
  const [catatKeuangan, setCatatKeuangan] = useState(false);
  const [revertInvoice, setRevertInvoice] = useState<Invoice | null>(null);
  const [updating, setUpdating] = useState(false);

  useEffect(() => {
    if (proyekId) {
      fetchData();
    }
  }, [proyekId]);

  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (proyekList.length === 0) {
    return (
      <EmptyProyekState
        title="Belum Ada Invoice"
        description="Buat proyek untuk mulai membuat dan mengelola invoice"
      />
    );
  }

  const fetchData = async () => {
    if (!proyekId) return;

    try {
      setLoading(true);
      const response = await fetch(`/api/proyek/${proyekId}/invoice`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result.invoices || []);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!proyekId) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/proyek/${proyekId}/invoice`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal membuat invoice");
      }
      setData((prev) => [result.invoice, ...prev]);
      setShowForm(false);
      setForm({ tanggal: "", total: 0, status: "draft" });
      toast.success("Invoice berhasil dibuat");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setSaving(false);
    }
  };

  const handlePrint = (inv: Invoice) => {
    try {
      exportInvoicePdf(inv);
      toast.success("Invoice dicetak");
    } catch (error) {
      toast.error("Gagal mencetak invoice");
      console.error(error);
    }
  };

  const handleMarkPaid = async () => {
    if (!proyekId || !markPaid) return;

    try {
      setUpdating(true);
      const response = await fetch(`/api/proyek/${proyekId}/invoice/${markPaid.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          status: "paid",
          catatKeuangan,
        }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal mengubah status invoice");
      }
      setData((prev) =>
        prev.map((i) => (i.id === result.invoice.id ? result.invoice : i))
      );
      setMarkPaid(null);
      setCatatKeuangan(false);
      toast.success(
        catatKeuangan
          ? "Invoice lunas, pemasukan draft dibuat di Keuangan"
          : "Invoice ditandai lunas"
      );
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setUpdating(false);
    }
  };

  const handleRevert = async () => {
    if (!proyekId || !revertInvoice) return;

    try {
      setUpdating(true);
      const response = await fetch(`/api/proyek/${proyekId}/invoice/${revertInvoice.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: "draft" }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal mengubah status invoice");
      }
      setData((prev) =>
        prev.map((i) => (i.id === result.invoice.id ? result.invoice : i))
      );
      setRevertInvoice(null);
      toast.success("Invoice dikembalikan ke draft, entri keuangan ditandai dibatalkan");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setUpdating(false);
    }
  };

  const columns: ColumnDef<Invoice>[] = [
    {
      accessorKey: "nomor",
      header: "No. Invoice",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.nomor}</span>
      ),
    },
    {
      accessorKey: "tanggal",
      header: "Tanggal",
      cell: ({ row }) => formatDateShort(row.original.tanggal),
    },
    {
      accessorKey: "total",
      header: "Total",
      cell: ({ row }) => (
        <span className="font-semibold">
          {formatCurrency(row.original.total)}
        </span>
      ),
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = INVOICE_STATUS.find(
          (s) => s.value === row.original.status
        );
        return (
          <Badge
            variant={
              row.original.status === "paid"
                ? "default"
                : row.original.status === "sent"
                ? "secondary"
                : "outline"
            }
          >
            {status?.label || row.original.status}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "Aksi",
      cell: ({ row }) => (
        <div className="flex items-center gap-1">
          {row.original.status === "paid" ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setRevertInvoice(row.original)}
            >
              <Undo2 className="w-4 h-4 mr-1" />
              Batalkan Lunas
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setMarkPaid(row.original)}
            >
              <CheckCircle2 className="w-4 h-4 mr-1" />
              Tandai Lunas
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handlePrint(row.original)}
          >
            <Printer className="w-4 h-4 mr-1" />
            Cetak
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Invoice"
        description="Kelola invoice untuk proyek"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => setShowForm(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            Buat Invoice
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data}
        loading={loading}
        emptyTitle="Belum ada invoice"
        emptyDescription="Buat invoice untuk tagihan proyek"
      />

      <Dialog open={showForm} onOpenChange={setShowForm}>        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Buat Invoice</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="tanggal">Tanggal *</Label>
              <Input
                id="tanggal"
                type="date"
                value={form.tanggal}
                onChange={(e) => setForm({ ...form, tanggal: e.target.value })}
                required
              />
            </div>
            <CurrencyInput
              label="Total *"
              value={form.total}
              onChange={(value) => setForm({ ...form, total: value })}
            />
            <div className="space-y-2">
              <Label>Status</Label>
              <Select
                value={form.status}
                onValueChange={(value) =>
                  setForm({ ...form, status: value as "draft" | "sent" | "paid" })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {INVOICE_STATUS.map((s) => (
                    <SelectItem key={s.value} value={s.value}>
                      {s.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowForm(false)}
              >
                Batal
              </Button>
              <Button
                type="submit"
                disabled={saving}
                className="bg-amber-500 hover:bg-amber-600"
              >
                {saving ? "Menyimpan..." : "Simpan"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={markPaid !== null} onOpenChange={(o) => !o && setMarkPaid(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Tandai Invoice Lunas</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              {markPaid?.nomor} · {formatCurrency(markPaid?.total || 0)} — yakin
              menandai invoice ini sebagai lunas?
            </p>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={catatKeuangan}
                onChange={(e) => setCatatKeuangan(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300 accent-amber-500"
              />
              <span>
                Catat pemasukan ke Keuangan{" "}
                <span className="text-xs text-gray-500">
                  (membuat entri pemasukan draft)
                </span>
              </span>
            </label>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setMarkPaid(null)}>
              Batal
            </Button>
            <Button
              onClick={handleMarkPaid}
              disabled={updating}
              className="bg-green-600 hover:bg-green-700"
            >
              {updating ? "Menyimpan..." : "Tandai Lunas"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={revertInvoice !== null}
        onOpenChange={(o) => !o && setRevertInvoice(null)}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Batalkan Lunas</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-gray-600">
            {revertInvoice?.nomor} akan dikembalikan ke status draft. Jika ada
            entri pemasukan terkait di Keuangan, entri tersebut akan ditandai
            "Dibatalkan" (tetap tersimpan).
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRevertInvoice(null)}>
              Batal
            </Button>
            <Button
              onClick={handleRevert}
              disabled={updating}
              className="bg-amber-500 hover:bg-amber-600"
            >
              {updating ? "Menyimpan..." : "Batalkan Lunas"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
