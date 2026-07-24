"use client";

import { Map } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function PavingPage() {
  return (
    <PekerjaanPage
      kategori="paving"
      title="Paving & Halaman"
      description="Pekerjaan paving block dan penghijauan"
      icon={Map}
      initialData={getPekerjaanByKategori("paving")}
    />
  );
}
