"use client";

import { Droplet } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function AcianPage() {
  return (
    <PekerjaanPage
      kategori="acian"
      title="Acian"
      description="Pekerjaan acian dinding"
      icon={Droplet}
      proyekId={1}
    />
  );
}
