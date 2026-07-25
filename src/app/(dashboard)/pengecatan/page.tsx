"use client";

import { Paintbrush } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function PengecatanPage() {
  return (
    <PekerjaanPage
      kategori="pengecatan"
      title="Cat & Plafon"
      description="Pekerjaan pengecatan dan plafon"
      icon={Paintbrush}
      proyekId={1}
    />
  );
}
