import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import type { WorkItem, WorkCategory } from '@/types';
import { apiClient } from '@/lib/api';

interface UsePekerjaanOptions {
  category: WorkCategory;
  projectId?: number;
}

export function useWorkItem({ category, projectId }: UsePekerjaanOptions) {
  const [data, setData] = useState<WorkItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (projectId) {
      fetchData();
    } else {
      setLoading(false);
    }
  }, [projectId, category]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const result = await apiClient.get<WorkItem[]>(`/work-items?projectId=${projectId}&category=${category}`);
      setData(Array.isArray(result) ? result : []);
    } catch (error) {
      console.error('Fetch workItem error:', error);
      toast.error('Gagal memuat data workItem');
    } finally {
      setLoading(false);
    }
  };

  const createPekerjaan = async (formData: Partial<WorkItem>) => {
    try {
      await apiClient.post('/work-items', { ...formData, projectId, category });
      await fetchData();
      toast.success('WorkItem berhasil ditambahkan');
      return true;
    } catch (error) {
      console.error('Create workItem error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal menambahkan workItem');
      return false;
    }
  };

  const updatePekerjaan = async (id: number, formData: Partial<WorkItem>) => {
    try {
      await apiClient.put(`/work-items/${id}`, formData);
      await fetchData();
      toast.success('WorkItem berhasil diupdate');
      return true;
    } catch (error) {
      console.error('Update workItem error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal update workItem');
      return false;
    }
  };

  const deletePekerjaan = async (id: number) => {
    try {
      await apiClient.delete(`/work-items/${id}`);
      await fetchData();
      toast.success('WorkItem berhasil dihapus');
      return true;
    } catch (error) {
      console.error('Delete workItem error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal menghapus workItem');
      return false;
    }
  };

  return {
    data,
    loading,
    fetchData,
    createPekerjaan,
    updatePekerjaan,
    deletePekerjaan,
  };
}
