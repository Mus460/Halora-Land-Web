"use client";

import { Sofa } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function InteriorPage() {
  return (
    <PekerjaanPage
      kategori="interior"
      title="Interior"
      description="Pekerjaan interior dan furnitur"
      icon={Sofa}
      initialData={getPekerjaanByKategori("interior")}
    />
  );
}
