"use client";

import { Zap } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function MEPPage() {
  return (
    <PekerjaanPage
      kategori="mep"
      title="Instalasi MEP"
      description="Pekerjaan mekanikal, elektrikal, dan plumbing"
      icon={Zap}
      proyekId={1}
    />
  );
}
