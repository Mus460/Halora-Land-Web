"use client";

import { Layers } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function PlesteranPage() {
  return (
    <PekerjaanPage
      kategori="plesteran"
      title="Plesteran"
      description="Pekerjaan plesteran dinding"
      icon={Layers}
      proyekId={1}
    />
  );
}
