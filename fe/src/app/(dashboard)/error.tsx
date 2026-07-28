"use client";

import { Button } from "@/components/ui/button";
import { AlertTriangle, RefreshCw } from "lucide-react";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex items-center justify-center min-h-[60vh] p-4">
      <div className="text-center space-y-4">
        <div className="flex justify-center">
          <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
            <AlertTriangle className="w-6 h-6 text-red-500" />
          </div>
        </div>
        <div className="space-y-1">
          <h2 className="text-lg font-semibold text-gray-900">
            Gagal Memuat Halaman
          </h2>
          <p className="text-sm text-gray-600 max-w-sm">
            {error.message || "Terjadi kesalahan saat memuat data."}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={reset}>
          <RefreshCw className="w-4 h-4 mr-2" />
          Muat Ulang
        </Button>
      </div>
    </div>
  );
}
