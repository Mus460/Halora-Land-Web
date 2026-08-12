"use client";

import { Layers } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function PlesteranPage() {
  return (
    <WorkItemPage
      category="plastering"
      title="Plesteran"
      description="Pekerjaan plesteran dinding"
      icon={Layers}
    />
  );
}
