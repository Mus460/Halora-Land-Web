"use client";

import { LayoutGrid } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";
import { JENIS_KERAMIK } from "@/lib/constants";

export default function KeramikPage() {
  return (
    <PekerjaanPage
      kategori="keramik"
      title="Lantai/Keramik"
      description="Pekerjaan lantai keramik dan granit"
      icon={LayoutGrid}
      initialData={getPekerjaanByKategori("keramik")}
      formFields={[
        {
          name: "jenis_keramik",
          label: "Jenis/Ukuran Keramik",
          type: "select",
          options: JENIS_KERAMIK.map((j) => ({ value: j, label: j })),
        },
      ]}
    />
  );
}
