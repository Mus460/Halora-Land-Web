"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { LucideIcon } from "lucide-react";

interface SidebarItemProps {
  href: string;
  label: string;
  icon: LucideIcon;
  badge?: number | string;
  collapsed?: boolean;
}

export function SidebarItem({
  href,
  label,
  icon: Icon,
  badge,
  collapsed,
}: SidebarItemProps) {
  const pathname = usePathname();
  const isActive = pathname === href || pathname.startsWith(href + "/");

  const content = (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
        "hover:bg-amber-500/10 hover:text-amber-400",
        isActive
          ? "bg-amber-500/20 text-amber-400"
          : "text-gray-300 hover:text-white"
      )}
    >
      <Icon className="w-5 h-5 shrink-0" />
      {!collapsed && (
        <>
          <span className="truncate">{label}</span>
          {badge !== undefined && (
            <span
              className={cn(
                "ml-auto text-xs px-1.5 py-0.5 rounded-full",
                typeof badge === "number" && badge > 0
                  ? "bg-amber-500 text-white animate-pulse-slow"
                  : "bg-gray-600 text-gray-300"
              )}
            >
              {badge}
            </span>
          )}
        </>
      )}
    </Link>
  );

  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger>{content}</TooltipTrigger>
        <TooltipContent side="right" className="z-[100]">
          {label}
        </TooltipContent>
      </Tooltip>
    );
  }

  return content;
}
