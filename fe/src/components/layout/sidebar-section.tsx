"use client";

import { useState } from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";

interface SidebarSectionProps {
  title: string;
  defaultOpen?: boolean;
  collapsed?: boolean;
  children: React.ReactNode;
}

export function SidebarSection({
  title,
  defaultOpen = true,
  collapsed,
  children,
}: SidebarSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  if (collapsed) {
    return <div className="space-y-1">{children}</div>;
  }

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger className="flex items-center justify-between w-full px-3 py-2 text-[11px] font-bold text-gray-400 uppercase tracking-wider hover:text-gray-200 transition-colors">
        <span className="truncate">{title}</span>
        <ChevronDown
          className={cn(
            "w-3.5 h-3.5 shrink-0 transition-transform",
            isOpen && "rotate-180"
          )}
        />
      </CollapsibleTrigger>
      <CollapsibleContent className="space-y-0.5 mt-0.5">
        {children}
      </CollapsibleContent>
    </Collapsible>
  );
}
