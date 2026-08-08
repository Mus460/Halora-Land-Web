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
import { formatCurrency, formatDuration } from '@/lib/utils'
import toast from 'react-hot-toast'

interface SearchResult {
  id: number
  code: string
  name: string
  unit: string
  unitPrice: number
  category: string
  ahspCode: string
  ahspSheet: string
}

interface PekerjaanItem {
  id: number
  category: string
  description: string
  volume: number
  unit: string
  unitPrice: number
  totalCost: number
  totalDuration: number | null
}

export default function RABPage() {
  const { currentProjectId: projectId } = useProject()
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<SearchResult[]>([])
  const [searching, setSearching] = useState(false)
  const [workItem, setPekerjaan] = useState<PekerjaanItem[]>([])
  const [loading, setLoading] = useState(true)
  const [addingItem, setAddingItem] = useState<number | null>(null)

  useEffect(() => {
    if (projectId) {
      fetchPekerjaan()
    }
  }, [projectId])

  const debouncedQuery = useDebouncedValue(searchQuery)

  useEffect(() => {
    if (debouncedQuery.trim().length >= 3) {
      performSearch()
    } else {
      setSearchResults([])
    }
  }, [debouncedQuery])

  const fetchPekerjaan = async () => {
    if (!projectId) return
    
    try {
      setLoading(true)
      const response = await fetch(`/api/projects/${projectId}/recaps`)
      if (!response.ok) throw new Error('Failed to fetch')
      
      const data = await response.json()
      // Extract workItem from grouped data
      const allPekerjaan: PekerjaanItem[] = []
      if (data.grouped) {
        Object.values(data.grouped || {}).forEach((items: any) => {
          allPekerjaan.push(...(items || []))
        })
      }
      setPekerjaan(allPekerjaan)
    } catch (error) {
      console.error('Fetch workItem error:', error)
      toast.error('Gagal memuat data RAB')
    } finally {
      setLoading(false)
    }
  }

  const performSearch = async () => {
    setSearching(true)
    try {
      const response = await fetch(
        `/api/analysis-masters/search?q=${encodeURIComponent(searchQuery)}&limit=10`
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
    if (!projectId) {
      toast.error('Pilih project terlebih dahulu')
      return
    }

    const volume = prompt(`Masukkan volume untuk:\n${item.name}\n\nVolume (${item.unit}):`)
    if (!volume || parseFloat(volume) <= 0) {
      toast.error('Volume harus diisi dan lebih dari 0')
      return
    }

    try {
      setAddingItem(item.id)
      
      const response = await fetch(`/api/projects/${projectId}/work_items/from-ahsp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          analysisMasterId: item.id,
          volume: parseFloat(volume),
          applyBreakdown: true
        })
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to add')
      }

      toast.success('Item berhasil ditambahkan ke RAB')
      
      // Refresh workItem list
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
  const subtotal = workItem.reduce((sum, p) => sum + Number(p.totalCost), 0)
  const totalDuration = workItem.reduce((sum, p) => sum + (p.totalDuration || 0), 0)
  const overhead = subtotal * 0.10
  const profit = (subtotal + overhead) * 0.10
  const ppn = (subtotal + overhead + profit) * 0.11
  const total = subtotal + overhead + profit + ppn

  // Group by category
  const grouped = workItem.reduce((acc, item) => {
    if (!acc[item.category]) acc[item.category] = []
    acc[item.category].push(item)
    return acc
  }, {} as Record<string, PekerjaanItem[]>)

  if (!projectId) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <Calculator className="w-12 h-12 text-gray-300 mb-3" />
        <p className="text-gray-600">Silakan pilih project terlebih dahulu</p>
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
          <CardTitle className="text-lg">Cari Item WorkItem AHSP</CardTitle>
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
                        {item.category}
                      </Badge>
                      <span className="text-xs text-gray-500">{item.ahspCode}</span>
                    </div>
                    <div className="font-medium text-sm truncate">{item.name}</div>
                    <div className="text-sm text-gray-600 mt-1">
                      {formatCurrency(item.unitPrice)} / {item.unit}
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
          ) : workItem.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              Belum ada item pekerjaan. Cari dan tambahkan dari AHSP di atas.
            </div>
          ) : (
            <div className="space-y-6">
              {Object.entries(grouped).map(([category, items]) => {
                const subtotalKategori = items.reduce((sum, p) => sum + Number(p.totalCost), 0)
                const waktuKategori = items.reduce((sum, p) => sum + (p.totalDuration || 0), 0)
                
                return (
                  <div key={category}>
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="font-semibold text-lg capitalize">
                        {category.replace('_', ' ')}
                      </h3>
                      <div className="text-sm text-gray-600">
                        Subtotal: {formatCurrency(subtotalKategori)}
                        <span className="ml-3 text-gray-400">
                          · {formatDuration(waktuKategori)}
                        </span>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      {items.map((item, idx) => (
                        <div key={item.id} className="flex items-center gap-3 p-3 border rounded-lg">
                          <div className="flex-1 min-w-0">
                            <div className="font-medium text-sm truncate">{item.description}</div>
                            <div className="text-xs text-gray-500 mt-1">
                              {item.volume} {item.unit} × {formatCurrency(item.unitPrice)}
                            </div>
                            <div className="text-xs text-gray-400 mt-0.5">
                              Estimasi Waktu: {formatDuration(item.totalDuration)}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="font-semibold">{formatCurrency(item.totalCost)}</div>
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
                  <span className="font-medium">{formatDuration(totalDuration)}</span>
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
