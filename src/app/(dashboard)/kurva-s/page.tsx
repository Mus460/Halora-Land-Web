"use client";
import { useProject } from "@/contexts/ProjectContext";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { FileDown } from "lucide-react";
import toast from "react-hot-toast";

export default function KurvaSPage() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const { currentProyekId: proyekId } = useProject();

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/proyek/${proyekId}/kurva-s`);
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

  return (
    <div className="space-y-6">
      <PageHeader
        title="Kurva S"
        description="Perbandingan rencana vs realisasi proyek"
        actions={
          <Button variant="outline">
            <FileDown className="w-4 h-4 mr-2" />
            Export Gambar
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle>Kurva S - Planned vs Actual</CardTitle>
        </CardHeader>
        <CardContent>
          {/* Simple chart placeholder */}
          <div className="h-[400px] flex items-end gap-2 p-4 border rounded-lg">
            {data.labels.map((label: string, i: number) => {
              const maxVal = Math.max(...data.planned, ...data.actual);
              const plannedHeight =
                maxVal > 0 ? (data.planned[i] / maxVal) * 350 : 0;
              const actualHeight =
                maxVal > 0 ? (data.actual[i] / maxVal) * 350 : 0;

              return (
                <div key={label} className="flex-1 flex flex-col items-center gap-1">
                  <div className="w-full flex gap-1 items-end justify-center">
                    <div
                      className="flex-1 bg-amber-200 rounded-t"
                      style={{ height: `${plannedHeight}px` }}
                    />
                    <div
                      className="flex-1 bg-amber-500 rounded-t"
                      style={{ height: `${actualHeight}px` }}
                    />
                  </div>
                  <span className="text-xs text-gray-500">{label}</span>
                </div>
              );
            })}
          </div>

          {/* Legend */}
          <div className="flex items-center justify-center gap-6 mt-4">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-amber-200 rounded" />
              <span className="text-sm text-gray-600">Planned</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-amber-500 rounded" />
              <span className="text-sm text-gray-600">Actual</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Data Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Data Kurva S</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left py-2 px-4">Bulan</th>
                <th className="text-right py-2 px-4">Planned (%)</th>
                <th className="text-right py-2 px-4">Actual (%)</th>
                <th className="text-right py-2 px-4">Deviasi</th>
              </tr>
            </thead>
            <tbody>
              {data.labels.map((label: string, i: number) => {
                const deviasi = data.actual[i] - data.planned[i];
                return (
                  <tr key={label} className="border-b">
                    <td className="py-2 px-4 font-medium">{label}</td>
                    <td className="py-2 px-4 text-right">
                      {data.planned[i]}%
                    </td>
                    <td className="py-2 px-4 text-right">
                      {data.actual[i]}%
                    </td>
                    <td
                      className={`py-2 px-4 text-right font-semibold ${
                        deviasi >= 0 ? "text-emerald-600" : "text-red-600"
                      }`}
                    >
                      {deviasi > 0 ? "+" : ""}
                      {deviasi}%
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  );
}
