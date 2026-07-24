"use client";

import { Construction } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function BajaPage() {
  return (
    <PekerjaanPage
      kategori="baja"
      title="Baja Struktural"
      description="Pekerjaan rangka baja dan struktur baja"
      icon={Construction}
      initialData={getPekerjaanByKategori("baja")}
    />
  );
}
