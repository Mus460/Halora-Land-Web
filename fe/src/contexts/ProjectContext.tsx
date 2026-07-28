'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiClient } from '@/lib/api';

interface Proyek {
  id: number;
  namaProyek: string;
  lokasi: string | null;
  userId: string;
}

interface ProjectContextType {
  currentProyekId: number | null;
  setCurrentProyekId: (id: number) => void;
  proyek: Proyek | null;
  proyekList: Proyek[];
  loading: boolean;
  refreshProyekList: () => void;
}

const ProjectContext = createContext<ProjectContextType>({
  currentProyekId: null,
  setCurrentProyekId: () => {},
  proyek: null,
  proyekList: [],
  loading: false,
  refreshProyekList: () => {},
});

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [currentProyekId, setCurrentProyekIdState] = useState<number | null>(null);
  const [proyek, setProyek] = useState<Proyek | null>(null);
  const [proyekList, setProyekList] = useState<Proyek[]>([]);
  const [loading, setLoading] = useState(true);

  // Load from localStorage on mount
  useEffect(() => {
    const stored = localStorage.getItem('currentProyekId');
    if (stored) {
      const id = parseInt(stored);
      if (!isNaN(id)) {
        setCurrentProyekIdState(id);
      }
    }
  }, []);

  // Fetch project list on mount
  useEffect(() => {
    fetchProyekList();
  }, []);

  // Fetch current project details when ID changes
  useEffect(() => {
    if (currentProyekId) {
      fetchProyek(currentProyekId);
    } else {
      setProyek(null);
    }
  }, [currentProyekId]);

  const fetchProyekList = async () => {
    try {
      setLoading(true);
      const data = await apiClient.get<Proyek[]>('/proyek');
      const list = Array.isArray(data) ? data : [];
      setProyekList(list);

      if (!currentProyekId && list.length > 0) {
        const firstId = list[0].id;
        setCurrentProyekIdState(firstId);
        localStorage.setItem('currentProyekId', String(firstId));
      }
    } catch (error) {
      console.error('Fetch proyek list error:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchProyek = async (id: number) => {
    try {
      const data = await apiClient.get<Proyek>(`/proyek/${id}`);
      setProyek(data);
    } catch (error) {
      console.error('Fetch proyek error:', error);
    }
  };

  const setCurrentProyekId = (id: number) => {
    setCurrentProyekIdState(id);
    localStorage.setItem('currentProyekId', String(id));
  };

  const refreshProyekList = () => {
    fetchProyekList();
  };

  return (
    <ProjectContext.Provider
      value={{
        currentProyekId,
        setCurrentProyekId,
        proyek,
        proyekList,
        loading,
        refreshProyekList,
      }}
    >
      {children}
    </ProjectContext.Provider>
  );
}

export const useProject = () => {
  const context = useContext(ProjectContext);
  if (!context) {
    throw new Error('useProject must be used within ProjectProvider');
  }
  return context;
};
