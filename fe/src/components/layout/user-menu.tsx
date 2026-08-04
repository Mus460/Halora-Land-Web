"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { LogOut, Settings, User } from "lucide-react";
import toast from "react-hot-toast";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useProjectStore } from "@/stores/use-project-store";

interface UserInfo {
  id: number;
  namaLengkap: string;
  email: string;
  role: string;
  accountType?: string;
  isDemo?: boolean;
}

const ROLE_LABELS: Record<string, string> = {
  ADMIN: "Admin",
  OWNER: "Pemilik",
  USER: "User",
  DEMO: "Demo",
};

function initialsOf(name: string): string {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0].toUpperCase())
    .join("");
}

export function UserMenu() {
  const router = useRouter();
  const setActiveProject = useProjectStore((s) => s.setActiveProject);
  const [user, setUser] = useState<UserInfo | null>(null);

  useEffect(() => {
    fetchUser();
  }, []);

  const fetchUser = async () => {
    try {
      const response = await fetch("/api/auth/me");
      if (!response.ok) return;
      const result = await response.json();
      setUser(result.user);
    } catch (error) {
      console.error("Fetch user error:", error);
    }
  };

  const handleLogout = async () => {
    try {
      const response = await fetch("/api/auth/logout", { method: "POST" });
      if (!response.ok) {
        const data = await response.json().catch(() => null);
        throw new Error(data?.error || "Logout gagal");
      }
      localStorage.removeItem("currentProyekId");
      setActiveProject(null);
      toast.success("Logout berhasil");
      router.push("/login");
      router.refresh();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Terjadi kesalahan");
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="flex items-center gap-2 hover:opacity-80 transition-opacity">
        <Avatar className="w-8 h-8">
          <AvatarFallback className="bg-amber-500 text-white text-sm font-bold">
            {user ? initialsOf(user.namaLengkap) : "U"}
          </AvatarFallback>
        </Avatar>
        <div className="hidden md:block text-left">
          <p className="text-sm font-medium text-gray-700">
            {user?.namaLengkap || "Memuat..."}
          </p>
          <p className="text-xs text-gray-500">
            {user ? ROLE_LABELS[user.role] || user.role : "..."}
          </p>
        </div>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuGroup>
          <DropdownMenuLabel>Akun Saya</DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => router.push("/profile")}>
          <User className="w-4 h-4 mr-2" />
          Profil
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => router.push("/profile")}>
          <Settings className="w-4 h-4 mr-2" />
          Pengaturan
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="text-red-600" onClick={handleLogout}>
          <LogOut className="w-4 h-4 mr-2" />
          Keluar
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
