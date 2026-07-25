"use client";

import { Bath } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function ToiletPage() {
  return (
    <PekerjaanPage
      kategori="toilet"
      title="Toilet/Sanitair"
      description="Pekerjaan sanitair dan perlengkapan toilet"
      icon={Bath}
      proyekId={1}
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
