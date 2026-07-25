import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import type { Pekerjaan, KategoriPekerjaan } from '@/types';

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
      const url = `/api/pekerjaan?proyekId=${proyekId}&tipe=${kategori.toUpperCase()}`;
      const response = await fetch(url);
      
      if (!response.ok) throw new Error('Failed to fetch');
      
      const result = await response.json();
      setData(result.pekerjaan || []);
    } catch (error) {
      console.error('Fetch pekerjaan error:', error);
      toast.error('Gagal memuat data pekerjaan');
    } finally {
      setLoading(false);
    }
  };

  const createPekerjaan = async (formData: Partial<Pekerjaan>) => {
    try {
      const response = await fetch('/api/pekerjaan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          proyekId,
          tipe: kategori.toUpperCase(),
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Create failed');
      }

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
      const response = await fetch(`/api/pekerjaan/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Update failed');
      }

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
      const response = await fetch(`/api/pekerjaan/${id}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Delete failed');
      }

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
