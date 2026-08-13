"use client";

import { LayoutGrid } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";
import { JENIS_KERAMIK } from "@/lib/constants";

export default function KeramikPage() {
  return (
    <WorkItemPage
      category="tiles"
      title="Lantai/Keramik"
      description="Pekerjaan lantai keramik dan granit"
      icon={LayoutGrid}
      formFields={[
        {
          name: "jenis_keramik",
          label: "Jenis/Ukuran Keramik",
          type: "select",
          required: true,
          options: JENIS_KERAMIK.map((j) => ({ value: j, label: j })),
        },
      ]}
    />
  );
}
