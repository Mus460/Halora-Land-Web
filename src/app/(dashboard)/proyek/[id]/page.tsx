"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  Building2,
  MapPin,
  Calendar,
  DollarSign,
  Users,
  FileText,
  TrendingUp,
  ArrowLeft,
  Pencil,
  Copy,
  Trash2,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { formatCurrency } from "@/lib/utils";
import toast from "react-hot-toast";

interface ProyekDetail {
  id: number;
  namaProyek: string;
  lokasi: string | null;
  tipe: string;
  nilaiKontrak: number | null;
  timeline: string | null;
  createdAt: string;
  user: {
    id: number;
    namaLengkap: string;
    email: string;
  };
  timProyek: Array<{
    id: number;
    role: string;
    user: {
      id: number;
      namaLengkap: string;
      email: string;
    };
  }>;
  pekerjaan: Array<{
    id: number;
    uraianPekerjaan: string;
    volume: number;
    satuan: string;
    hargaSatuan: number;
    totalBiaya: number;
    kategori: string;
  }>;
  _count: {
    pekerjaan: number;
    rekap: number;
    invoice: number;
  };
}

export default function ProyekDetailPage() {
  const params = useParams();
  const router = useRouter();
  const [proyek, setProyek] = useState<ProyekDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetchProyekDetail();
  }, [params.id]);

  const fetchProyekDetail = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/proyek/${params.id}`);
      
      if (!response.ok) {
        if (response.status === 404) {
          toast.error("Proyek tidak ditemukan");
          router.push("/proyek");
          return;
        }
        if (response.status === 403) {
          toast.error("Anda tidak memiliki akses ke proyek ini");
          router.push("/proyek");
          return;
        }
        throw new Error("Gagal memuat detail proyek");
      }

      const data = await response.json();
      setProyek(data.proyek);
    } catch (error) {
      console.error("Error fetching proyek:", error);
      toast.error("Gagal memuat detail proyek");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Yakin ingin menghapus proyek ini?")) return;

    try {
      const response = await fetch(`/api/proyek/${params.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error("Gagal menghapus proyek");
      }

      toast.success("Proyek berhasil dihapus");
      router.push("/proyek");
    } catch (error) {
      console.error("Error deleting proyek:", error);
      toast.error("Gagal menghapus proyek");
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-amber-500 mx-auto"></div>
          <p className="mt-4 text-gray-600">Memuat detail proyek...</p>
        </div>
      </div>
    );
  }

  if (!proyek) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Building2 className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-600">Proyek tidak ditemukan</p>
          <Button
            onClick={() => router.push("/proyek")}
            className="mt-4"
            variant="outline"
          >
            Kembali ke Daftar Proyek
          </Button>
        </div>
      </div>
    );
  }

  const totalBiaya = proyek.pekerjaan.reduce(
    (sum, p) => sum + p.totalBiaya,
    0
  );

  return (
    <div className="container mx-auto p-6 max-w-7xl">
      {/* Header */}
      <div className="mb-6">
        <Button
          variant="ghost"
          onClick={() => router.push("/proyek")}
          className="mb-4"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Kembali
        </Button>

        <div className="flex items-start justify-between">
          <div className="flex items-start gap-4">
            <div className="w-16 h-16 bg-amber-100 rounded-xl flex items-center justify-center shrink-0">
              <Building2 className="w-8 h-8 text-amber-600" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">
                {proyek.namaProyek}
              </h1>
              <div className="flex items-center gap-3 text-sm text-gray-600">
                <Badge variant="outline">
                  {proyek.tipe === "gedung" ? "Gedung" : "Infrastruktur"}
                </Badge>
                {proyek.lokasi && (
                  <>
                    <span>•</span>
                    <div className="flex items-center gap-1">
                      <MapPin className="w-4 h-4" />
                      {proyek.lokasi}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>

          <div className="flex gap-2">
            <Button variant="outline" size="icon">
              <Pencil className="w-4 h-4" />
            </Button>
            <Button variant="outline" size="icon">
              <Copy className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              onClick={handleDelete}
              className="text-red-600 hover:text-red-700"
            >
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">Total Pekerjaan</p>
                <p className="text-2xl font-bold text-gray-900">
                  {proyek._count.pekerjaan}
                </p>
              </div>
              <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
                <FileText className="w-6 h-6 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">Total Biaya</p>
                <p className="text-2xl font-bold text-gray-900">
                  {formatCurrency(totalBiaya)}
                </p>
              </div>
              <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
                <DollarSign className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">Invoice</p>
                <p className="text-2xl font-bold text-gray-900">
                  {proyek._count.invoice}
                </p>
              </div>
              <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
                <TrendingUp className="w-6 h-6 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">Tim</p>
                <p className="text-2xl font-bold text-gray-900">
                  {proyek.timProyek.length + 1}
                </p>
              </div>
              <div className="w-12 h-12 bg-amber-100 rounded-lg flex items-center justify-center">
                <Users className="w-6 h-6 text-amber-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Project Info */}
          <Card>
            <CardHeader>
              <CardTitle>Informasi Proyek</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {proyek.timeline && (
                <div className="flex items-center gap-3">
                  <Calendar className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-600">Timeline</p>
                    <p className="font-medium text-gray-900">{proyek.timeline}</p>
                  </div>
                </div>
              )}
              {proyek.nilaiKontrak && (
                <div className="flex items-center gap-3">
                  <DollarSign className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-600">Nilai Kontrak</p>
                    <p className="font-medium text-gray-900">
                      {formatCurrency(proyek.nilaiKontrak)}
                    </p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Recent Pekerjaan */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Pekerjaan Terbaru</CardTitle>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => router.push(`/proyek/${proyek.id}/pekerjaan`)}
                >
                  Lihat Semua
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {proyek.pekerjaan.length === 0 ? (
                <div className="text-center py-8">
                  <FileText className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                  <p className="text-gray-500">Belum ada pekerjaan</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {proyek.pekerjaan.map((pekerjaan) => (
                    <div
                      key={pekerjaan.id}
                      className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                    >
                      <div className="flex-1">
                        <p className="font-medium text-gray-900 mb-1">
                          {pekerjaan.uraianPekerjaan}
                        </p>
                        <p className="text-sm text-gray-600">
                          {pekerjaan.volume} {pekerjaan.satuan} × {formatCurrency(pekerjaan.hargaSatuan)}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="font-semibold text-gray-900">
                          {formatCurrency(pekerjaan.totalBiaya)}
                        </p>
                        <Badge variant="outline" className="mt-1">
                          {pekerjaan.kategori}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Owner */}
          <Card>
            <CardHeader>
              <CardTitle>Pemilik Proyek</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-amber-100 rounded-full flex items-center justify-center">
                  <span className="text-sm font-semibold text-amber-600">
                    {proyek.user.namaLengkap.charAt(0)}
                  </span>
                </div>
                <div>
                  <p className="font-medium text-gray-900">
                    {proyek.user.namaLengkap}
                  </p>
                  <p className="text-sm text-gray-600">{proyek.user.email}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Tim Proyek */}
          <Card>
            <CardHeader>
              <CardTitle>Tim Proyek</CardTitle>
            </CardHeader>
            <CardContent>
              {proyek.timProyek.length === 0 ? (
                <p className="text-sm text-gray-500 text-center py-4">
                  Belum ada anggota tim
                </p>
              ) : (
                <div className="space-y-3">
                  {proyek.timProyek.map((tim) => (
                    <div key={tim.id} className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
                        <span className="text-xs font-semibold text-gray-600">
                          {tim.user.namaLengkap.charAt(0)}
                        </span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {tim.user.namaLengkap}
                        </p>
                        <p className="text-xs text-gray-500">{tim.role}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
