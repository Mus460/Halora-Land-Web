"use client";

import { Grid3X3 } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { JENIS_DINDING } from "@/lib/constants";

export default function DindingPage() {
  return (
    <PekerjaanPage
      kategori="dinding"
      title="Pas. Dinding"
      description="Pekerjaan pasangan dinding"
      icon={Grid3X3}
      proyekId={1}
      formFields={[
        {
          name: "jenis_dinding",
          label: "Jenis Dinding",
          type: "select",
          options: JENIS_DINDING.map((j) => ({ value: j, label: j })),
        },
      ]}
    />
  );
}
