"use client";

import { House } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";
import { JENIS_RANGKA_ATAP, JENIS_ATAP, BENTUK_ATAP } from "@/lib/constants";

export default function AtapPage() {
  return (
    <PekerjaanPage
      kategori="atap"
      title="Pek. Atap"
      description="Pekerjaan rangka atap dan penutup atap"
      icon={House}
      initialData={getPekerjaanByKategori("atap")}
      formFields={[
        {
          name: "jenis_rangka",
          label: "Jenis Rangka",
          type: "select",
          options: JENIS_RANGKA_ATAP.map((j) => ({ value: j, label: j })),
        },
        {
          name: "jenis_atap",
          label: "Jenis Atap",
          type: "select",
          options: JENIS_ATAP.map((j) => ({ value: j, label: j })),
        },
        {
          name: "bentuk_atap",
          label: "Bentuk Atap",
          type: "select",
          options: BENTUK_ATAP.map((b) => ({ value: b, label: b })),
        },
      ]}
    />
  );
}
