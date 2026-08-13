"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { CurrencyInput } from "@/components/shared/currency-input";
import { formatCurrency } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import toast from "react-hot-toast";
import type { Invoice, InvoiceItem, Client } from "@/types";

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
  status?: "draft" | "sent";
}

interface InvoiceFormProps {
  invoice?: Invoice | null;
  onSubmit: (values: InvoiceFormValues) => void;
  onCancel: () => void;
  submitting?: boolean;
  statusField?: boolean;
}

const emptyItem = (): InvoiceItem => ({
  description: "",
  qty: 1,
  unit: "ls",
  unitPrice: 0,
});

export function InvoiceForm({
  invoice,
  onSubmit,
  onCancel,
  submitting,
  statusField,
}: InvoiceFormProps) {
  const [values, setValues] = useState<InvoiceFormValues>(() => {
    const date = invoice?.date ? invoice.date.slice(0, 10) : new Date().toISOString().slice(0, 10);
    const due = invoice?.dueDate ? invoice.dueDate.slice(0, 10) : "";
    return {
      date,
      dueDate: due || addDays(date, 14),
      poNumber: invoice?.poNumber || "",
      buyerName: invoice?.buyerName || "",
      buyerAddress: invoice?.buyerAddress || "",
      buyerContact: invoice?.buyerContact || "",
      discount: Number(invoice?.discount) || 0,
      taxRate: Number(invoice?.taxRate) || 11,
      paymentBank: invoice?.paymentBank || "",
      paymentAccountNumber: invoice?.paymentAccountNumber || "",
      paymentAccountName: invoice?.paymentAccountName || "",
      notes: invoice?.notes || "",
      financeName: invoice?.financeName || "",
      items:
        invoice?.items && invoice.items.length > 0
          ? invoice.items.map((i) => ({ ...i }))
          : [emptyItem()],
    };
  });
  const [status, setStatus] = useState<"draft" | "sent">("draft");
  const [clients, setClients] = useState<Client[]>([]);
  const [clientId, setClientId] = useState("");
  const [showClientForm, setShowClientForm] = useState(false);

  useEffect(() => {
    fetch("/api/clients")
      .then((res) => (res.ok ? res.json() : { clients: [] }))
      .then((data) => setClients(data.clients || []))
      .catch(() => console.error("Failed to load clients"));
  }, []);

  const handlePickClient = (id: string | null) => {
    if (!id) return;
    const c = clients.find((x) => x.id === Number(id));
    if (!c) return;
    setClientId(id);
    set("buyerName", c.name);
    set("buyerAddress", c.address || "");
    set("buyerContact", c.contact || "");
  };

  const handleClientCreated = (c: Client) => {
    setClients((prev) => [...prev, c]);
    setClientId(String(c.id));
    set("buyerName", c.name);
    set("buyerAddress", c.address || "");
    set("buyerContact", c.contact || "");
  };

  const set = <K extends keyof InvoiceFormValues>(key: K, value: InvoiceFormValues[K]) =>
    setValues((prev) => ({ ...prev, [key]: value }));

  const updateItem = (index: number, patch: Partial<InvoiceItem>) => {
    setValues((prev) => ({
      ...prev,
      items: prev.items.map((item, i) => (i === index ? { ...item, ...patch } : item)),
    }));
  };

  const addItem = () => setValues((prev) => ({ ...prev, items: [...prev.items, emptyItem()] }));
  const removeItem = (index: number) =>
    setValues((prev) => ({
      ...prev,
      items: prev.items.length > 1 ? prev.items.filter((_, i) => i !== index) : prev.items,
    }));

  const subtotal = values.items.reduce(
    (sum, item) => sum + (Number(item.qty) || 0) * (Number(item.unitPrice) || 0),
    0
  );
  const taxable = Math.max(subtotal - values.discount, 0);
  const tax = (taxable * values.taxRate) / 100;
  const total = taxable + tax;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({ ...values, items: values.items, status });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-h-[70vh] overflow-y-auto pr-1">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="inv-date">Tanggal *</Label>
          <Input
            id="inv-date"
            type="date"
            value={values.date}
            onChange={(e) => set("date", e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="inv-due">Jatuh Tempo</Label>
          <Input
            id="inv-due"
            type="date"
            value={values.dueDate}
            onChange={(e) => set("dueDate", e.target.value)}
          />
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="inv-po">No. PO (jika ada)</Label>
          <Input
            id="inv-po"
            placeholder="PO-2026-001"
            value={values.poNumber}
            onChange={(e) => set("poNumber", e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-3 rounded-lg border border-gray-200 p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Label className="text-sm font-semibold">Informasi Pembeli (Bill To)</Label>
          <div className="ml-auto flex items-center gap-2">
            <Select
              value={clientId || undefined}
              onValueChange={handlePickClient}
            >
              <SelectTrigger className="w-56">
                <SelectValue placeholder="Pilih klien tersimpan..." />
              </SelectTrigger>
              <SelectContent>
                {clients.length === 0 && (
                  <SelectItem value="__none__" disabled>
                    Belum ada klien
                  </SelectItem>
                )}
                {clients.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setShowClientForm(true)}
            >
              <Plus className="w-4 h-4 mr-1" />
              Baru
            </Button>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="inv-buyer">Nama Klien / Perusahaan *</Label>
            <Input
              id="inv-buyer"
              placeholder="PT Klien Sejahtera"
              value={values.buyerName}
              onChange={(e) => set("buyerName", e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="inv-pic">Kontak PIC</Label>
            <Input
              id="inv-pic"
              placeholder="Ibu Rina (Finance) 0812-xxxx"
              value={values.buyerContact}
              onChange={(e) => set("buyerContact", e.target.value)}
            />
          </div>
          <div className="space-y-2 md:col-span-2">
            <Label htmlFor="inv-buyer-address">Alamat Penagihan *</Label>
            <Input
              id="inv-buyer-address"
              placeholder="Jl. Melati No. 12, Jakarta"
              value={values.buyerAddress}
              onChange={(e) => set("buyerAddress", e.target.value)}
              required
            />
          </div>
        </div>
      </div>

      <div className="space-y-3 rounded-lg border border-gray-200 p-4">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">Rincian Produk/Jasa</Label>
          <Button type="button" variant="outline" size="sm" onClick={addItem}>
            <Plus className="w-4 h-4 mr-1" />
            Tambah Item
          </Button>
        </div>
        <div className="space-y-2">
          {values.items.map((item, index) => (
            <div key={index} className="grid grid-cols-12 gap-2 items-end">
              <div className="col-span-12 md:col-span-4 space-y-1">
                <Label className="text-xs">Deskripsi *</Label>
                <Input
                  placeholder="Pekerjaan..."
                  value={item.description}
                  onChange={(e) => updateItem(index, { description: e.target.value })}
                  required
                />
              </div>
              <div className="col-span-3 md:col-span-1 space-y-1">
                <Label className="text-xs">Qty</Label>
                <Input
                  type="number"
                  min={0}
                  step="0.01"
                  value={item.qty}
                  onChange={(e) => updateItem(index, { qty: Number(e.target.value) })}
                />
              </div>
              <div className="col-span-3 md:col-span-1 space-y-1">
                <Label className="text-xs">Unit</Label>
                <Input
                  placeholder="m2"
                  value={item.unit}
                  onChange={(e) => updateItem(index, { unit: e.target.value })}
                />
              </div>
              <div className="col-span-3 md:col-span-3 space-y-1">
                <Label className="text-xs">Harga Satuan</Label>
                <Input
                  type="number"
                  min={0}
                  value={item.unitPrice}
                  onChange={(e) => updateItem(index, { unitPrice: Number(e.target.value) })}
                />
              </div>
              <div className="col-span-2 md:col-span-2 space-y-1">
                <Label className="text-xs">Total</Label>
                <p className="text-sm font-semibold tabular-nums truncate">
                  {formatCurrency((Number(item.qty) || 0) * (Number(item.unitPrice) || 0))}
                </p>
              </div>
              <div className="col-span-1 flex justify-end">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-gray-400 hover:text-red-500"
                  onClick={() => removeItem(index)}
                  title="Hapus item"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <CurrencyInput
          label="Diskon (Rp)"
          value={values.discount}
          onChange={(v) => set("discount", v)}
        />
        <div className="space-y-2">
          <Label htmlFor="inv-tax">Pajak PPN (%)</Label>
          <Input
            id="inv-tax"
            type="number"
            min={0}
            step="0.01"
            value={values.taxRate}
            onChange={(e) => set("taxRate", Number(e.target.value))}
          />
        </div>
      </div>

      <div className="space-y-2 rounded-lg bg-gray-50 p-4 text-sm">
        <div className="flex justify-between text-gray-600">
          <span>Subtotal</span>
          <span className="tabular-nums">{formatCurrency(subtotal)}</span>
        </div>
        <div className="flex justify-between text-gray-600">
          <span>Diskon</span>
          <span className="tabular-nums">- {formatCurrency(values.discount)}</span>
        </div>
        <div className="flex justify-between text-gray-600">
          <span>PPN ({values.taxRate}%)</span>
          <span className="tabular-nums">{formatCurrency(tax)}</span>
        </div>
        <div className="flex justify-between font-bold text-gray-900 border-t border-gray-200 pt-2">
          <span>Grand Total</span>
          <span className="tabular-nums">{formatCurrency(total)}</span>
        </div>
      </div>

      <div className="space-y-3 rounded-lg border border-gray-200 p-4">
        <Label className="text-sm font-semibold">Instruksi Pembayaran</Label>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="space-y-2">
            <Label htmlFor="inv-bank">Bank *</Label>
            <Input
              id="inv-bank"
              placeholder="BCA"
              value={values.paymentBank}
              onChange={(e) => set("paymentBank", e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="inv-accno">No. Rekening *</Label>
            <Input
              id="inv-accno"
              placeholder="123-456-7890"
              value={values.paymentAccountNumber}
              onChange={(e) => set("paymentAccountNumber", e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="inv-accname">Atas Nama *</Label>
            <Input
              id="inv-accname"
              placeholder="Halora Land"
              value={values.paymentAccountName}
              onChange={(e) => set("paymentAccountName", e.target.value)}
              required
            />
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="inv-notes">Catatan</Label>
          <Textarea
            id="inv-notes"
            rows={2}
            placeholder="Syarat pelunasan, berita transfer, dll."
            value={values.notes}
            onChange={(e) => set("notes", e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="inv-finance">Penandatangan (Bagian Keuangan) *</Label>
          <Input
            id="inv-finance"
            placeholder="Nama bagian keuangan"
            value={values.financeName}
            onChange={(e) => set("financeName", e.target.value)}
            required
          />
        </div>
      </div>

      {statusField && (
        <div className="space-y-2">
          <Label>Status</Label>
          <div className="flex gap-2">
            {(["draft", "sent"] as const).map((s) => (
              <Button
                key={s}
                type="button"
                variant={status === s ? "default" : "outline"}
                className={status === s ? "bg-amber-500 hover:bg-amber-600" : ""}
                onClick={() => setStatus(s)}
              >
                {s === "draft" ? "Draft" : "Terkirim"}
              </Button>
            ))}
          </div>
        </div>
      )}

      <ClientFormDialog
        open={showClientForm}
        onOpenChange={setShowClientForm}
        onCreated={handleClientCreated}
      />

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Batal
        </Button>
        <Button
          type="submit"
          disabled={submitting}
          className="bg-amber-500 hover:bg-amber-600"
        >
          {submitting ? "Menyimpan..." : invoice ? "Simpan Perubahan" : "Simpan"}
        </Button>
      </div>
    </form>
  );
}

function addDays(date: string, days: number): string {
  const d = new Date(date);
  if (isNaN(d.getTime())) return date;
  d.setDate(d.getDate() + days);
  return d.toISOString().slice(0, 10);
}

function ClientFormDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: (client: Client) => void;
}) {
  const [form, setForm] = useState({ name: "", address: "", contact: "" });
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      const response = await fetch("/api/clients", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Create failed");
      }
      const created = await response.json();
      onCreated(created);
      onOpenChange(false);
      setForm({ name: "", address: "", contact: "" });
      toast.success("Klien berhasil disimpan");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menyimpan klien");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Tambah Klien</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="new-client-name">Nama *</Label>
            <Input
              id="new-client-name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Nama klien / perusahaan"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-client-address">Alamat</Label>
            <Input
              id="new-client-address"
              value={form.address}
              onChange={(e) => setForm({ ...form, address: e.target.value })}
              placeholder="Alamat tagihan"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-client-contact">Kontak</Label>
            <Input
              id="new-client-contact"
              value={form.contact}
              onChange={(e) => setForm({ ...form, contact: e.target.value })}
              placeholder="PIC / No. telepon"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={saving}
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
  );
}
