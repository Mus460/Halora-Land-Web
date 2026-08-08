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
  Trash2,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { formatCurrency, formatTimeline } from "@/lib/utils";
import toast from "react-hot-toast";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";

interface ProjectDetail {
  id: number;
  name: string;
  location: string | null;
  type: string;
  contractValue: number | null;
  timelineMonths: number;
  timelineDays: number;
  createdAt: string;
  user: {
    id: number;
    fullName: string;
    email: string;
  };
  projectTeam: Array<{
    id: number;
    role: string;
    user: {
      id: number;
      fullName: string;
      email: string;
    };
  }>;
  work_items: Array<{
    id: number;
    description: string;
    volume: number;
    unit: string;
    unitPrice: number;
    totalCost: number;
    category: string;
  }>;
  _count: {
    work_items: number;
    recaps: number;
    invoices: number;
  };
}

export default function ProyekDetailPage() {
  const params = useParams();
  const router = useRouter();
  const [project, setProject] = useState<ProjectDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    fetchProjectDetail();
  }, [params.id]);

  const fetchProjectDetail = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/projects/${params.id}`);
      
      if (!response.ok) {
        if (response.status === 404) {
          toast.error("Proyek tidak ditemukan");
          router.push("/projects");
          return;
        }
        if (response.status === 403) {
          toast.error("Anda tidak memiliki akses ke proyek ini");
          router.push("/projects");
          return;
        }
        throw new Error("Gagal memuat detail proyek");
      }

      const data = await response.json();
      setProject(data.projects);
    } catch (error) {
      console.error("Error fetching proyek:", error);
      toast.error("Gagal memuat detail proyek");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    setConfirmDelete(false);

    try {
      const response = await fetch(`/api/projects/${params.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error("Gagal menghapus proyek");
      }

      toast.success("Proyek berhasil dihapus");
      router.push("/projects");
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
          <p className="mt-4 text-gray-600">Memuat detail project...</p>
        </div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Building2 className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-600">Project tidak ditemukan</p>
          <Button
            onClick={() => router.push("/projects")}
            className="mt-4"
            variant="outline"
          >
            Kembali ke Daftar Project
          </Button>
        </div>
      </div>
    );
  }

  const totalCost = (project.work_items || []).reduce(
    (sum, p) => sum + (Number(p.totalCost) || 0),
    0
  );

  return (
    <div className="container mx-auto p-6 max-w-7xl">
      {/* Header */}
      <div className="mb-6">
        <Button
          variant="ghost"
          onClick={() => router.push("/projects")}
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
                {project.name}
              </h1>
              <div className="flex items-center gap-3 text-sm text-gray-600">
                <Badge variant="outline">
                  {project.type === "building" ? "Gedung" : "Infrastruktur"}
                </Badge>
                {project.location && (
                  <>
                    <span>•</span>
                    <div className="flex items-center gap-1">
                      <MapPin className="w-4 h-4" />
                      {project.location}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>

          <div className="flex gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={() => setConfirmDelete(true)}
              className="text-red-600 hover:text-red-700"
            >
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Hapus Proyek"
        description={`Yakin ingin menghapus proyek "${project.name}"? Semua data terkait akan dihapus.`}
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={handleDelete}
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">Total WorkItem</p>
                <p className="text-2xl font-bold text-gray-900">
                  {project._count.work_items}
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
                <p className="text-sm text-gray-600 mb-1">Total Cost</p>
                <p className="text-2xl font-bold text-gray-900">
                  {formatCurrency(totalCost)}
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
                  {project._count.invoices}
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
                <p className="text-sm text-gray-600 mb-1">Team</p>
                <p className="text-2xl font-bold text-gray-900">
                  {(project.projectTeam?.length || 0) + 1}                </p>
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
              {(project.timelineMonths > 0 || project.timelineDays > 0) && (
                <div className="flex items-center gap-3">
                  <Calendar className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-600">Timeline</p>
                    <p className="font-medium text-gray-900">
                      {formatTimeline(project.timelineMonths, project.timelineDays)}
                    </p>
                  </div>
                </div>
              )}
              {project.contractValue && (
                <div className="flex items-center gap-3">
                  <DollarSign className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="text-sm text-gray-600">Nilai Kontrak</p>
                    <p className="font-medium text-gray-900">
                      {formatCurrency(project.contractValue)}
                    </p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Recent WorkItem */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Pekerjaan Terbaru</CardTitle>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => router.push(`/project/${project.id}/workItem`)}
                >
                  Lihat Semua
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {(project.work_items || []).length === 0 ? (
                <div className="text-center py-8">
                  <FileText className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                  <p className="text-gray-500">Belum ada pekerjaan</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {(project.work_items || []).map((workItem) => (
                    <div
                      key={workItem.id}
                      className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                    >
                      <div className="flex-1">
                        <p className="font-medium text-gray-900 mb-1">
                          {workItem.description}
                        </p>
                        <p className="text-sm text-gray-600">
                          {workItem.volume} {workItem.unit} × {formatCurrency(workItem.unitPrice)}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="font-semibold text-gray-900">
                          {formatCurrency(workItem.totalCost)}
                        </p>
                        <Badge variant="outline" className="mt-1">
                          {workItem.category}
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
                    {project.user.fullName.charAt(0)}
                  </span>
                </div>
                <div>
                  <p className="font-medium text-gray-900">
                    {project.user.fullName}
                  </p>
                  <p className="text-sm text-gray-600">{project.user.email}</p>
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
              {(project.projectTeam || []).length === 0 ? (
                <p className="text-sm text-gray-500 text-center py-4">
                  Belum ada anggota tim
                </p>
              ) : (
                <div className="space-y-3">
                  {(project.projectTeam || []).map((tim) => (
                    <div key={tim.id} className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
                        <span className="text-xs font-semibold text-gray-600">
                          {tim.user.fullName.charAt(0)}
                        </span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {tim.user.fullName}
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
