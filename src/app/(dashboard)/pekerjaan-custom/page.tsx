"use client";

import { PenTool } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function PekerjaanCustomPage() {
  return (
    <PekerjaanPage
      kategori="custom"
      title="Pekerjaan Custom"
      description="Pekerjaan di luar kategori standar"
      icon={PenTool}
      initialData={getPekerjaanByKategori("custom")}
    />
  );
}
