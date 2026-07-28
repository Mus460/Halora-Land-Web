"use client";

import { Box } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { TIPE_PONDASI } from "@/lib/constants";

export default function PondasiPage() {
  return (
    <PekerjaanPage
      kategori="pondasi"
      title="Pondasi"
      description="Pekerjaan pondasi bangunan"
      icon={Box}
      proyekId={1}
      formFields={[
        {
          name: "tipe_pondasi",
          label: "Tipe Pondasi",
          type: "select",
          options: TIPE_PONDASI.map((t) => ({
            value: t,
            label: t.replace("_", " ").toUpperCase(),
          })),
        },
      ]}
    />
  );
}
