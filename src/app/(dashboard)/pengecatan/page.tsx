"use client";

import { Paintbrush } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function PengecatanPage() {
  return (
    <PekerjaanPage
      kategori="pengecatan"
      title="Cat & Plafon"
      description="Pekerjaan pengecatan dan plafon"
      icon={Paintbrush}
      initialData={getPekerjaanByKategori("pengecatan")}
    />
  );
}
