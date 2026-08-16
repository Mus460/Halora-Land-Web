"use client";

import { useState, useEffect } from "react";
import { Plus, LayoutGrid, List } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/shared/page-header";
import { SearchInput } from "@/components/shared/search-input";
import { EmptyState } from "@/components/shared/empty-state";
import { ProjectCard } from "@/components/projects/project-card";
import { ProjectForm } from "@/components/projects/project-form";
import { useProjectStore } from "@/stores/use-project-store";
import { useProject } from "@/contexts/ProjectContext";
import { useDebouncedValue } from "@/hooks/use-debounce";
import toast from "react-hot-toast";
import type { Project } from "@/types";

export default function ProjectPage() {
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search);
  const [showForm, setShowForm] = useState(false);
  const [editProject, setEditProject] = useState<Project | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { activeProject, setActiveProject } = useProjectStore();
  const { setCurrentProjectId, refreshProjectList } = useProject();

  const fetchProjects = async () => {
    try {
      setIsLoading(true);
      const response = await fetch("/api/projects");
      const data = await response.json();
      
      if (!response.ok) {
        throw new Error(data.error || 'Failed to fetch projects');
      }
      
      setProjects(data.projects || []);
    } catch (error) {
      toast.error('Gagal memuat data project');
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const filtered = projects.filter((p) =>
    p.name.toLowerCase().includes(debouncedSearch.toLowerCase())
  );

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/projects/${id}`, {
        method: 'DELETE',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to delete project');
      }

      setProjects((prev) => prev.filter((p) => p.id !== id));
      refreshProjectList();

      if (activeProject?.id === id) {
        setActiveProject(null);
      }
      
      toast.success("Proyek berhasil dihapus");
    } catch (error: any) {
      toast.error(error.message || 'Gagal menghapus project');
    }
  };

  const handleSetActive = async (id: number) => {
    const project = projects.find((p) => p.id === id);
    if (!project) return;

    // A project cannot be active, pitching and done at the same time
    if (project.isPitching || project.isDone) {
      try {
        const response = await fetch(`/api/projects/${id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ isPitching: false, isDone: false }),
        });
        const result = await response.json();
        if (response.ok) {
          setProjects((prev) =>
            prev.map((p) => (p.id === id ? result.projects : p))
          );
        }
      } catch (error) {
        console.error('Gagal mengubah status project:', error);
      }
    }

    setActiveProject(project);
    toast.success(`"${project.name}" dijadikan proyek aktif`);
  };

  const handleSetPitching = async (id: number, isPitching: boolean) => {
    try {
      const response = await fetch(`/api/projects/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isPitching, isDone: false }),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to update project status');
      }

      setProjects((prev) =>
        prev.map((p) => (p.id === id ? result.projects : p))
      );

      // A project cannot be active and pitching at the same time
      if (isPitching && activeProject?.id === id) {
        setActiveProject(null);
      }

      toast.success(
        isPitching
          ? `"${result.projects.name}" dijadikan proyek pitching`
          : `"${result.projects.name}" dijadikan proyek aktif`
      );
    } catch (error: any) {
      toast.error(error.message || 'Gagal mengubah status project');
    }
  };

  const handleSetDone = async (id: number, isDone: boolean) => {
    try {
      const response = await fetch(`/api/projects/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isDone, isPitching: false }),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to update project status');
      }

      setProjects((prev) =>
        prev.map((p) => (p.id === id ? result.projects : p))
      );

      // A done project cannot be the active project
      if (isDone && activeProject?.id === id) {
        setActiveProject(null);
      }

      toast.success(
        isDone
          ? `"${result.projects.name}" ditandai selesai`
          : `"${result.projects.name}" dibuka kembali`
      );
    } catch (error: any) {
      toast.error(error.message || 'Gagal mengubah status project');
    }
  };

  const handleSubmit = async (data: Partial<Project>) => {
    try {
      if (editProject) {
        // Update existing
        const response = await fetch(`/api/projects/${editProject.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (!response.ok) {
          throw new Error(result.error || 'Failed to update project');
        }

        setProjects((prev) =>
          prev.map((p) => (p.id === editProject.id ? result.projects : p))
        );
        refreshProjectList();
        toast.success("Proyek berhasil diupdate");
      } else {
        // Create new
        const response = await fetch("/api/projects", {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (!response.ok) {
          throw new Error(result.error || 'Failed to create project');
        }

        setProjects((prev) => [result.projects, ...prev]);
        setActiveProject(result.projects);
        setCurrentProjectId(result.projects.id);
        refreshProjectList();
        toast.success("Proyek berhasil dibuat");
      }
      
      setEditProject(null);
      setShowForm(false);
    } catch (error: any) {
      toast.error(error.message || 'Terjadi kesalahan');
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader
          title="Data Project"
          description="Kelola semua proyek konstruksi Anda"
        />
        <div className="text-center py-12">
          <p className="text-gray-500">Memuat data project...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Data Project"
        description="Kelola semua proyek konstruksi Anda"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => {
              setEditProject(null);
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
          {filtered.map((project) => (
            <ProjectCard
              key={project.id}
              project={project}
              onDelete={handleDelete}
              onSetActive={handleSetActive}
              onSetPitching={handleSetPitching}
              onSetDone={handleSetDone}
              isActive={activeProject?.id === project.id}
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
                setEditProject(null);
                setShowForm(true);
              }}
            >
              <Plus className="w-4 h-4 mr-2" />
              Buat Project
            </Button>
          }
        />
      )}

      <ProjectForm
        open={showForm}
        onOpenChange={setShowForm}
        project={editProject}
        onSubmit={handleSubmit}
      />
    </div>
  );
}
