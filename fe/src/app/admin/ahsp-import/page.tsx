"use client";

import { useState } from "react";
import { Upload, FileSpreadsheet, CheckCircle, AlertCircle } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import toast from "react-hot-toast";

export default function AHSPImportPage() {
  const [file, setFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [result, setResult] = useState<{
    success: boolean;
    message: string;
    count?: number;
  } | null>(null);

  const handleImport = async () => {
    if (!file) return;

    setImporting(true);
    setResult(null);

    // Simulate import
    await new Promise((resolve) => setTimeout(resolve, 2000));

    setResult({
      success: true,
      message: "Import berhasil!",
      count: 1250,
    });
    setImporting(false);
    toast.success("AHSP berhasil diimport");
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Import AHSP"
        description="Import database AHSP dari file Excel"
      />

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileSpreadsheet className="w-5 h-5" />
            Upload File Excel
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center">
            <Upload className="w-12 h-12 mx-auto text-gray-400 mb-4" />
            <p className="text-sm text-gray-600 mb-2">
              Drag & drop file Excel (.xlsx, .xlsm) atau
            </p>
            <label>
              <Button variant="outline" className="cursor-pointer">
                Pilih File
              </Button>
              <input
                type="file"
                accept=".xlsx,.xlsm,.xls"
                className="hidden"
                onChange={(e) => setFile(e.target.files?.[0] || null)}
              />
            </label>
            {file && (
              <p className="text-sm text-amber-600 mt-3">
                File: {file.name}
              </p>
            )}
          </div>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Pastikan format file sesuai dengan template AHSP PUPR 2026.
              File harus memiliki sheet: Master Analisa, Rincian Analisa, dan
              Master Harga.
            </AlertDescription>
          </Alert>

          <Button
            onClick={handleImport}
            disabled={!file || importing}
            className="bg-amber-500 hover:bg-amber-600"
          >
            {importing ? "Mengimport..." : "Mulai Import"}
          </Button>

          {result && (
            <Alert
              className={
                result.success
                  ? "border-emerald-200 bg-emerald-50"
                  : "border-red-200 bg-red-50"
              }
            >
              {result.success ? (
                <CheckCircle className="h-4 w-4 text-emerald-600" />
              ) : (
                <AlertCircle className="h-4 w-4 text-red-600" />
              )}
              <AlertDescription>
                {result.message}
                {result.count && (
                  <span className="font-semibold"> ({result.count} item)</span>
                )}
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
