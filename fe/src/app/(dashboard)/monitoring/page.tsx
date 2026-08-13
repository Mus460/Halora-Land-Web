"use client";
import { useProject } from "@/contexts/ProjectContext";

import { useState, useEffect } from "react";
import { ClipboardCheck, CheckCircle2, History, Pencil } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import toast from "react-hot-toast";
import { EmptyProjectState } from "@/components/shared/empty-project-state";
import { formatWeight } from "@/lib/utils";

interface MonItem {
  id: number;
  description: string;
  volume: string;
  unit: string;
  progress: number;
  weight: number;
  lastUpdated: string | null;
}

interface MonCategory {
  category: string;
  progress: number;
  items: MonItem[];
}

interface ProgressLog {
  id: number;
  workItemId: number;
  progress: number;
  note: string | null;
  createdAt: string;
}

export default function MonitoringPage() {
  const { currentProjectId: projectId, projectList, loading: proyekLoading } = useProject();

  if (proyekLoading) return <div className="p-8 text-center">Memuat data...</div>;
  if (projectList.length === 0) {
    return (
      <EmptyProjectState
        title="Belum Ada Data Progress"
        description="Buat proyek untuk memantau progress pekerjaan konstruksi"
      />
    );
  }

  const [monitoring, setMonitoring] = useState<MonCategory[]>([]);
  const [overall, setOverall] = useState(0);
  const [loading, setLoading] = useState(true);
  const [savingId, setSavingId] = useState<number | null>(null);
  const [confirm, setConfirm] = useState<{ item: MonItem; value: number } | null>(null);
  const [drafts, setDrafts] = useState<Record<number, number>>({});
  const [logModal, setLogModal] = useState<{ item: MonItem; logs: ProgressLog[] } | null>(null);
  const [logLoading, setLogLoading] = useState(false);

  const handleDraftChange = (item: MonItem, raw: string) => {
    const value = Number(raw);
    setDrafts((prev) => ({
      ...prev,
      [item.id]: Number.isNaN(value) ? 0 : Math.min(100, Math.max(0, value)),
    }));
  };

  const requestUpdate = (item: MonItem) => {
    const value = drafts[item.id];
    if (value === undefined || Number.isNaN(value)) {
      toast.error("Masukkan nilai progress");
      return;
    }
    if (value <= item.progress) {
      toast.error("Progress baru harus lebih tinggi dari progress saat ini");
      return;
    }
    if (value > 100) {
      toast.error("Progress maksimal 100%");
      return;
    }
    setConfirm({ item, value });
  };

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/monitoring?projectId=${projectId}`);
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setMonitoring(result.monitoring || []);
      setOverall(result.overall ?? 0);
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (projectId) {
      fetchData();
    }
  }, [projectId]);

  const saveProgress = async (item: MonItem, value: number) => {
    setSavingId(item.id);
    try {
      const response = await fetch(`/api/work-items/${item.id}/progress`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ progress: value }),
      });
      const result = await response.json();
      if (!response.ok) throw new Error(result.error || "Gagal menyimpan progress");
      const updatedAt =
        result.log?.createdAt ??
        new Date().toISOString();
      setMonitoring((prev) =>
        prev.map((cat) => ({
          ...cat,
          items: cat.items.map((i) =>
            i.id === item.id ? { ...i, progress: value, lastUpdated: updatedAt } : i
          ),
        }))
      );
      setDrafts((prev) => {
        const next = { ...prev };
        delete next[item.id];
        return next;
      });
      toast.success(`Progress diperbarui ke ${value}%`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Gagal menyimpan progress");
    } finally {
      setSavingId((cur) => (cur === item.id ? null : cur));
    }
  };

  const handleConfirm = () => {
    if (!confirm) return;
    const { item, value } = confirm;
    setConfirm(null);
    saveProgress(item, value);
  };

  const openHistory = async (item: MonItem) => {
    setLogModal({ item, logs: [] });
    setLogLoading(true);
    try {
      const response = await fetch(`/api/work-items/${item.id}/progress-logs`);
      if (!response.ok) throw new Error("Failed to fetch");
      const logs = await response.json();
      setLogModal({ item, logs: Array.isArray(logs) ? logs : [] });
    } catch {
      toast.error("Gagal memuat riwayat progress");
    } finally {
      setLogLoading(false);
    }
  };

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

  if (loading) {
    return <div className="p-8 text-center">Memuat data...</div>;
  }

  const totalItems = monitoring.reduce(
    (sum, cat) => sum + cat.items.length,
    0
  );
  const completedItems = monitoring.reduce(
    (sum, cat) => sum + cat.items.filter((i) => i.progress === 100).length,
    0
  );
  const overallProgress = overall;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Progress Monitoring"
        description="Tracking progres pekerjaan per kategori"
      />

      {/* Overall Progress */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="w-14 h-14 bg-amber-100 rounded-full flex items-center justify-center">
              <ClipboardCheck className="w-7 h-7 text-amber-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Progres Keseluruhan</p>
              <p className="text-3xl font-bold text-gray-900">
                {overallProgress}%
              </p>
            </div>
          </div>
          <Progress value={overallProgress} className="h-3" />
          <p className="text-xs text-gray-500 mt-2">
            {completedItems} dari {totalItems} item selesai — progress tidak dapat diturunkan, isi nilai lebih tinggi dari sebelumnya
          </p>
        </CardContent>
      </Card>

      {/* Per Category */}
      <div className="space-y-4">
        {monitoring.map((category) => {
          const kategoriTotal = category.items.length;
          const kategoriDone = category.items.filter(
            (i) => i.progress === 100
          ).length;
          const kategoriProgress = category.progress;

          return (
            <Card key={category.category}>
              <CardHeader className="py-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-semibold">
                    {category.category}
                  </CardTitle>
                  <Badge variant="outline">
                    {kategoriDone}/{kategoriTotal} selesai
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <Progress value={kategoriProgress} className="h-2" />
                {category.items.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between gap-4 py-2 border-b last:border-0"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium truncate flex items-center gap-2">
                        {item.progress === 100 && (
                          <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0" />
                        )}
                        {item.description}
                      </p>
                      <p className="text-xs text-gray-500">
                        {item.volume} {item.unit}
                      </p>
                      {item.lastUpdated ? (
                        <p className="text-xs text-gray-400 tabular-nums">
                          Terakhir diperbarui: {formatDate(item.lastUpdated)}
                        </p>
                      ) : (
                        <p className="text-xs text-gray-400">
                          Belum ada pembaruan
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <span className="text-xs text-gray-500">
                        Bobot {formatWeight(item.weight)}
                      </span>
                      <div className="flex items-center gap-1">
                        <input
                          type="number"
                          min={item.progress + 1}
                          max={100}
                          value={drafts[item.id] ?? item.progress}
                          onChange={(e) => handleDraftChange(item, e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") requestUpdate(item);
                          }}
                          className="w-20 h-9 rounded-lg border border-input bg-transparent px-2 text-sm text-right tabular-nums focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:bg-input/50"
                          aria-label={`Progress ${item.description}`}
                        />
                        <span className="text-sm text-gray-500">%</span>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={savingId === item.id}
                        onClick={() => requestUpdate(item)}
                        className="h-9"
                      >
                        <Pencil className="w-3.5 h-3.5 mr-1" />
                        {savingId === item.id ? "Menyimpan..." : "Perbarui"}
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-gray-400 hover:text-amber-600"
                        onClick={() => openHistory(item)}
                        title="Riwayat progress"
                      >
                        <History className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Confirm Update Dialog */}
      <Dialog open={confirm !== null} onOpenChange={(open) => !open && setConfirm(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base">Konfirmasi Update Progress</DialogTitle>
            <DialogDescription className="line-clamp-2">
              {confirm?.item.description}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between px-3 py-2 bg-gray-50 rounded-lg">
              <span className="text-gray-500">Progress saat ini</span>
              <span className="font-semibold tabular-nums">{confirm?.item.progress}%</span>
            </div>
            <div className="flex items-center justify-between px-3 py-2 bg-gray-50 rounded-lg">
              <span className="text-gray-500">Progress baru</span>
              <span className="font-semibold tabular-nums text-amber-600">{confirm?.value}%</span>
            </div>
            <p className="text-xs text-gray-400">
              Perubahan ini tercatat di riwayat dengan timestamp dan tidak dapat dibatalkan.
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirm(null)}>
              Batal
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={savingId === confirm?.item.id}
              className="bg-amber-500 hover:bg-amber-600"
            >
              Konfirmasi
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Progress History Dialog */}
      <Dialog open={logModal !== null} onOpenChange={(open) => !open && setLogModal(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base">Riwayat Progress</DialogTitle>
            <DialogDescription className="line-clamp-2">
              {logModal?.item.description}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2 max-h-80 overflow-y-auto pr-1">
            {logLoading ? (
              <p className="text-sm text-gray-400 py-4 text-center">Memuat riwayat...</p>
            ) : logModal && logModal.logs.length === 0 ? (
              <p className="text-sm text-gray-400 py-4 text-center">
                Belum ada perubahan progress
              </p>
            ) : (
              logModal?.logs.map((log) => (
                <div
                  key={log.id}
                  className="flex items-center justify-between gap-3 py-2 px-3 bg-gray-50 rounded-lg"
                >
                  <div className="min-w-0">
                    <p
                      className={`text-sm font-semibold tabular-nums ${
                        log.progress === 100
                          ? "text-emerald-600"
                          : log.progress > 0
                          ? "text-amber-600"
                          : "text-gray-500"
                      }`}
                    >
                      {log.progress}%
                    </p>
                    {log.note && (
                      <p className="text-xs text-gray-500 truncate">{log.note}</p>
                    )}
                  </div>
                  <p className="text-xs text-gray-400 shrink-0 tabular-nums">
                    {formatDate(log.createdAt)}
                  </p>
                </div>
              ))
            )}
          </div>
          {!logLoading && logModal && logModal.logs.length > 0 && (
            <p className="text-xs text-gray-400 text-right">
              Terakhir diperbarui: {formatDate(logModal.logs[logModal.logs.length - 1].createdAt)}
            </p>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
