"use client";

import { DoorOpen } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";
import { getPekerjaanByKategori } from "@/mock";

export default function PintuPage() {
  return (
    <PekerjaanPage
      kategori="pintu"
      title="Kusen/Pintu"
      description="Pekerjaan kusen, pintu, dan jendela"
      icon={DoorOpen}
      initialData={getPekerjaanByKategori("pintu")}
    />
  );
}
