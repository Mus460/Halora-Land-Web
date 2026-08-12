"use client";

import { Ruler } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";
import { TIPE_PERSIAPAN } from "@/lib/constants";

export default function PersiapanPage() {
  return (
    <WorkItemPage
      category="preparation"
      title="Persiapan"
      description="Pekerjaan persiapan lokasi konstruksi"
      icon={Ruler}
      formFields={[
        {
          name: "tipe_persiapan",
          label: "Tipe Persiapan",
          type: "select",
          options: TIPE_PERSIAPAN.map((t) => ({ value: t, label: t })),
        },
      ]}
    />
  );
}
