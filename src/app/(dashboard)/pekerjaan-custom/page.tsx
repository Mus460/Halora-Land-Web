"use client";

import { PenTool } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function PekerjaanCustomPage() {
  return (
    <PekerjaanPage
      kategori="custom"
      title="Pekerjaan Custom"
      description="Pekerjaan di luar kategori standar"
      icon={PenTool}
      proyekId={1}
    />
  );
}
