"use client";

import { useState, useEffect } from "react";
import { Plus, LayoutGrid, List } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { SearchInput } from "@/components/shared/search-input";
import { EmptyState } from "@/components/shared/empty-state";
import { ProyekCard } from "@/components/proyek/proyek-card";
import { ProyekForm } from "@/components/proyek/proyek-form";
import { useProjectStore } from "@/stores/use-project-store";
import toast from "react-hot-toast";
import type { Proyek } from "@/types";

export default function ProyekPage() {
  const [search, setSearch] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [editProyek, setEditProyek] = useState<Proyek | null>(null);
  const [projects, setProjects] = useState<Proyek[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { activeProject, setActiveProject } = useProjectStore();

  const fetchProjects = async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/api/proyek');
      const data = await response.json();
      
      if (!response.ok) {
        throw new Error(data.error || 'Failed to fetch projects');
      }
      
      setProjects(data.proyek);
    } catch (error) {
      toast.error('Gagal memuat data proyek');
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const filtered = projects.filter((p) =>
    p.namaProyek.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/proyek/${id}`, {
        method: 'DELETE',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to delete project');
      }

      setProjects((prev) => prev.filter((p) => p.id !== id));
      
      if (activeProject?.id === id) {
        setActiveProject(null);
      }
      
      toast.success("Proyek berhasil dihapus");
    } catch (error: any) {
      toast.error(error.message || 'Gagal menghapus proyek');
    }
  };

  const handleDuplicate = (id: number) => {
    const project = projects.find((p) => p.id === id);
    if (project) {
      const duplicated: Partial<Proyek> = {
        namaProyek: `${project.namaProyek} (Copy)`,
        lokasi: project.lokasi,
        tipe: project.tipe,
        nilaiKontrak: project.nilaiKontrak,
        timeline: project.timeline,
      };
      
      setEditProyek(null);
      setShowForm(true);
      // Pass duplicated data to form
    }
  };

  const handleSetActive = (id: number) => {
    const project = projects.find((p) => p.id === id);
    if (project) {
      setActiveProject(project);
      toast.success(`"${project.namaProyek}" dijadikan proyek aktif`);
    }
  };

  const handleSubmit = async (data: Partial<Proyek>) => {
    try {
      if (editProyek) {
        // Update existing
        const response = await fetch(`/api/proyek/${editProyek.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (!response.ok) {
          throw new Error(result.error || 'Failed to update project');
        }

        setProjects((prev) =>
          prev.map((p) => (p.id === editProyek.id ? result.proyek : p))
        );
        toast.success("Proyek berhasil diupdate");
      } else {
        // Create new
        const response = await fetch('/api/proyek', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (!response.ok) {
          throw new Error(result.error || 'Failed to create project');
        }

        setProjects((prev) => [result.proyek, ...prev]);
        toast.success("Proyek berhasil dibuat");
      }
      
      setEditProyek(null);
      setShowForm(false);
    } catch (error: any) {
      toast.error(error.message || 'Terjadi kesalahan');
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Data Proyek"
          description="Kelola semua proyek konstruksi Anda"
        />
        <div className="text-center py-12">
          <p className="text-gray-500">Memuat data proyek...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Data Proyek"
        description="Kelola semua proyek konstruksi Anda"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => {
              setEditProyek(null);
              setShowForm(true);
            }}
          >
            <Plus className="w-4 h-4 mr-2" />
            Proyek Baru
          </Button>
        }
      />

      <div className="flex items-center gap-4">
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Cari proyek..."
          className="max-w-sm"
        />
      </div>

      {filtered.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((proyek) => (
            <ProyekCard
              key={proyek.id}
              proyek={proyek}
              onDelete={handleDelete}
              onDuplicate={handleDuplicate}
              onSetActive={handleSetActive}
              isActive={activeProject?.id === proyek.id}
            />
          ))}
        </div>
      ) : (
        <EmptyState
          title="Belum ada proyek"
          description="Buat proyek baru untuk mulai menghitung RAB konstruksi"
          action={
            <Button
              className="bg-amber-500 hover:bg-amber-600"
              onClick={() => {
                setEditProyek(null);
                setShowForm(true);
              }}
            >
              <Plus className="w-4 h-4 mr-2" />
              Buat Proyek
            </Button>
          }
        />
      )}

      <ProyekForm
        open={showForm}
        onOpenChange={setShowForm}
        proyek={editProyek}
        onSubmit={handleSubmit}
      />
    </div>
  );
}
