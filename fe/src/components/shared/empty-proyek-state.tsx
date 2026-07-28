"use client";

import Link from "next/link";
import { Building2, Calculator, TrendingUp, ClipboardCheck, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "./empty-state";

interface EmptyProyekStateProps {
  title?: string;
  description?: string;
}

const features = [
  { icon: Calculator, label: "Hitung RAB" },
  { icon: TrendingUp, label: "Analisa Biaya" },
  { icon: ClipboardCheck, label: "Track Progress" },
];

export function EmptyProyekState({
  title = "Belum Ada Proyek",
  description = "Buat proyek untuk melihat analisa dan laporan",
}: EmptyProyekStateProps) {
  return (
    <div className="bg-white rounded-lg border-2 border-dashed border-gray-300 p-8">
      <EmptyState
        className="min-h-[50vh]"
        title={title}
        description={description}
        icon={<Building2 className="w-8 h-8 text-amber-600" />}
        iconClassName="bg-amber-100"
        action={
          <>
            <Button
              nativeButton={false}
              render={<Link href="/proyek" />}
              className="h-11 px-6 bg-amber-500 hover:bg-amber-600 text-white mb-8"
            >
              <Plus className="w-5 h-5" />
              Buat Proyek Pertama
            </Button>
            
            <div className="grid grid-cols-3 gap-6 mt-4">
              {features.map(({ icon: Icon, label }) => (
                <div key={label} className="text-center">
                  <div className="w-12 h-12 bg-amber-50 rounded-lg flex items-center justify-center mx-auto mb-2">
                    <Icon className="w-6 h-6 text-amber-600" />
                  </div>
                  <p className="text-sm text-gray-600">{label}</p>
                </div>
              ))}
            </div>
          </>
        }
      />
    </div>
  );
}
