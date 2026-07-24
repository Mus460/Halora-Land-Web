"use client";

import { useState } from "react";
import { ClipboardCheck } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { getMonitoringData } from "@/mock";

export default function MonitoringPage() {
  const monitoring = getMonitoringData();

  const totalItems = monitoring.reduce(
    (sum, cat) => sum + cat.items.length,
    0
  );
  const completedItems = monitoring.reduce(
    (sum, cat) => sum + cat.items.filter((i) => i.progress === 100).length,
    0
  );
  const overallProgress =
    totalItems > 0 ? Math.round((completedItems / totalItems) * 100) : 0;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Progress Monitoring"
        description="Tracking progres pekerjaan per kategori"
      />

      {/* Overall Progress */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="w-14 h-14 bg-amber-100 rounded-full flex items-center justify-center">
              <ClipboardCheck className="w-7 h-7 text-amber-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Progres Keseluruhan</p>
              <p className="text-3xl font-bold text-gray-900">
                {overallProgress}%
              </p>
            </div>
          </div>
          <Progress value={overallProgress} className="h-3" />
          <p className="text-xs text-gray-500 mt-2">
            {completedItems} dari {totalItems} item selesai
          </p>
        </CardContent>
      </Card>

      {/* Per Kategori */}
      <div className="space-y-4">
        {monitoring.map((kategori) => {
          const kategoriTotal = kategori.items.length;
          const kategoriDone = kategori.items.filter(
            (i) => i.progress === 100
          ).length;
          const kategoriProgress =
            kategoriTotal > 0
              ? Math.round((kategoriDone / kategoriTotal) * 100)
              : 0;

          return (
            <Card key={kategori.kategori}>
              <CardHeader className="py-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-semibold">
                    {kategori.kategori}
                  </CardTitle>
                  <Badge variant="outline">
                    {kategoriDone}/{kategoriTotal} selesai
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <Progress value={kategoriProgress} className="h-2" />
                {kategori.items.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between py-2 border-b last:border-0"
                  >
                    <div>
                      <p className="text-sm font-medium">{item.uraian}</p>
                      <p className="text-xs text-gray-500">
                        {item.volume} {item.satuan}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="w-24">
                        <Progress value={item.progress} className="h-1.5" />
                      </div>
                      <span className="text-sm font-semibold w-10 text-right">
                        {item.progress}%
                      </span>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
