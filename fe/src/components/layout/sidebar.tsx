"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Building2,
  Calculator,
  FileText,
  Package,
  TrendingUp,
  Wallet,
  ClipboardCheck,
  Tag,
  LibraryBig,
  Ruler,
  Box,
  Pickaxe,
  Umbrella,
  Construction,
  MoveUpRight,
  House,
  Grid3X3,
  Layers,
  Droplet,
  LayoutGrid,
  Map,
  Paintbrush,
  DoorOpen,
  Sofa,
  Bath,
  Zap,
  PenTool,
  UserRound,
  Settings,
  Shield,
  X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useSidebarStore } from "@/stores/use-sidebar-store";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { SidebarItem } from "./sidebar-item";
import { SidebarSection } from "./sidebar-section";

const SIDEBAR_SECTIONS = [
  {
    id: "utama",
    title: "Utama",
    items: [
      { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
      { href: "/projects", label: "Data Project", icon: Building2 },
      { href: "/rab", label: "RAB", icon: Calculator },
    ],
  },
  {
    id: "analisa",
    title: "Analisa & Laporan",
    items: [
      { href: "/recaps", label: "Rekapitulasi", icon: Calculator },
      { href: "/invoices", label: "Invoice", icon: FileText },
      { href: "/logistics", label: "Logistik", icon: Package },
      { href: "/s-curve", label: "Kurva S", icon: TrendingUp },
      { href: "/transactions", label: "Keuangan", icon: Wallet },
      { href: "/monitoring", label: "Progress", icon: ClipboardCheck },
    ],
  },
  {
    id: "master",
    title: "Master Data",
    items: [
      { href: "/price-masters", label: "Master Harga", icon: Tag },
      { href: "/analysis-masters", label: "Master Analisa", icon: LibraryBig },
      { href: "/clients", label: "Master Klien", icon: UserRound },
    ],
  },
  {
    id: "konstruksi",
    title: "Pekerjaan Konstruksi",
    hideOnMobile: true,
    items: [
      { href: "/preparation", label: "Persiapan", icon: Ruler },
      { href: "/foundation", label: "Pondasi", icon: Box },
      { href: "/concrete", label: "Beton Struktur", icon: Pickaxe },
      { href: "/canopy", label: "Kanopi", icon: Umbrella },
      { href: "/steel", label: "Baja Struktural", icon: Construction },
      { href: "/stairs", label: "Tangga", icon: MoveUpRight },
      { href: "/roof", label: "Pek. Atap", icon: House },
    ],
  },
  {
    id: "arsitektur",
    title: "Arsitektur & MEP",
    hideOnMobile: true,
    items: [
      { href: "/wall", label: "Pas. Dinding", icon: Grid3X3 },
      { href: "/plastering", label: "Plesteran", icon: Layers },
      { href: "/finishing", label: "Acian", icon: Droplet },
      { href: "/tiles", label: "Lantai/Keramik", icon: LayoutGrid },
      { href: "/paving", label: "Paving & Halaman", icon: Map },
      { href: "/painting", label: "Cat & Plafon", icon: Paintbrush },
      { href: "/doors", label: "Kusen/Pintu", icon: DoorOpen },
      { href: "/interior", label: "Interior", icon: Sofa },
      { href: "/toilet", label: "Toilet/Sanitair", icon: Bath },
      { href: "/mep", label: "Instalasi MEP", icon: Zap },
    ],
  },
  {
    id: "tambahan",
    title: "Pekerjaan Tambahan",
    hideOnMobile: true,
    items: [
      { href: "/custom-work", label: "Pekerjaan Custom", icon: PenTool },
    ],
  },
  {
    id: "pengaturan",
    title: "Pengaturan",
    items: [
      { href: "/profile", label: "Pengaturan Usaha", icon: Settings },
      { href: "/admin", label: "Admin Center", icon: Shield, badge: 1 },
    ],
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const { isOpen, setOpen } = useSidebarStore();

  return (
    <>
      {/* Overlay for mobile */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          "fixed left-0 top-0 h-full z-50 w-70 bg-gray-800 border-r border-gray-700 flex flex-col",
          "transition-transform duration-200 ease-in-out",
          "lg:translate-x-0",
          isOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        {/* Logo */}
        <div className="flex items-center justify-between h-16 px-4 border-b border-gray-700">
          <Link href="/dashboard" className="flex items-center gap-2">
            <div className="w-8 h-8 bg-amber-500 rounded-lg flex items-center justify-center">
              <Calculator className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-sm font-bold text-white leading-tight">
                Halora Land
              </h1>
              <p className="text-[10px] text-amber-400 font-medium">
                V1 Edition
              </p>
            </div>
          </Link>
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden text-gray-400 hover:text-white"
            onClick={() => setOpen(false)}
          >
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* AHSP Banner */}
        <div className="mx-3 mt-3 px-3 py-2 bg-amber-500/20 border border-amber-500/30 rounded-md">
          <p className="text-[10px] font-bold text-amber-400 text-center leading-tight">
            DATABASE AHSP 2026
            <br />
            SUDAH AKTIF
          </p>
        </div>

        {/* Navigation */}
        <ScrollArea className="flex-1 min-h-0 custom-scrollbar">
          <nav className="px-2 py-3 space-y-4">
            {SIDEBAR_SECTIONS.map((section) => (
              <SidebarSection
                key={section.id}
                title={section.title}
                defaultOpen={["utama", "analisa"].includes(section.id)}
              >
                {section.items.map((item) => (
                  <SidebarItem
                    key={item.href}
                    href={item.href}
                    label={item.label}
                    icon={item.icon}
                    badge={"badge" in item ? item.badge : undefined}
                  />
                ))}
              </SidebarSection>
            ))}
          </nav>
        </ScrollArea>
      </aside>
    </>
  );
}
