"use client";

import { Map } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function PavingPage() {
  return (
    <WorkItemPage
      category="paving"
      title="Paving & Halaman"
      description="Pekerjaan paving block dan penghijauan"
      icon={Map}
    />
  );
}
