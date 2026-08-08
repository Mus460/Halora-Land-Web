"use client";

import { Menu } from "lucide-react";
import { useSidebarStore } from "@/stores/use-sidebar-store";
import { Button } from "@/components/ui/button";
import { UserMenu } from "./user-menu";
import { ProjectSelector } from "@/components/shared/project-selector";

export function Header() {
  const { toggle } = useSidebarStore();

  return (
    <header className="sticky top-0 z-30 h-16 bg-white border-b border-gray-200 flex items-center justify-between px-4 lg:px-6">
      {/* Left */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={toggle}
          className="text-gray-500 hover:text-gray-700 lg:hidden"
        >
          <Menu className="w-5 h-5" />
        </Button>

        {/* Project Selector */}
        <ProjectSelector />
      </div>

      {/* Right */}
      <div className="flex items-center gap-2">
        <UserMenu />
      </div>
    </header>
  );
}
