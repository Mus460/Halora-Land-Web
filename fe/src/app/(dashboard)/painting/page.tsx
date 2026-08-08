"use client";

import { Paintbrush } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function PengecatanPage() {
  return (
    <WorkItemPage
      category="painting"
      title="Cat & Plafon"
      description="Pekerjaan pengecatan dan plafon"
      icon={Paintbrush}
    />
  );
}
