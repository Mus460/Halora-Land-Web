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
  const { currentProyekId, setCurrentProyekId, proyekList, loading } = useProject();

  if (loading && proyekList.length === 0) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-sm text-gray-500">
        <Building2 className="w-4 h-4" />
        <span>Memuat proyek...</span>
      </div>
    );
  }

  if (proyekList.length === 0) {
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
        value={currentProyekId?.toString() || ''}
        onValueChange={(value) => setCurrentProyekId(parseInt(value))}
      >
        <SelectTrigger className="w-[200px] sm:w-[250px]">
          <SelectValue placeholder="Pilih proyek" />
        </SelectTrigger>
        <SelectContent>
          {proyekList.map((proyek) => (
            <SelectItem key={proyek.id} value={proyek.id.toString()}>
              <div className="flex flex-col">
                <span className="font-medium">{proyek.namaProyek}</span>
                {proyek.lokasi && (
                  <span className="text-xs text-gray-500">{proyek.lokasi}</span>
                )}
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
