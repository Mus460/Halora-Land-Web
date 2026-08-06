"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatCurrency, formatWaktu } from "@/lib/utils";
import type { MasterAnalisa, RincianAnalisa } from "@/types";
import toast from "react-hot-toast";

const tipeLabel: Record<string, string> = {
  upah: "Tenaga Kerja",
  material: "Bahan",
  alat: "Peralatan",
};

export default function MasterAnalisaDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [item, setItem] = useState<MasterAnalisa | null>(null);
  const [rincian, setRincian] = useState<RincianAnalisa[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!params?.id) return;
    (async () => {
      try {
        setLoading(true);
        const [metaRes, rinRes] = await Promise.all([
          fetch(`/api/master-analisa/${params.id}`),
          fetch(`/api/master-analisa/${params.id}/rincian`),
        ]);
        if (!metaRes.ok) throw new Error("Gagal memuat analisa");
        setItem(await metaRes.json());
        if (rinRes.ok) setRincian(await rinRes.json());
      } catch (error) {
        console.error(error);
        toast.error("Gagal memuat detail analisa");
      } finally {
        setLoading(false);
      }
    })();
  }, [params?.id]);

  if (loading) {
    return <p className="p-8 text-center text-gray-500">Memuat data...</p>;
  }
  if (!item) {
    return <p className="p-8 text-center text-gray-500">Analisa tidak ditemukan</p>;
  }

  const totalJumlah = rincian.reduce((sum, r) => sum + (r.jumlahHarga || 0), 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Detail Analisa"
        description="Rincian komponen analisa harga satuan pekerjaan"
      />

      <Card>
        <CardContent className="p-4 space-y-3">
          <button
            onClick={() => router.push("/master-analisa")}
            className="inline-flex items-center gap-1 text-sm text-amber-600 hover:text-amber-700"
          >
            <ArrowLeft className="w-4 h-4" /> Kembali
          </button>
          <div className="flex items-center gap-2">
            <span className="text-xs text-gray-400 font-mono">{item.kode}</span>
            <h2 className="font-semibold">{item.nama}</h2>
          </div>
          <div className="flex flex-wrap gap-2">
            {item.satuan && <Badge variant="outline">{item.satuan}</Badge>}
            {item.hargaSatuan != null && item.hargaSatuan > 0 && (
              <Badge variant="outline" className="text-amber-700">
                {formatCurrency(item.hargaSatuan)}
              </Badge>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-xs uppercase text-gray-500">
                  <th className="px-4 py-3 w-10">No</th>
                  <th className="px-4 py-3">Uraian</th>
                  <th className="px-4 py-3">Kode</th>
                  <th className="px-4 py-3">Satuan</th>
                  <th className="px-4 py-3 text-right">Koefisien</th>
                  <th className="px-4 py-3 text-right">Harga Satuan</th>
                  <th className="px-4 py-3 text-right">Jumlah Harga</th>
                  <th className="px-4 py-3 text-right">Waktu (1/koef)</th>
                </tr>
              </thead>
              <tbody>
                {rincian.length === 0 && (
                  <tr>
                    <td colSpan={8} className="px-4 py-6 text-center text-gray-500">
                      Tidak ada rincian
                    </td>
                  </tr>
                )}
                {rincian.map((r, idx) => (
                  <tr key={r.id} className="border-b last:border-0">
                    <td className="px-4 py-2.5 text-gray-500">{idx + 1}</td>
                    <td className="px-4 py-2.5 max-w-[280px]">
                      <span
                        className={
                          r.tipe === "upah" || r.tipe === "material" || r.tipe === "alat"
                            ? "font-medium block truncate"
                            : "block truncate"
                        }
                      >
                        {r.nama || "-"}
                      </span>
                      {tipeLabel[r.tipe] && (
                        <span className="ml-2 text-xs text-gray-400">
                          {tipeLabel[r.tipe]}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-2.5 font-mono text-xs text-gray-500">
                      {r.kodeReferensi || "-"}
                    </td>
                    <td className="px-4 py-2.5">{r.satuan || "-"}</td>
                    <td className="px-4 py-2.5 text-right font-mono">{r.koef}</td>
                    <td className="px-4 py-2.5 text-right">
                      {formatCurrency(r.hargaSatuan || 0)}
                    </td>
                    <td className="px-4 py-2.5 text-right font-medium">
                      {formatCurrency(r.jumlahHarga || 0)}
                    </td>
                    <td className="px-4 py-2.5 text-right font-mono">
                      {formatWaktu(r.waktu)}
                    </td>
                  </tr>
                ))}
              </tbody>
              {rincian.length > 0 && (
                <tfoot>
                  <tr className="border-t font-semibold">
                    <td colSpan={6} className="px-4 py-3 text-right">
                      Total (A+B+C)
                    </td>
                    <td className="px-4 py-3 text-right">{formatCurrency(totalJumlah)}</td>
                    <td />
                  </tr>
                </tfoot>
              )}
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
