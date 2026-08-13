"use client";

import { Grid3X3 } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";
import { JENIS_DINDING } from "@/lib/constants";

export default function DindingPage() {
  return (
    <WorkItemPage
      category="wall"
      title="Pas. Dinding"
      description="Pekerjaan pasangan dinding"
      icon={Grid3X3}
      formFields={[
        {
          name: "jenis_dinding",
          label: "Jenis Dinding",
          type: "select",
          required: true,
          options: JENIS_DINDING.map((j) => ({ value: j, label: j })),
        },
      ]}
    />
  );
}
