"use client";

import { useState } from "react";
import { MessageSquare, Plus, Send } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/shared/empty-state";
import { formatDate } from "@/lib/utils";
import { getFeedbackList } from "@/mock";
import { FEEDBACK_STATUS } from "@/lib/constants";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import toast from "react-hot-toast";
import type { Feedback } from "@/types";

export default function FeedbackPage() {
  const [data, setData] = useState<Feedback[]>(getFeedbackList());
  const [showForm, setShowForm] = useState(false);
  const [selected, setSelected] = useState<Feedback | null>(null);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Feedback & Support"
        description="Kirim feedback atau laporkan masalah"
        actions={
          <Button
            className="bg-amber-500 hover:bg-amber-600"
            onClick={() => setShowForm(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            Kirim Feedback
          </Button>
        }
      />

      {data.length > 0 ? (
        <div className="space-y-3">
          {data.map((item) => {
            const status = FEEDBACK_STATUS.find((s) => s.value === item.status);
            return (
              <Card
                key={item.id}
                className="cursor-pointer hover:shadow-md transition-shadow"
                onClick={() => setSelected(item)}
              >
                <CardContent className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <MessageSquare className="w-5 h-5 text-amber-500 mt-0.5" />
                      <div>
                        <h3 className="font-semibold text-gray-900">
                          {item.subject}
                        </h3>
                        <p className="text-sm text-gray-500 line-clamp-2 mt-1">
                          {item.message}
                        </p>
                        <p className="text-xs text-gray-400 mt-2">
                          {formatDate(item.createdAt)}
                        </p>
                      </div>
                    </div>
                    <Badge variant="outline">{status?.label}</Badge>
                  </div>
                  {item.replies && item.replies.length > 0 && (
                    <div className="mt-3 pt-3 border-t">
                      <p className="text-xs text-gray-500">
                        {item.replies.length} balasan
                      </p>
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      ) : (
        <EmptyState
          title="Belum ada feedback"
          description="Kirim feedback atau pertanyaan ke admin"
          action={
            <Button
              className="bg-amber-500 hover:bg-amber-600"
              onClick={() => setShowForm(true)}
            >
              <Plus className="w-4 h-4 mr-2" />
              Kirim Feedback
            </Button>
          }
        />
      )}

      {/* New Feedback Dialog */}
      <FeedbackFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        onSubmit={(data) => {
          const newFeedback: Feedback = {
            id: Date.now(),
            userId: 2,
            subject: data.subject || "",
            message: data.message || "",
            status: "open",
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            replies: [],
          };
          setData((prev) => [newFeedback, ...prev]);
          toast.success("Feedback berhasil dikirim");
        }}
      />

      {/* Detail Dialog */}
      {selected && (
        <Dialog open={!!selected} onOpenChange={() => setSelected(null)}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>{selected.subject}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="p-3 bg-gray-50 rounded-lg">
                <p className="text-sm">{selected.message}</p>
                <p className="text-xs text-gray-400 mt-2">
                  {formatDate(selected.createdAt)}
                </p>
              </div>
              {selected.replies?.map((reply) => (
                <div
                  key={reply.id}
                  className={`p-3 rounded-lg ${
                    reply.isAdmin
                      ? "bg-amber-50 border border-amber-200"
                      : "bg-gray-50"
                  }`}
                >
                  <p className="text-xs font-semibold mb-1">
                    {reply.isAdmin ? "Admin" : "Anda"}
                  </p>
                  <p className="text-sm">{reply.message}</p>
                  <p className="text-xs text-gray-400 mt-2">
                    {formatDate(reply.createdAt)}
                  </p>
                </div>
              ))}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}

function FeedbackFormDialog({
  open,
  onOpenChange,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { subject?: string; message?: string }) => void;
}) {
  const [form, setForm] = useState({ subject: "", message: "" });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form);
    onOpenChange(false);
    setForm({ subject: "", message: "" });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Kirim Feedback</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Subjek</Label>
            <Input
              value={form.subject}
              onChange={(e) => setForm({ ...form, subject: e.target.value })}
              placeholder="Judul feedback"
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Pesan</Label>
            <Textarea
              value={form.message}
              onChange={(e) => setForm({ ...form, message: e.target.value })}
              placeholder="Jelaskan feedback atau masalah Anda..."
              rows={4}
              required
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Batal
            </Button>
            <Button type="submit" className="bg-amber-500 hover:bg-amber-600">
              <Send className="w-4 h-4 mr-2" />
              Kirim
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
