"use client";

import { Zap } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function MEPPage() {
  return (
    <WorkItemPage
      category="mep"
      title="Instalasi MEP"
      description="Pekerjaan mekanikal, elektrikal, dan plumbing"
      icon={Zap}
    />
  );
}
