"use client";

import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";

interface CurrencyInputProps {
  value: number;
  onChange: (value: number) => void;
  label?: string;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

export function CurrencyInput({
  value,
  onChange,
  label,
  placeholder = "0",
  disabled,
  className,
}: CurrencyInputProps) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.replace(/[^0-9]/g, "");
    onChange(Number(raw) || 0);
  };

  const formatted = value > 0 ? value.toLocaleString("id-ID") : "";

  return (
    <div className={cn("space-y-2", className)}>
      {label && <Label>{label}</Label>}
      <div className="relative">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-500">
          Rp
        </span>
        <Input
          value={formatted}
          onChange={handleChange}
          placeholder={placeholder}
          disabled={disabled}
          className="pl-10"
          inputMode="numeric"
        />
      </div>
    </div>
  );
}
