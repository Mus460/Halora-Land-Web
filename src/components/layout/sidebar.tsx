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
  MessageSquare,
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
      { href: "/proyek", label: "Data Proyek", icon: Building2 },
    ],
  },
  {
    id: "analisa",
    title: "Analisa & Laporan",
    items: [
      { href: "/rekap", label: "Rekapitulasi", icon: Calculator },
      { href: "/invoice", label: "Invoice", icon: FileText },
      { href: "/logistik", label: "Logistik", icon: Package },
      { href: "/kurva-s", label: "Kurva S", icon: TrendingUp },
      { href: "/realisasi", label: "Keuangan", icon: Wallet },
      { href: "/monitoring", label: "Progress", icon: ClipboardCheck },
    ],
  },
  {
    id: "master",
    title: "Master Data",
    items: [
      { href: "/master-harga", label: "Master Harga", icon: Tag },
      { href: "/master-analisa", label: "Master Analisa", icon: LibraryBig },
    ],
  },
  {
    id: "konstruksi",
    title: "Pekerjaan Konstruksi",
    hideOnMobile: true,
    items: [
      { href: "/persiapan", label: "Persiapan", icon: Ruler },
      { href: "/pondasi", label: "Pondasi", icon: Box },
      { href: "/beton", label: "Beton Struktur", icon: Pickaxe },
      { href: "/kanopi", label: "Kanopi", icon: Umbrella },
      { href: "/baja", label: "Baja Struktural", icon: Construction },
      { href: "/tangga", label: "Tangga", icon: MoveUpRight },
      { href: "/atap", label: "Pek. Atap", icon: House },
    ],
  },
  {
    id: "arsitektur",
    title: "Arsitektur & MEP",
    hideOnMobile: true,
    items: [
      { href: "/dinding", label: "Pas. Dinding", icon: Grid3X3 },
      { href: "/plesteran", label: "Plesteran", icon: Layers },
      { href: "/acian", label: "Acian", icon: Droplet },
      { href: "/keramik", label: "Lantai/Keramik", icon: LayoutGrid },
      { href: "/paving", label: "Paving & Halaman", icon: Map },
      { href: "/pengecatan", label: "Cat & Plafon", icon: Paintbrush },
      { href: "/pintu", label: "Kusen/Pintu", icon: DoorOpen },
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
      { href: "/pekerjaan-custom", label: "Pekerjaan Custom", icon: PenTool },
    ],
  },
  {
    id: "pengaturan",
    title: "Pengaturan",
    items: [
      { href: "/feedback", label: "Feedback & Support", icon: MessageSquare, badge: 1 },
      { href: "/profile", label: "Pengaturan Usaha", icon: Settings },
      { href: "/admin", label: "Admin Center", icon: Shield, badge: 2 },
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
          "fixed left-0 top-0 h-full z-50 w-70 bg-gray-800 border-r border-gray-700",
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
                V3 Pro Edition
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
        <ScrollArea className="h-[calc(100vh-10rem)] custom-scrollbar">
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

        {/* Footer */}
        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-700">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-amber-500 rounded-full flex items-center justify-center text-white text-sm font-bold">
              B
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">
                Budi Kontraktor
              </p>
              <p className="text-xs text-gray-400 truncate">budi@example.com</p>
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}
