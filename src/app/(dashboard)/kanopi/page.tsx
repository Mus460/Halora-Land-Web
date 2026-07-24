"use client";

import { Umbrella } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";
import { JENIS_KANOPI } from "@/lib/constants";

export default function KanopiPage() {
  return (
    <PekerjaanPage
      kategori="kanopi"
      title="Kanopi"
      description="Pekerjaan kanopi dan atap tambahan"
      icon={Umbrella}
      initialData={getPekerjaanByKategori("kanopi")}
      formFields={[
        {
          name: "jenis_kanopi",
          label: "Jenis Kanopi",
          type: "select",
          options: JENIS_KANOPI.map((j) => ({ value: j, label: j })),
        },
      ]}
    />
  );
}
