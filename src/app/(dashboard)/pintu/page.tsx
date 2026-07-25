"use client";

import { DoorOpen } from "lucide-react";
import { PekerjaanPage } from "@/components/pekerjaan/pekerjaan-page";

export default function PintuPage() {
  return (
    <PekerjaanPage
      kategori="pintu"
      title="Kusen/Pintu"
      description="Pekerjaan kusen, pintu, dan jendela"
      icon={DoorOpen}
      proyekId={1}
    />
  );
}
