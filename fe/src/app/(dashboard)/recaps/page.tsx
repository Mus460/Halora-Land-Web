"use client";

import { useState, useEffect } from "react";
import { Calculator, FileDown } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatCurrency, formatVolume } from "@/lib/utils";
import toast from "react-hot-toast";
import { useProject } from "@/contexts/ProjectContext";
import { EmptyProjectState } from "@/components/shared/empty-project-state";

export default function RekapPage() {
  const { currentProjectId: projectId, projectList, loading: proyekLoading } = useProject();
  
  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (projectList.length === 0) {
    return (
      <EmptyProjectState
        title="Belum Ada Data Rekapitulasi"
        description="Buat proyek dan tambahkan pekerjaan untuk melihat rekapitulasi biaya"
      />
    );
  }
  
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);
  const [ownerName, setOwnerName] = useState("");

  useEffect(() => {
    fetch("/api/auth/me")
      .then((r) => r.json())
      .then((res) => setOwnerName(res?.user?.fullName || ""))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (projectId) {
      fetchData();
    }
  }, [projectId]);

  const fetchData = async () => {
    if (!projectId) return;
    
    try {
      setLoading(true);
      const response = await fetch(`/api/projects/${projectId}/recaps`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setData(result);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  if (loading || !data) {
    return <div className="p-8 text-center">Memuat data...</div>;
  }

  const grouped = data.grouped || {};
  const summary = data.summary || {};
  const subtotal = Number(summary.subtotal || 0);
  const ppn = Number(summary.totalPPN || 0);
  const total = Number(summary.totalFinal || 0);
  const rekapItemCount = Object.values(grouped).reduce(
    (sum: number, items: any) => sum + (Array.isArray(items) ? items.length : 0),
    0
  );

  const handleExportPDF = async () => {
    try {
      setExporting(true);
      const { exportRekapPDF } = await import("@/lib/export-rekap-pdf");
      exportRekapPDF(data, { ownerName });
      toast.success('PDF berhasil diunduh');
    } catch (error) {
      console.error('Export PDF error:', error);
      toast.error('Gagal membuat PDF');
    } finally {
      setExporting(false);
    }
  };

return (
    <div className="space-y-6">
      <PageHeader
        title="Rekapitulasi RAB"
        description="Ringkasan Rencana Anggaran Biaya proyek"
        actions={
          <div className="flex gap-2">
            <Button
              className="bg-amber-500 hover:bg-amber-600"
              onClick={handleExportPDF}
              disabled={exporting}
            >
              <FileDown className="w-4 h-4 mr-2" />
              {exporting ? "Menyiapkan..." : "Export PDF"}
            </Button>
          </div>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Table */}
        <div className="lg:col-span-2 space-y-4">
          {Object.entries(grouped).map(([category, items]) => {
            const itemsArray = items as any[];
            const kategoriTotal = itemsArray.reduce(
              (sum, item) => sum + Number(item.totalCost),
              0
            );
            return (
              <Card key={category}>
                <CardHeader className="py-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-semibold uppercase text-gray-700">
                      {category}
                    </CardTitle>
                    <Badge variant="outline">{formatCurrency(kategoriTotal)}</Badge>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-gray-50">
                        <th className="text-left px-4 py-2 font-medium text-gray-600">
                          Description
                        </th>
                        <th className="text-right px-4 py-2 font-medium text-gray-600">
                          Volume
                        </th>
                        <th className="text-right px-4 py-2 font-medium text-gray-600">
                          Price Unit
                        </th>
                        <th className="text-right px-4 py-2 font-medium text-gray-600">
                          Total
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {itemsArray.map((item: any) => (
                        <tr key={item.id} className="border-b last:border-0">
                          <td className="px-4 py-2 max-w-[320px]">
                            <p className="font-medium truncate">{item.description}</p>
														{/* {item.level && (
                            //   <p className="text-xs text-gray-500">
                            //     {item.level}
                            //   </p>
                            // )}
													  */}
                          </td>
                          <td className="text-right px-4 py-2 text-gray-600">
                            {formatVolume(item.volume)} {item.unit}
                          </td>
                          <td className="text-right px-4 py-2 text-gray-600">
                            {formatCurrency(Number(item.unitPrice))}
                          </td>
                          <td className="text-right px-4 py-2 font-semibold">
                            {formatCurrency(Number(item.totalCost))}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </CardContent>
              </Card>
            );
          })}

          {rekapItemCount === 0 && (
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
                <span className="text-gray-600">PPN</span>
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
                  {rekapItemCount}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      </div>
  );
}
