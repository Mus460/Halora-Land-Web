"use client";

import { Umbrella } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";
import { JENIS_KANOPI } from "@/lib/constants";

export default function KanopiPage() {
  return (
    <WorkItemPage
      category="canopy"
      title="Kanopi"
      description="Pekerjaan kanopi dan atap tambahan"
      icon={Umbrella}
      formFields={[
        {
          name: "jenis_kanopi",
          label: "Jenis Kanopi",
          type: "select",
          required: true,
          options: JENIS_KANOPI.map((j) => ({ value: j, label: j })),
        },
      ]}
    />
  );
}
