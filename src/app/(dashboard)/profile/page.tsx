"use client";

import { useState } from "react";
import { Settings, Save, Lock } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import toast from "react-hot-toast";

export default function ProfilePage() {
  const [profile, setProfile] = useState({
    namaLengkap: "Budi Kontraktor",
    email: "budi@example.com",
    namaUsaha: "CV. Bangun Jaya",
    alamat: "Jl. Merdeka No. 123, Bandung",
    telepon: "081234567890",
  });

  const [password, setPassword] = useState({
    current: "",
    new: "",
    confirm: "",
  });

  const handleSaveProfile = () => {
    toast.success("Profil berhasil disimpan");
  };

  const handleChangePassword = () => {
    if (password.new !== password.confirm) {
      toast.error("Password baru tidak cocok");
      return;
    }
    toast.success("Password berhasil diubah");
    setPassword({ current: "", new: "", confirm: "" });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Pengaturan Usaha"
        description="Kelola profil dan pengaturan akun Anda"
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Profile */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings className="w-5 h-5" />
              Informasi Usaha
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Nama Lengkap</Label>
              <Input
                value={profile.namaLengkap}
                onChange={(e) =>
                  setProfile({ ...profile, namaLengkap: e.target.value })
                }
              />
            </div>
            <div className="space-y-2">
              <Label>Email</Label>
              <Input
                value={profile.email}
                onChange={(e) =>
                  setProfile({ ...profile, email: e.target.value })
                }
                type="email"
              />
            </div>
            <div className="space-y-2">
              <Label>Nama Usaha</Label>
              <Input
                value={profile.namaUsaha}
                onChange={(e) =>
                  setProfile({ ...profile, namaUsaha: e.target.value })
                }
              />
            </div>
            <div className="space-y-2">
              <Label>Alamat</Label>
              <Input
                value={profile.alamat}
                onChange={(e) =>
                  setProfile({ ...profile, alamat: e.target.value })
                }
              />
            </div>
            <div className="space-y-2">
              <Label>Telepon</Label>
              <Input
                value={profile.telepon}
                onChange={(e) =>
                  setProfile({ ...profile, telepon: e.target.value })
                }
              />
            </div>
            <Button
              onClick={handleSaveProfile}
              className="bg-amber-500 hover:bg-amber-600"
            >
              <Save className="w-4 h-4 mr-2" />
              Simpan Profil
            </Button>
          </CardContent>
        </Card>

        {/* Password */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Lock className="w-5 h-5" />
              Ubah Password
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Password Saat Ini</Label>
              <Input
                type="password"
                value={password.current}
                onChange={(e) =>
                  setPassword({ ...password, current: e.target.value })
                }
                placeholder="Masukkan password saat ini"
              />
            </div>
            <Separator />
            <div className="space-y-2">
              <Label>Password Baru</Label>
              <Input
                type="password"
                value={password.new}
                onChange={(e) =>
                  setPassword({ ...password, new: e.target.value })
                }
                placeholder="Minimal 8 karakter"
              />
            </div>
            <div className="space-y-2">
              <Label>Konfirmasi Password Baru</Label>
              <Input
                type="password"
                value={password.confirm}
                onChange={(e) =>
                  setPassword({ ...password, confirm: e.target.value })
                }
                placeholder="Ulangi password baru"
              />
            </div>
            <Button onClick={handleChangePassword} variant="outline">
              <Lock className="w-4 h-4 mr-2" />
              Ubah Password
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
