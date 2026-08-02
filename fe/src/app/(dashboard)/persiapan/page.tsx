"use client";

import { Ruler } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { TIPE_PERSIAPAN } from "@/lib/constants";

export default function PersiapanPage() {
  return (
    <PekerjaanPage
      kategori="persiapan"
      title="Persiapan"
      description="Pekerjaan persiapan lokasi konstruksi"
      icon={Ruler}
      proyekId={1}
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
