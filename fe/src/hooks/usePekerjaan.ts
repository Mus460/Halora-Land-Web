import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import type { Pekerjaan, KategoriPekerjaan } from '@/types';
import { apiClient } from '@/lib/api';

interface UsePekerjaanOptions {
  kategori: KategoriPekerjaan;
  proyekId?: number;
}

export function usePekerjaan({ kategori, proyekId }: UsePekerjaanOptions) {
  const [data, setData] = useState<Pekerjaan[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (proyekId) {
      fetchData();
    } else {
      setLoading(false);
    }
  }, [proyekId, kategori]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const result = await apiClient.get<Pekerjaan[]>(`/pekerjaan?proyekId=${proyekId}&kategori=${kategori}`);
      setData(Array.isArray(result) ? result : []);
    } catch (error) {
      console.error('Fetch pekerjaan error:', error);
      toast.error('Gagal memuat data pekerjaan');
    } finally {
      setLoading(false);
    }
  };

  const createPekerjaan = async (formData: Partial<Pekerjaan>) => {
    try {
      await apiClient.post('/pekerjaan', { ...formData, proyekId, kategori });
      await fetchData();
      toast.success('Pekerjaan berhasil ditambahkan');
      return true;
    } catch (error) {
      console.error('Create pekerjaan error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal menambahkan pekerjaan');
      return false;
    }
  };

  const updatePekerjaan = async (id: number, formData: Partial<Pekerjaan>) => {
    try {
      await apiClient.put(`/pekerjaan/${id}`, formData);
      await fetchData();
      toast.success('Pekerjaan berhasil diupdate');
      return true;
    } catch (error) {
      console.error('Update pekerjaan error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal update pekerjaan');
      return false;
    }
  };

  const deletePekerjaan = async (id: number) => {
    try {
      await apiClient.delete(`/pekerjaan/${id}`);
      await fetchData();
      toast.success('Pekerjaan berhasil dihapus');
      return true;
    } catch (error) {
      console.error('Delete pekerjaan error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal menghapus pekerjaan');
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
