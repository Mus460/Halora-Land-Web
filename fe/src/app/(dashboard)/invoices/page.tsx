"use client";

import { useState, useEffect } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Plus, Printer, CheckCircle2, Undo2, Send, Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { PageHeader } from "@/components/shared/page-header";
import { DataTable } from "@/components/shared/data-table";
import { formatCurrency, formatDateShort } from "@/lib/utils";
import { INVOICE_STATUS } from "@/lib/constants";
import { InvoiceForm } from "@/components/invoices/invoice-form";
import type { Invoice, InvoiceItem } from "@/types";
import toast from "react-hot-toast";
import { useProject } from "@/contexts/ProjectContext";
import { EmptyProjectState } from "@/components/shared/empty-project-state";

interface InvoiceFormValues {
  date: string;
  dueDate: string;
  poNumber: string;
  buyerName: string;
  buyerAddress: string;
  buyerContact: string;
  discount: number;
  taxRate: number;
  paymentBank: string;
  paymentAccountNumber: string;
  paymentAccountName: string;
  notes: string;
  financeName: string;
  items: InvoiceItem[];
}

const OPTIONAL_STRINGS = [
  "dueDate",
  "poNumber",
  "buyerName",
  "buyerAddress",
  "buyerContact",
  "paymentBank",
  "paymentAccountNumber",
  "paymentAccountName",
  "notes",
  "financeName",
] as const;

function sanitizeValues(values: InvoiceFormValues): Record<string, unknown> {
  const out: Record<string, unknown> = { ...values };
  for (const key of OPTIONAL_STRINGS) {
    if (out[key] === "") out[key] = null;
  }
  return out;
}

export default function InvoicePage() {
  const { currentProjectId: projectId, projectList, loading: proyekLoading } = useProject();

  const [data, setData] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [markPaid, setMarkPaid] = useState<Invoice | null>(null);
  const [recordExpense, setCatatKeuangan] = useState(false);
  const [revertInvoice, setRevertInvoice] = useState<Invoice | null>(null);
  const [editInvoice, setEditInvoice] = useState<Invoice | null>(null);
  const [updating, setUpdating] = useState(false);

  const fetchData = async () => {
    if (!projectId) return;

    try {
      setLoading(true);
      const response = await fetch(`/api/projects/${projectId}/invoices`);
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

  useEffect(() => {
    if (projectId) {
      fetchData();
    }
  }, [projectId]);

  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (projectList.length === 0) {
    return (
      <EmptyProjectState
        title="Belum Ada Invoice"
        description="Buat proyek untuk mulai membuat dan mengelola invoice"
      />
    );
  }

  const handleCreate = async (values: InvoiceFormValues) => {
    if (!projectId) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/projects/${projectId}/invoices`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(sanitizeValues(values)),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal membuat invoice");
      }
      setData((prev) => [result.invoices, ...prev]);
      setShowForm(false);
      toast.success("Invoice berhasil dibuat");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setSaving(false);
    }
  };

  const handlePrint = async (inv: Invoice) => {
    try {
      const project = projectList.find((p) => p.id === projectId);
      const { exportInvoicePdf } = await import("@/lib/export-invoice-pdf");
      exportInvoicePdf(inv, {
        name: project?.name,
        location: project?.location,
      });
      toast.success("Invoice dicetak");
    } catch (error) {
      toast.error("Gagal mencetak invoice");
      console.error(error);
    }
  };

  const handleMarkPaid = async () => {
    if (!projectId || !markPaid) return;

    try {
      setUpdating(true);
      const response = await fetch(`/api/projects/${projectId}/invoices/${markPaid.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          status: "paid",
          recordExpense,
        }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal mengubah status invoice");
      }
      setData((prev) =>
        prev.map((i) => (i.id === result.invoices.id ? result.invoices : i))
      );
      setMarkPaid(null);
      setCatatKeuangan(false);
      toast.success(
        recordExpense
          ? "Invoice lunas, pemasukan draft dibuat di Keuangan"
          : "Invoice ditandai lunas"
      );
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setUpdating(false);
    }
  };

  const handleMarkSent = async (inv: Invoice) => {
    if (!projectId) return;
    try {
      const response = await fetch(`/api/projects/${projectId}/invoices/${inv.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: "sent" }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal mengubah status invoice");
      }
      setData((prev) =>
        prev.map((i) => (i.id === result.invoices.id ? result.invoices : i))
      );
      toast.success("Invoice ditandai terkirim");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    }
  };

  const openEdit = (inv: Invoice) => {
    setEditInvoice(inv);
  };

  const handleEdit = async (values: InvoiceFormValues) => {
    if (!projectId || !editInvoice) return;

    try {
      setUpdating(true);
      const response = await fetch(`/api/projects/${projectId}/invoices/${editInvoice.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(sanitizeValues(values)),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal mengubah invoice");
      }
      setData((prev) =>
        prev.map((i) => (i.id === result.invoices.id ? result.invoices : i))
      );
      setEditInvoice(null);
      toast.success("Invoice berhasil diubah");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    } finally {
      setUpdating(false);
    }
  };

  const handleRevert = async () => {
    if (!projectId || !revertInvoice) return;

    try {
      setUpdating(true);
      const response = await fetch(`/api/projects/${projectId}/invoices/${revertInvoice.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: "draft" }),
      });
      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.error || "Gagal mengubah status invoice");
      }
      setData((prev) =>
        prev.map((i) => (i.id === result.invoices.id ? result.invoices : i))
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
      accessorKey: "number",
      header: "No. Invoice",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.number}</span>
      ),
    },
    {
      accessorKey: "date",
      header: "Tanggal",
      cell: ({ row }) => formatDateShort(row.original.date),
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
          {row.original.status === "draft" && (
            <>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleMarkSent(row.original)}
              >
                <Send className="w-4 h-4 mr-1" />
                Tandai Terkirim
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => openEdit(row.original)}
              >
                <Pencil className="w-4 h-4 mr-1" />
                Edit
              </Button>
            </>
          )}
          {row.original.status === "sent" && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setMarkPaid(row.original)}
            >
              <CheckCircle2 className="w-4 h-4 mr-1" />
              Tandai Lunas
            </Button>
          )}
          {row.original.status === "paid" && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setRevertInvoice(row.original)}
            >
              <Undo2 className="w-4 h-4 mr-1" />
              Batalkan Lunas
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

      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>Buat Invoice</DialogTitle>
          </DialogHeader>
          <InvoiceForm
            onSubmit={handleCreate}
            onCancel={() => setShowForm(false)}
            submitting={saving}
            statusField
          />
        </DialogContent>
      </Dialog>

      <Dialog open={editInvoice !== null} onOpenChange={(o) => !o && setEditInvoice(null)}>
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>Edit Invoice {editInvoice?.number}</DialogTitle>
          </DialogHeader>
          <InvoiceForm
            invoice={editInvoice}
            onSubmit={handleEdit}
            onCancel={() => setEditInvoice(null)}
            submitting={updating}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={markPaid !== null} onOpenChange={(o) => !o && setMarkPaid(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Tandai Invoice Lunas</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              {markPaid?.number} · {formatCurrency(markPaid?.total || 0)} — yakin
              menandai invoice ini sebagai lunas?
            </p>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={recordExpense}
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
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Batalkan Lunas</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-gray-600">
            {revertInvoice?.number} akan dikembalikan ke status draft. Jika ada
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
