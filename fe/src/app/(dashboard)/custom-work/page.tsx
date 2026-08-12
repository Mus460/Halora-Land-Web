"use client";

import { PenTool } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function PekerjaanCustomPage() {
  return (
    <WorkItemPage
      category="custom"
      title="Pekerjaan Custom"
      description="Pekerjaan di luar kategori standar"
      icon={PenTool}
    />
  );
}
