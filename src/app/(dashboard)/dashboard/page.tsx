import {
  Building2,
  Calculator,
  ClipboardCheck,
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
import { getDashboardData, getProyekList, getNewsList } from "@/mock";

export default function DashboardPage() {
  const dashboard = getDashboardData();
  const projects = getProyekList();
  const news = getNewsList();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Selamat datang kembali, Budi!"
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
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Proyek"
          value={dashboard.stats.totalProyek}
          icon={<Building2 className="w-6 h-6" />}
          description="proyek terdaftar"
        />
        <StatCard
          title="Proyek Aktif"
          value={dashboard.stats.proyekAktif}
          icon={<ClipboardCheck className="w-6 h-6" />}
          description="sedang dikerjakan"
        />
        <StatCard
          title="Total RAB"
          value={formatCurrency(dashboard.stats.totalRAB)}
          icon={<Calculator className="w-6 h-6" />}
          description="nilai keseluruhan"
        />
        <StatCard
          title="Total Pekerjaan"
          value={dashboard.stats.totalPekerjaan}
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
              {projects.slice(0, 5).map((project) => (
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
                        {project.namaProyek}
                      </p>
                      <p className="text-xs text-gray-500">
                        {project.lokasi || "Lokasi belum diatur"}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-semibold text-gray-900">
                      {project.nilaiKontrak
                        ? formatCurrency(project.nilaiKontrak)
                        : "-"}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatDate(project.updatedAt)}
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* News */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Pengumuman</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {news.map((item) => (
                <div
                  key={item.id}
                  className="p-3 rounded-lg border bg-amber-50/50"
                >
                  <h4 className="font-semibold text-sm text-gray-900 mb-1">
                    {item.title}
                  </h4>
                  <p className="text-xs text-gray-600 line-clamp-3">
                    {item.content}
                  </p>
                  <p className="text-xs text-gray-400 mt-2">
                    {formatDate(item.createdAt)}
                  </p>
                </div>
              ))}
            </div>
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
