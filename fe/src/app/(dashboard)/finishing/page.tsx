"use client";

import { Droplet } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function AcianPage() {
  return (
    <WorkItemPage
      category="finishing"
      title="Acian"
      description="Pekerjaan acian dinding"
      icon={Droplet}
    />
  );
}
