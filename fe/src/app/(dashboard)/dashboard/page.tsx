"use client";

import { useState, useEffect } from "react";
import {
  Building2,
  Calculator,
  ClipboardCheck,
  Handshake,
  Plus,
  TrendingUp,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/shared/page-header";
import { StatCard } from "@/components/shared/stat-card";
import { formatCurrency, formatDate } from "@/lib/utils";
import Link from "next/link";
import toast from "react-hot-toast";

export default function DashboardPage() {
  const [data, setData] = useState<any>(null);
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
    fetchUser();
  }, []);

  const fetchUser = async () => {
    try {
      const response = await fetch('/api/auth/me');
      if (response.ok) {
        const result = await response.json();
        setUser(result.user);
      }
    } catch (error) {
      console.error('Fetch user error:', error);
    }
  };

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/dashboard/stats');
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

  const { stats, recentProjects } = data;

  // Guard against null/undefined stats
  if (!stats) {
    return <div className="p-8 text-center">Data tidak tersedia</div>;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description={`Selamat datang kembali, ${user?.namaLengkap || 'User'}!`}
        actions={
          <Link href="/proyek">
            <Button className="bg-amber-500 hover:bg-amber-600">
              <Plus className="w-4 h-4 mr-2" />
              Proyek Baru
            </Button>
          </Link>
        }
      />

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
        <StatCard
          title="Total Proyek"
          value={stats.totalProyek || 0}
          icon={<Building2 className="w-6 h-6" />}
          description="proyek terdaftar"
        />
        <StatCard
          title="Proyek Aktif"
          value={stats.proyekAktif || 0}
          icon={<ClipboardCheck className="w-6 h-6" />}
          description="sedang dikerjakan"
        />
        <StatCard
          title="Proyek Pitching"
          value={stats.proyekPitching || 0}
          icon={<Handshake className="w-6 h-6" />}
          description="dalam penawaran"
        />
        <StatCard
          title="Total RAB"
          value={formatCurrency(stats.totalRAB || 0)}
          icon={<Calculator className="w-6 h-6" />}
          description="nilai keseluruhan"
        />
        <StatCard
          title="Total Pekerjaan"
          value={stats.totalPekerjaan || 0}
          icon={<TrendingUp className="w-6 h-6" />}
          description="item pekerjaan"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Projects */}
        <Card className="lg:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Proyek Terbaru</CardTitle>
            <Link href="/proyek">
              <Button variant="ghost" size="sm" className="text-amber-600">
                Lihat Semua
              </Button>
            </Link>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {(recentProjects || []).map((project: any) => (
                <Link
                  key={project.id}
                  href={`/proyek/${project.id}`}
                  className="flex items-center justify-between p-3 rounded-lg border hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center">
                      <Building2 className="w-5 h-5 text-amber-600" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">
                        {project.nama}
                      </p>
                      <p className="text-xs text-gray-500">
                        {project.lokasi || "Lokasi belum diatur"}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-semibold text-gray-900">
                      {project.totalRAB
                        ? formatCurrency(project.totalRAB)
                        : "-"}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatDate(project.createdAt)}
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* News - removed, no API */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Pengumuman</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-500">Tidak ada pengumuman</p>
          </CardContent>
        </Card>
      </div>

      {/* Quick Links */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Aksi Cepat</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <Link href="/proyek">
              <Button
                variant="outline"
                className="w-full h-auto py-4 flex flex-col gap-2"
              >
                <Building2 className="w-5 h-5 text-amber-600" />
                <span className="text-xs">Buat Proyek</span>
              </Button>
            </Link>
            <Link href="/master-harga">
              <Button
                variant="outline"
                className="w-full h-auto py-4 flex flex-col gap-2"
              >
                <Calculator className="w-5 h-5 text-amber-600" />
                <span className="text-xs">Master Harga</span>
              </Button>
            </Link>
            <Link href="/rekap">
              <Button
                variant="outline"
                className="w-full h-auto py-4 flex flex-col gap-2"
              >
                <TrendingUp className="w-5 h-5 text-amber-600" />
                <span className="text-xs">Rekapitulasi</span>
              </Button>
            </Link>
            <Link href="/monitoring">
              <Button
                variant="outline"
                className="w-full h-auto py-4 flex flex-col gap-2"
              >
                <ClipboardCheck className="w-5 h-5 text-amber-600" />
                <span className="text-xs">Progress</span>
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
