"use client";

import { MoveUpRight } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function TanggaPage() {
  return (
    <PekerjaanPage
      kategori="tangga"
      title="Tangga"
      description="Pekerjaan tangga dan railing"
      icon={MoveUpRight}
      initialData={getPekerjaanByKategori("tangga")}
    />
  );
}
