"use client";
import { useProject } from "@/contexts/ProjectContext";

import { useState, useEffect } from "react";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
} from "chart.js";
import { Line } from "react-chartjs-2";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { FileDown } from "lucide-react";
import toast from "react-hot-toast";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
);

export default function KurvaSPage() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const { currentProjectId: projectId } = useProject();

  useEffect(() => {
    fetchData();
  }, [projectId]);

  const fetchData = async () => {
    if (!projectId) return;
    try {
      setLoading(true);
      const response = await fetch(`/api/projects/${projectId}/s-curve`);
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

  const chartData = {
    labels: data.labels,
    datasets: [
      {
        label: "Planned",
        data: data.planned,
        borderColor: "#d97706",
        borderDash: [6, 4],
        backgroundColor: "rgba(217, 119, 6, 0.05)",
        fill: true,
        tension: 0.35,
        pointRadius: 3,
        pointBackgroundColor: "#d97706",
      },
      {
        label: "Actual",
        data: data.actual,
        borderColor: "#f59e0b",
        backgroundColor: "rgba(245, 158, 11, 0.05)",
        fill: false,
        tension: 0.35,
        pointRadius: 4,
        pointBackgroundColor: "#f59e0b",
      },
    ],
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: "index" as const, intersect: false },
    scales: {
      y: {
        beginAtZero: true,
        ticks: { callback: (value: number | string) => `${value}%` },
      },
    },
    plugins: {
      legend: { position: "bottom" as const },
      tooltip: { callbacks: { label: (ctx: any) => ` ${ctx.dataset.label}: ${ctx.parsed.y}%` } },
    },
  };

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
          <div className="h-[400px]">
            <Line data={chartData} options={chartOptions} />
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
