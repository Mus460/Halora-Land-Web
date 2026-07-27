"use client"

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { PageHeader } from '@/components/shared/page-header'
import { Database, Download, CheckCircle, Circle, Loader2 } from 'lucide-react'
import toast from 'react-hot-toast'

interface ImportStatus {
  sheetName: string
  kategori: string
  imported: boolean
  count: number
}

export default function AdminAHSPPage() {
  const [status, setStatus] = useState<ImportStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [importing, setImporting] = useState<string | null>(null)

  useEffect(() => {
    fetchStatus()
  }, [])

  const fetchStatus = async () => {
    try {
      setLoading(true)
      const response = await fetch('/api/admin/ahsp/import')
      if (!response.ok) throw new Error('Failed to fetch')
      const data = await response.json()
      setStatus(data.status || [])
    } catch (error) {
      console.error('Fetch status error:', error)
      toast.error('Gagal memuat status import')
    } finally {
      setLoading(false)
    }
  }

  const handleImport = async (sheetName: string, forceReimport = false) => {
    try {
      setImporting(sheetName)
      
      const response = await fetch('/api/admin/ahsp/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sheetName, forceReimport })
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.error || 'Import failed')
      }

      if (data.skipped) {
        toast.success(`${sheetName} sudah di-import (${data.count} items)`)
      } else {
        toast.success(`Berhasil import ${data.imported} items dari ${sheetName}`)
      }

      // Refresh status
      await fetchStatus()

    } catch (error: any) {
      console.error('Import error:', error)
      toast.error(error.message || 'Gagal import')
    } finally {
      setImporting(null)
    }
  }

  const handleImportAll = async () => {
    const notImported = status.filter(s => !s.imported)
    
    if (notImported.length === 0) {
      toast.error('Semua kategori sudah di-import')
      return
    }

    if (!confirm(`Import ${notImported.length} kategori? Proses ini bisa memakan waktu beberapa menit.`)) {
      return
    }

    for (const item of notImported) {
      await handleImport(item.sheetName)
    }

    toast.success('Semua kategori berhasil di-import!')
  }

  const totalImported = status.filter(s => s.imported).length
  const totalCount = status.reduce((sum, s) => sum + s.count, 0)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-amber-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="AHSP Import Manager"
        description="Import data AHSP PUPR 2026 ke database"
        actions={
          <Button
            onClick={handleImportAll}
            disabled={importing !== null}
            className="bg-amber-500 hover:bg-amber-600"
          >
            <Download className="w-4 h-4 mr-2" />
            Import Semua ({status.filter(s => !s.imported).length})
          </Button>
        }
      />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-600">
              Total Kategori
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{status.length}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-600">
              Sudah Di-import
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">{totalImported}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-gray-600">
              Total Items
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-amber-600">{totalCount}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Status Import per Kategori</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {status.map((item) => (
              <div
                key={item.sheetName}
                className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50"
              >
                <div className="flex items-center gap-3">
                  {item.imported ? (
                    <CheckCircle className="w-5 h-5 text-green-600" />
                  ) : (
                    <Circle className="w-5 h-5 text-gray-300" />
                  )}
                  <div>
                    <div className="font-medium">{item.sheetName}</div>
                    <div className="text-sm text-gray-500">
                      Kategori: {item.kategori} • {item.count} items
                    </div>
                  </div>
                </div>

                <div className="flex gap-2">
                  {item.imported ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleImport(item.sheetName, true)}
                      disabled={importing !== null}
                    >
                      {importing === item.sheetName ? (
                        <>
                          <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                          Re-importing...
                        </>
                      ) : (
                        <>
                          <Database className="w-4 h-4 mr-2" />
                          Re-import
                        </>
                      )}
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      onClick={() => handleImport(item.sheetName)}
                      disabled={importing !== null}
                      className="bg-amber-500 hover:bg-amber-600"
                    >
                      {importing === item.sheetName ? (
                        <>
                          <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                          Importing...
                        </>
                      ) : (
                        <>
                          <Download className="w-4 h-4 mr-2" />
                          Import Now
                        </>
                      )}
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
