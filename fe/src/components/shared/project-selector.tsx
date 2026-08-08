'use client';

import { useProject } from '@/contexts/ProjectContext';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Building2 } from 'lucide-react';

export function ProjectSelector() {
  const { currentProjectId, setCurrentProjectId, projectList, loading } = useProject();
  const selectedProyek = projectList.find((p) => p.id === currentProjectId);

  if (loading && projectList.length === 0) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-sm text-gray-500">
        <Building2 className="w-4 h-4" />
        <span>Memuat proyek...</span>
      </div>
    );
  }

  if (projectList.length === 0) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-sm text-gray-500">
        <Building2 className="w-4 h-4" />
        <span>Belum ada proyek</span>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Building2 className="w-4 h-4 text-gray-500 hidden sm:block" />
      <Select
        value={currentProjectId?.toString() || ''}
        onValueChange={(value) => value && setCurrentProjectId(parseInt(value))}
      >
        <SelectTrigger className="w-[200px] sm:w-[250px]">
          <SelectValue placeholder="Pilih proyek">
            {selectedProyek?.name}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {projectList.map((project) => (
            <SelectItem key={project.id} value={project.id.toString()}>
              <div className="flex flex-col">
                <span className="font-medium">{project.name}</span>
                {project.location && (
                  <span className="text-xs text-gray-500">{project.location}</span>
                )}
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
