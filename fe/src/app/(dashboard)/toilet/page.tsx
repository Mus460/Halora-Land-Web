"use client";

import { Bath } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function ToiletPage() {
  return (
    <WorkItemPage
      category="toilet"
      title="Toilet/Sanitair"
      description="Pekerjaan sanitair dan perlengkapan toilet"
      icon={Bath}
      formFields={[
        {
          name: "bathub",
          label: "Bathub",
          type: "select",
          options: [
            { value: "tidak", label: "Tidak" },
            { value: "ya", label: "Ya" },
          ],
        },
        {
          name: "kloset",
          label: "Jenis Kloset",
          type: "select",
          options: [
            { value: "tidak_ada", label: "Tidak Ada" },
            { value: "duduk", label: "Duduk" },
            { value: "jongkok", label: "Jongkok" },
          ],
        },
      ]}
    />
  );
}
