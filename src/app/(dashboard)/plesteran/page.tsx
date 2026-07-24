"use client";

import { Layers } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function PlesteranPage() {
  return (
    <PekerjaanPage
      kategori="plesteran"
      title="Plesteran"
      description="Pekerjaan plesteran dinding"
      icon={Layers}
      initialData={getPekerjaanByKategori("plesteran")}
    />
  );
}
