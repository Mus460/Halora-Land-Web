"use client";

import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";

interface VolumeInputProps {
  unit?: string;
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
  className?: string;
}

const DIM_FIELDS: Record<string, "p" | "pl" | "plt" | null> = {
  m3: "plt",
  m2: "pl",
  m1: "p",
  m: "p",
  "m'": "p",
};

export function VolumeInput({
  unit = "m3",
  panjang,
  lebar,
  tinggi,
  onChange,
  label,
  disabled,
  className,
}: VolumeInputProps) {
  const dims = DIM_FIELDS[unit] ?? null;
  const handleChange = (field: "panjang" | "lebar" | "tinggi", value: string) => {
    const num = Number(value.replace(/[^0-9.]/g, "")) || 0;
    onChange({ panjang, lebar, tinggi, [field]: num });
  };

  const volume =
    dims === "plt" && panjang > 0 && lebar > 0 && tinggi > 0
      ? panjang * lebar * tinggi
      : dims === "pl" && panjang > 0 && lebar > 0
        ? panjang * lebar
        : dims === "p" && panjang > 0
          ? panjang
          : 0;

  return (
    <div className={cn("space-y-2", className)}>
      {label && <Label>{label}</Label>}
      {dims === null ? (
        <p className="text-xs text-gray-500">
          Satuan ini dihitung dari jumlah langsung, tanpa dimensi P/L/T.
        </p>
      ) : (
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
          {dims !== "p" && (
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
          )}
          {dims === "plt" && (
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
      )}
      {volume > 0 && (
        <p className="text-xs text-gray-500">
          Volume: <span className="font-semibold text-gray-700">{volume.toFixed(2)}</span> {unit}
        </p>
      )}
    </div>
  );
}
