"use client";

import { Droplet } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function AcianPage() {
  return (
    <PekerjaanPage
      kategori="acian"
      title="Acian"
      description="Pekerjaan acian dinding"
      icon={Droplet}
      initialData={getPekerjaanByKategori("acian")}
    />
  );
}
