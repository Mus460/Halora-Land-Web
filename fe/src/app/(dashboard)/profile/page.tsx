"use client";

import { useState, useEffect } from "react";
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
    fullName: "",
    email: "",
  });
  const [loading, setLoading] = useState(true);

  const [password, setPassword] = useState({
    current: "",
    new: "",
    confirm: "",
  });

  useEffect(() => {
    fetchProfile();
  }, []);

  const fetchProfile = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/auth/me');
      if (!response.ok) throw new Error('Failed to fetch');
      const result = await response.json();
      setProfile({
        fullName: result.user.fullName || "",
        email: result.user.email || "",
      });
    } catch (error) {
      console.error('Fetch error:', error);
      toast.error('Gagal memuat profil');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveProfile = async () => {
    try {
      const response = await fetch('/api/auth/me', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fullName: profile.fullName,
          email: profile.email,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Update failed');
      }

      toast.success("Profil berhasil disimpan");
    } catch (error) {
      console.error('Update error:', error);
      toast.error(error instanceof Error ? error.message : 'Gagal update profil');
    }
  };

  const handleChangePassword = async () => {
    if (password.new !== password.confirm) {
      toast.error("Password baru tidak cocok");
      return;
    }

    if (password.new.length < 6) {
      toast.error("Password minimal 6 karakter");
      return;
    }

    try {
      const response = await fetch('/api/auth/update-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password: password.new }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Gagal mengubah password');
      }

      toast.success("Password berhasil diubah");
      setPassword({ current: "", new: "", confirm: "" });
    } catch (error: any) {
      toast.error(error.message || "Gagal mengubah password");
    }
  };

  if (loading) {
    return <div className="p-8 text-center">Memuat data...</div>;
  }

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
              Informasi Profil
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Nama Lengkap</Label>
              <Input
                value={profile.fullName}
                onChange={(e) =>
                  setProfile({ ...profile, fullName: e.target.value })
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
                placeholder="Minimal 6 karakter"
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
