"use client";

import { Construction } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function BajaPage() {
  return (
    <WorkItemPage
      category="steel"
      title="Baja Struktural"
      description="Pekerjaan rangka baja dan struktur baja"
      icon={Construction}
    />
  );
}
