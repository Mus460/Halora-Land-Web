"use client";

import { MoveUpRight } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function TanggaPage() {
  return (
    <PekerjaanPage
      kategori="tangga"
      title="Tangga"
      description="Pekerjaan tangga dan railing"
      icon={MoveUpRight}
      proyekId={1}
    />
  );
}
