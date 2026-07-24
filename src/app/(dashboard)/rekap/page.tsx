"use client";

import { useState } from "react";
import { Calculator, FileDown, Settings } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatCurrency } from "@/lib/utils";
import { getRekap, getAllPekerjaan } from "@/mock";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function RekapPage() {
  const rekapItems = getRekap();
  const pekerjaan = getAllPekerjaan(1);
  const [showMargin, setShowMargin] = useState(false);
  const [margin, setMargin] = useState(10);

  const subtotal = pekerjaan.reduce((sum, p) => sum + p.totalBiaya, 0);
  const overhead = subtotal * 0.1;
  const profit = (subtotal + overhead) * (margin / 100);
  const ppn = (subtotal + overhead + profit) * 0.11;
  const total = subtotal + overhead + profit + ppn;

  const grouped = pekerjaan.reduce((acc, item) => {
    if (!acc[item.kategori]) acc[item.kategori] = [];
    acc[item.kategori].push(item);
    return acc;
  }, {} as Record<string, typeof pekerjaan>);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Rekapitulasi RAB"
        description="Ringkasan Rencana Anggaran Biaya proyek"
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowMargin(true)}>
              <Settings className="w-4 h-4 mr-2" />
              Margin ({margin}%)
            </Button>
            <Button className="bg-amber-500 hover:bg-amber-600">
              <FileDown className="w-4 h-4 mr-2" />
              Export PDF
            </Button>
          </div>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Table */}
        <div className="lg:col-span-2 space-y-4">
          {Object.entries(grouped).map(([kategori, items]) => {
            const kategoriTotal = items.reduce(
              (sum, item) => sum + item.totalBiaya,
              0
            );
            return (
              <Card key={kategori}>
                <CardHeader className="py-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-semibold uppercase text-gray-700">
                      {kategori}
                    </CardTitle>
                    <Badge variant="outline">{formatCurrency(kategoriTotal)}</Badge>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-gray-50">
                        <th className="text-left px-4 py-2 font-medium text-gray-600">
                          Uraian
                        </th>
                        <th className="text-right px-4 py-2 font-medium text-gray-600">
                          Volume
                        </th>
                        <th className="text-right px-4 py-2 font-medium text-gray-600">
                          Harga Satuan
                        </th>
                        <th className="text-right px-4 py-2 font-medium text-gray-600">
                          Total
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {items.map((item) => (
                        <tr key={item.id} className="border-b last:border-0">
                          <td className="px-4 py-2">
                            <p className="font-medium">{item.uraianPekerjaan}</p>
                            {item.levelPekerjaan && (
                              <p className="text-xs text-gray-500">
                                {item.levelPekerjaan}
                              </p>
                            )}
                          </td>
                          <td className="text-right px-4 py-2 text-gray-600">
                            {item.volume} {item.satuan}
                          </td>
                          <td className="text-right px-4 py-2 text-gray-600">
                            {formatCurrency(item.hargaSatuan)}
                          </td>
                          <td className="text-right px-4 py-2 font-semibold">
                            {formatCurrency(item.totalBiaya)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </CardContent>
              </Card>
            );
          })}

          {pekerjaan.length === 0 && (
            <Card>
              <CardContent className="py-12 text-center text-gray-500">
                <Calculator className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p>Belum ada item pekerjaan di RAB.</p>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Summary Sidebar */}
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Ringkasan RAB</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Subtotal</span>
                <span className="font-semibold">{formatCurrency(subtotal)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Overhead (10%)</span>
                <span>{formatCurrency(overhead)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Profit ({margin}%)</span>
                <span>{formatCurrency(profit)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">PPN (11%)</span>
                <span>{formatCurrency(ppn)}</span>
              </div>
              <div className="border-t pt-3 flex justify-between">
                <span className="font-bold text-gray-900">Grand Total</span>
                <span className="font-bold text-lg text-amber-600">
                  {formatCurrency(total)}
                </span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="text-center">
                <p className="text-sm text-gray-500">Total Item</p>
                <p className="text-2xl font-bold text-gray-900">
                  {pekerjaan.length}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Margin Dialog */}
      <Dialog open={showMargin} onOpenChange={setShowMargin}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Atur Margin Profit</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Margin (%)</Label>
              <Input
                type="number"
                value={margin}
                onChange={(e) => setMargin(Number(e.target.value) || 0)}
                min={0}
                max={100}
              />
            </div>
            <div className="p-3 bg-gray-50 rounded-lg text-sm">
              <div className="flex justify-between mb-1">
                <span>Profit</span>
                <span className="font-semibold">
                  {formatCurrency(profit)}
                </span>
              </div>
              <div className="flex justify-between font-bold">
                <span>Grand Total</span>
                <span className="text-amber-600">
                  {formatCurrency(total)}
                </span>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setShowMargin(false)}>Simpan</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
