"use client";

import { Box } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";
import { TIPE_PONDASI } from "@/lib/constants";

export default function PondasiPage() {
  return (
    <WorkItemPage
      category="foundation"
      title="Pondasi"
      description="Pekerjaan pondasi bangunan"
      icon={Box}
      formFields={[
        {
          name: "tipe_pondasi",
          label: "Tipe Pondasi",
          type: "select",
          required: true,
          options: TIPE_PONDASI.map((t) => ({
            value: t,
            label: t.replace("_", " ").toUpperCase(),
          })),
        },
      ]}
    />
  );
}
