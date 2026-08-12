'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiClient } from '@/lib/api';

interface Project {
  id: number;
  name: string;
  location: string | null;
  userId: string;
}

interface ProjectContextType {
  currentProjectId: number | null;
  setCurrentProjectId: (id: number) => void;
  project: Project | null;
  projectList: Project[];
  loading: boolean;
  refreshProjectList: () => void;
}

const ProjectContext = createContext<ProjectContextType>({
  currentProjectId: null,
  setCurrentProjectId: () => {},
  project: null,
  projectList: [],
  loading: false,
  refreshProjectList: () => {},
});

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [currentProjectId, setCurrentProjectIdState] = useState<number | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [projectList, setProjectList] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);

  // Load from localStorage on mount
  useEffect(() => {
    const stored = localStorage.getItem('currentProjectId');
    if (stored) {
      const id = parseInt(stored);
      if (!isNaN(id)) {
        setCurrentProjectIdState(id);
      }
    }
  }, []);

  // Fetch project list on mount
  useEffect(() => {
    fetchProjectList();
  }, []);

  // Fetch current project details when ID changes
  useEffect(() => {
    if (currentProjectId) {
      fetchProject(currentProjectId);
    } else {
      setProject(null);
    }
  }, [currentProjectId]);

  const fetchProjectList = async () => {
    try {
      setLoading(true);
      const data = await apiClient.get<{ projects: Project[] }>("/projects");
      const list = Array.isArray(data?.projects) ? data?.projects : [];
      setProjectList(list);

      const stored = localStorage.getItem('currentProjectId');
      const storedId = stored ? parseInt(stored) : NaN;
      const storedIsValid = list.some((p) => p.id === storedId);
      if (list.length > 0 && !storedIsValid) {
        const firstId = list[0].id;
        setCurrentProjectIdState(firstId);
        localStorage.setItem('currentProjectId', String(firstId));
      }
    } catch (error) {
      console.error('Fetch project list error:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchProject = async (id: number) => {
    try {
      const data = await apiClient.get<{ projects: Project }>(`/projects/${id}`);
      setProject(data?.projects || null);
    } catch (error) {
      console.error('Fetch project error:', error);
    }
  };

  const setCurrentProjectId = (id: number) => {
    setCurrentProjectIdState(id);
    localStorage.setItem('currentProjectId', String(id));
  };

  const refreshProjectList = () => {
    fetchProjectList();
  };

  return (
    <ProjectContext.Provider
      value={{
        currentProjectId,
        setCurrentProjectId,
        project,
        projectList,
        loading,
        refreshProjectList,
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
