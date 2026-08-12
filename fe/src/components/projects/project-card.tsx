"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Building2,
  Handshake,
  CheckCircle2,
  RotateCcw,
  MoreVertical,
  Trash2,
  MapPin,
  Calendar,
  DollarSign,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { formatCurrency, formatDate, formatTimeline } from "@/lib/utils";
import type { Project } from "@/types";

interface ProjectCardProps {
  project: Project;
  onDelete: (id: number) => void;
  onSetActive: (id: number) => void;
  onSetPitching: (id: number, isPitching: boolean) => void;
  onSetDone: (id: number, isDone: boolean) => void;
  isActive?: boolean;
}

export function ProjectCard({
  project,
  onDelete,
  onSetActive,
  onSetPitching,
  onSetDone,
  isActive,
}: ProjectCardProps) {
  const [showDelete, setShowDelete] = useState(false);

  return (
    <>
      <Card
        className={`hover:shadow-md transition-shadow ${
          isActive ? "ring-2 ring-amber-500" : ""
        }`}
      >
        <CardContent className="p-5">
          <div className="flex items-start justify-between mb-3">
            <Link href={`/projects/${project.id}`} className="flex-1 min-w-0">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center shrink-0">
                  <Building2 className="w-5 h-5 text-amber-600" />
                </div>
                <div className="min-w-0">
                  <h3 className="font-semibold text-gray-900 truncate">
                    {project.name}
                  </h3>
                  <div className="flex items-center gap-1.5 mt-1">
                    <Badge variant="outline" className="text-xs">
                      {project.type === "building" ? "Gedung" : "Infrastruktur"}
                    </Badge>
                    {project.isDone && (
                      <Badge variant="default" className="text-xs bg-emerald-600">
                        Selesai
                      </Badge>
                    )}
                    {project.isPitching && (
                      <Badge variant="secondary" className="text-xs bg-amber-100 text-amber-700">
                        Pitching
                      </Badge>
                    )}
                  </div>
                </div>
              </div>
            </Link>
            <DropdownMenu>
              <DropdownMenuTrigger className="inline-flex items-center justify-center h-8 w-8 rounded-md hover:bg-gray-100 transition-colors">
                <MoreVertical className="w-4 h-4" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {!isActive && (project.isPitching || project.isDone) && (
                  <DropdownMenuItem onClick={() => onSetActive(project.id)}>
                    <Building2 className="w-4 h-4 mr-2" />
                    Jadikan Aktif
                  </DropdownMenuItem>
                )}
                {!project.isPitching && !project.isDone && (
                  <DropdownMenuItem
                    onClick={() => onSetPitching(project.id, true)}
                  >
                    <Handshake className="w-4 h-4 mr-2" />
                    Jadikan Pitching
                  </DropdownMenuItem>
                )}
                {!project.isDone && (
                  <DropdownMenuItem onClick={() => onSetDone(project.id, true)}>
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                    Tandai Selesai
                  </DropdownMenuItem>
                )}
                {project.isDone && (
                  <DropdownMenuItem onClick={() => onSetDone(project.id, false)}>
                    <RotateCcw className="w-4 h-4 mr-2" />
                    Buka Kembali
                  </DropdownMenuItem>
                )}
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="text-red-600"
                  onClick={() => setShowDelete(true)}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Hapus
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <Link href={`/projects/${project.id}`}>
            <div className="space-y-2 text-sm text-gray-600">
              {project.location && (
                <div className="flex items-center gap-2">
                  <MapPin className="w-4 h-4 text-gray-400" />
                  <span className="truncate">{project.location}</span>
                </div>
              )}
              {(project.timelineMonths > 0 || project.timelineDays > 0) && (
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-gray-400" />
                  <span>{formatTimeline(project.timelineMonths, project.timelineDays)}</span>
                </div>
              )}
              {project.contractValue && (
                <div className="flex items-center gap-2">
                  <DollarSign className="w-4 h-4 text-gray-400" />
                  <span className="font-semibold text-gray-900">
                    {formatCurrency(project.contractValue)}
                  </span>
                </div>
              )}
            </div>
          </Link>

          <div className="mt-4 pt-3 border-t flex items-center justify-between text-xs text-gray-500">
            <span>{project._count?.work_items || 0} item pekerjaan</span>
            <span>Update: {formatDate(project.updatedAt)}</span>
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={showDelete}
        onOpenChange={setShowDelete}
        title="Hapus Project"
        description={`Apakah Anda yakin ingin menghapus "${project.name}"? Semua data terkait akan dihapus.`}
        confirmText="Ya, Hapus"
        variant="destructive"
        onConfirm={() => {
          onDelete(project.id);
          setShowDelete(false);
        }}
      />
    </>
  );
}
