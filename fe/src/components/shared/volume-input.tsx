"use client";

import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";

interface VolumeInputProps {
  panjang: number;
  lebar: number;
  tinggi: number;
  onChange: (values: {
    panjang: number;
    lebar: number;
    tinggi: number;
  }) => void;
  label?: string;
  disabled?: boolean;
  showTinggi?: boolean;
  className?: string;
}

export function VolumeInput({
  panjang,
  lebar,
  tinggi,
  onChange,
  label,
  disabled,
  showTinggi = true,
  className,
}: VolumeInputProps) {
  const handleChange = (field: "panjang" | "lebar" | "tinggi", value: string) => {
    const num = Number(value.replace(/[^0-9.]/g, "")) || 0;
    onChange({ panjang, lebar, tinggi, [field]: num });
  };

  const volume = showTinggi
    ? panjang * lebar * tinggi
    : panjang * lebar;

  return (
    <div className={cn("space-y-2", className)}>
      {label && <Label>{label}</Label>}
      <div className="grid grid-cols-3 gap-2">
        <div>
          <Label className="text-xs text-gray-500">P (m)</Label>
          <Input
            value={panjang || ""}
            onChange={(e) => handleChange("panjang", e.target.value)}
            placeholder="P"
            disabled={disabled}
            inputMode="decimal"
          />
        </div>
        <div>
          <Label className="text-xs text-gray-500">L (m)</Label>
          <Input
            value={lebar || ""}
            onChange={(e) => handleChange("lebar", e.target.value)}
            placeholder="L"
            disabled={disabled}
            inputMode="decimal"
          />
        </div>
        {showTinggi && (
          <div>
            <Label className="text-xs text-gray-500">T (m)</Label>
            <Input
              value={tinggi || ""}
              onChange={(e) => handleChange("tinggi", e.target.value)}
              placeholder="T"
              disabled={disabled}
              inputMode="decimal"
            />
          </div>
        )}
      </div>
      {volume > 0 && (
        <p className="text-xs text-gray-500">
          Volume: <span className="font-semibold text-gray-700">{volume.toFixed(2)}</span> m³
        </p>
      )}
    </div>
  );
}
