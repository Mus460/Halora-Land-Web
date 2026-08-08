"use client";

import { Pickaxe } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";
import { TIPE_BETON, LEVEL_PEKERJAAN, MUTU_BETON } from "@/lib/constants";

export default function BetonPage() {
  return (
    <WorkItemPage
      category="concrete"
      title="Beton Struktur"
      description="Pekerjaan beton bertulang (kolom, balok, plat, dll)"
      icon={Pickaxe}
      showLevelPekerjaan={true}
      showTipePekerjaan={true}
      tipeOptions={TIPE_BETON}
      formFields={[
        {
          name: "mutu_beton",
          label: "Muti Beton",
          type: "select",
          options: MUTU_BETON.map((m) => ({ value: m, label: m })),
        },
      ]}
    />
  );
}
