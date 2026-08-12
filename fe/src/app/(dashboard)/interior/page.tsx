"use client";

import { Sofa } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function InteriorPage() {
  return (
    <WorkItemPage
      category="interior"
      title="Interior"
      description="Pekerjaan interior dan furnitur"
      icon={Sofa}
    />
  );
}
