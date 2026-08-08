"use client";

import { DoorOpen } from "lucide-react";
import { WorkItemPage } from "@/components/work-items/work-item-page";

export default function PintuPage() {
  return (
    <WorkItemPage
      category="doors"
      title="Kusen/Pintu"
      description="Pekerjaan kusen, pintu, dan jendela"
      icon={DoorOpen}
    />
  );
}
