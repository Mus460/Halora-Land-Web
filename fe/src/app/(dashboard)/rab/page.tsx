"use client"

import { useState, useEffect } from 'react'
import { useProject } from '@/contexts/ProjectContext'
import { Search, Plus, Calculator } from 'lucide-react'
import { PageHeader } from '@/components/shared/page-header'
import { useDebouncedValue } from '@/hooks/use-debounce'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatCurrency, formatWaktu } from '@/lib/utils'
import toast from 'react-hot-toast'

interface SearchResult {
  id: number
  kode: string
  nama: string
  satuan: string
  hargaSatuan: number
  kategori: string
  ahspKode: string
  ahspSheet: string
}

interface PekerjaanItem {
  id: number
  kategori: string
  uraianPekerjaan: string
  volume: number
  satuan: string
  hargaSatuan: number
  totalBiaya: number
  totalWaktu: number | null
}

export default function RABPage() {
  const { currentProyekId: proyekId } = useProject()
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [pekerjaan, setPekerjaan] = useState<PekerjaanItem[]>([])
  const [loading, setLoading] = useState(true)
  const [addingItem, setAddingItem] = useState<number | null>(null)

  useEffect(() => {
    if (proyekId) {
      fetchPekerjaan()
    }
  }, [proyekId])

  const debouncedQuery = useDebouncedValue(searchQuery)

  useEffect(() => {
    if (debouncedQuery.trim().length >= 3) {
      performSearch()
    } else {
      setSearchResults([])
    }
  }, [debouncedQuery])

  const fetchPekerjaan = async () => {
    if (!proyekId) return
    
    try {
      setLoading(true)
      const response = await fetch(`/api/proyek/${proyekId}/rekap`)
      if (!response.ok) throw new Error('Failed to fetch')
      
      const data = await response.json()
      // Extract pekerjaan from grouped data
      const allPekerjaan: PekerjaanItem[] = []
      if (data.grouped) {
        Object.values(data.grouped || {}).forEach((items: any) => {
          allPekerjaan.push(...(items || []))
        })
      }
      setPekerjaan(allPekerjaan)
    } catch (error) {
      console.error('Fetch pekerjaan error:', error)
      toast.error('Gagal memuat data RAB')
    } finally {
      setLoading(false)
    }
  }

  const performSearch = async () => {
    setSearching(true)
    try {
      const response = await fetch(
        `/api/master-analisa/search?q=${encodeURIComponent(searchQuery)}&limit=10`
      )
      if (!response.ok) throw new Error('Search failed')
      
      const data = await response.json()
      setSearchResults(data.results || [])
    } catch (error) {
      console.error('Search error:', error)
      toast.error('Gagal mencari')
    } finally {
      setSearching(false)
    }
  }

  const handleAddItem = async (item: SearchResult) => {
    if (!proyekId) {
      toast.error('Pilih proyek terlebih dahulu')
      return
    }

    const volume = prompt(`Masukkan volume untuk:\n${item.nama}\n\nVolume (${item.satuan}):`)
    if (!volume || parseFloat(volume) <= 0) {
      toast.error('Volume harus diisi dan lebih dari 0')
      return
    }

    try {
      setAddingItem(item.id)
      
      const response = await fetch(`/api/proyek/${proyekId}/pekerjaan/from-ahsp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          masterAnalisaId: item.id,
          volume: parseFloat(volume),
          applyBreakdown: true
        })
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to add')
      }

      toast.success('Item berhasil ditambahkan ke RAB')
      
      // Refresh pekerjaan list
      await fetchPekerjaan()
      
      // Clear search
      setSearchQuery('')
      setSearchResults([])
      
    } catch (error: any) {
      console.error('Add item error:', error)
      toast.error(error.message || 'Gagal menambahkan item')
    } finally {
      setAddingItem(null)
    }
  }

  // Calculate totals
  const subtotal = pekerjaan.reduce((sum, p) => sum + Number(p.totalBiaya), 0)
  const totalWaktu = pekerjaan.reduce((sum, p) => sum + (p.totalWaktu || 0), 0)
  const overhead = subtotal * 0.10
  const profit = (subtotal + overhead) * 0.10
  const ppn = (subtotal + overhead + profit) * 0.11
  const total = subtotal + overhead + profit + ppn

  // Group by kategori
  const grouped = pekerjaan.reduce((acc, item) => {
    if (!acc[item.kategori]) acc[item.kategori] = []
    acc[item.kategori].push(item)
    return acc
  }, {} as Record<string, PekerjaanItem[]>)

  if (!proyekId) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <Calculator className="w-12 h-12 text-gray-300 mb-3" />
        <p className="text-gray-600">Silakan pilih proyek terlebih dahulu</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Rencana Anggaran Biaya (RAB)"
        description="Kelola RAB proyek dengan data AHSP PUPR 2026"
      />

      {/* Search Section */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Cari Item Pekerjaan AHSP</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <Input
              placeholder="Cari pekerjaan... (misal: beton k-300, penulangan, atap)"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          {/* Search Results */}
          {searchResults.length > 0 && (
            <div className="mt-4 space-y-2 max-h-96 overflow-y-auto">
              {searchResults.map((item) => (
                <div
                  key={item.id}
                  className="flex items-start justify-between p-3 border rounded-lg hover:bg-gray-50"
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant="outline" className="text-xs">
                        {item.kategori}
                      </Badge>
                      <span className="text-xs text-gray-500">{item.ahspKode}</span>
                    </div>
                    <div className="font-medium text-sm truncate">{item.nama}</div>
                    <div className="text-sm text-gray-600 mt-1">
                      {formatCurrency(item.hargaSatuan)} / {item.satuan}
                    </div>
                  </div>
                  <Button
                    size="sm"
                    onClick={() => handleAddItem(item)}
                    disabled={addingItem === item.id}
                    className="ml-3 bg-amber-500 hover:bg-amber-600"
                  >
                    {addingItem === item.id ? 'Menambahkan...' : 'Tambah'}
                  </Button>
                </div>
              ))}
            </div>
          )}

          {searching && (
            <div className="mt-4 text-center text-gray-500">
              Mencari...
            </div>
          )}

          {searchQuery.length >= 3 && !searching && searchResults.length === 0 && (
            <div className="mt-4 text-center text-gray-500">
              Tidak ada hasil ditemukan. Coba kata kunci lain.
            </div>
          )}
        </CardContent>
      </Card>

      {/* RAB Items */}
      <Card>
        <CardHeader>
          <CardTitle>Item RAB Saat Ini</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8 text-gray-500">Memuat data...</div>
          ) : pekerjaan.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              Belum ada item pekerjaan. Cari dan tambahkan dari AHSP di atas.
            </div>
          ) : (
            <div className="space-y-6">
              {Object.entries(grouped).map(([kategori, items]) => {
                const subtotalKategori = items.reduce((sum, p) => sum + Number(p.totalBiaya), 0)
                const waktuKategori = items.reduce((sum, p) => sum + (p.totalWaktu || 0), 0)
                
                return (
                  <div key={kategori}>
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="font-semibold text-lg capitalize">
                        {kategori.replace('_', ' ')}
                      </h3>
                      <div className="text-sm text-gray-600">
                        Subtotal: {formatCurrency(subtotalKategori)}
                        <span className="ml-3 text-gray-400">
                          · {formatWaktu(waktuKategori)}
                        </span>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      {items.map((item, idx) => (
                        <div key={item.id} className="flex items-center gap-3 p-3 border rounded-lg">
                          <div className="flex-1 min-w-0">
                            <div className="font-medium text-sm truncate">{item.uraianPekerjaan}</div>
                            <div className="text-xs text-gray-500 mt-1">
                              {item.volume} {item.satuan} × {formatCurrency(item.hargaSatuan)}
                            </div>
                            <div className="text-xs text-gray-400 mt-0.5">
                              Estimasi waktu: {formatWaktu(item.totalWaktu)}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="font-semibold">{formatCurrency(item.totalBiaya)}</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })}

              {/* Totals */}
              <div className="border-t pt-4 space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Subtotal Pekerjaan:</span>
                  <span className="font-medium">{formatCurrency(subtotal)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>Estimasi Waktu:</span>
                  <span className="font-medium">{formatWaktu(totalWaktu)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>Overhead (10%):</span>
                  <span className="font-medium">{formatCurrency(overhead)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>Profit (10%):</span>
                  <span className="font-medium">{formatCurrency(profit)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>PPN (11%):</span>
                  <span className="font-medium">{formatCurrency(ppn)}</span>
                </div>
                <div className="flex justify-between text-lg font-bold pt-2 border-t">
                  <span>TOTAL RAB:</span>
                  <span className="text-amber-600">{formatCurrency(total)}</span>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
