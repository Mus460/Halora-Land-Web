"use client";

import { MoveUpRight } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function TanggaPage() {
  return (
    <WorkItemPage
      category="stairs"
      title="Tangga"
      description="Pekerjaan tangga dan railing"
      icon={MoveUpRight}
    />
  );
}
