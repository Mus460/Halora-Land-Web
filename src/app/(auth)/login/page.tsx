"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Calculator, Mail } from "lucide-react";
import toast from "react-hot-toast";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

export default function LoginPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showResendButton, setShowResendButton] = useState(false);
  const [isResending, setIsResending] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    
    // Show success message for verified email
    if (params.get('verified') === 'true') {
      toast.success('Email berhasil diverifikasi! Silakan login.');
    }
    
    // Show success message for new registration
    if (params.get('registered') === 'true') {
      toast.success('Akun berhasil dibuat! Cek email untuk verifikasi.');
    }
    
    // Show error if any
    const error = params.get('error');
    if (error) {
      toast.error(decodeURIComponent(error));
    }
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        const errorMsg = data.error || "Login gagal";
        
        // Detect email not confirmed error
        if (errorMsg.includes('Email not confirmed')) {
          toast.error('Email belum diverifikasi. Cek inbox Anda.');
          setShowResendButton(true);
        } else {
          toast.error(errorMsg);
        }
        return;
      }

      toast.success("Login berhasil!");
      router.push("/dashboard");
      router.refresh();
    } catch (error) {
      toast.error("Terjadi kesalahan. Silakan coba lagi.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendVerification = async () => {
    if (!email) {
      toast.error('Masukkan email terlebih dahulu');
      return;
    }

    setIsResending(true);

    try {
      const response = await fetch('/api/auth/resend-verification', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      const data = await response.json();

      if (!response.ok) {
        toast.error(data.error || 'Gagal mengirim email');
        return;
      }

      toast.success('Email verifikasi telah dikirim. Cek inbox Anda.');
      setShowResendButton(false);
    } catch (error) {
      toast.error('Terjadi kesalahan. Silakan coba lagi.');
    } finally {
      setIsResending(false);
    }
  };

  const handleDemoLogin = async () => {
    setEmail("demo@haloraland.id");
    setPassword("password123");
    setIsLoading(true);

    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: "demo@haloraland.id", password: "password123" }),
      });

      const data = await response.json();

      if (!response.ok) {
        toast.error(data.error || "Login gagal");
        return;
      }

      toast.success("Login sebagai demo berhasil!");
      router.push("/dashboard");
      router.refresh();
    } catch (error) {
      toast.error("Terjadi kesalahan. Silakan coba lagi.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="w-14 h-14 bg-amber-500 rounded-xl flex items-center justify-center">
              <Calculator className="w-8 h-8 text-white" />
            </div>
          </div>
          <CardTitle className="text-2xl font-bold">Halora Land</CardTitle>
          <CardDescription>
            Masuk ke akun Anda untuk melanjutkan
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleLogin}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="nama@email.com"
                className="h-11"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <Link
                  href="#"
                  className="text-xs text-amber-600 hover:text-amber-700"
                >
                  Lupa password?
                </Link>
              </div>
              <Input
                id="password"
                type="password"
                placeholder="Masukkan password"
                className="h-11"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
            <Button 
              type="submit"
              className="w-full h-11 bg-amber-500 hover:bg-amber-600 text-white font-semibold"
              disabled={isLoading}
            >
              {isLoading ? "Memproses..." : "Masuk"}
            </Button>
            {showResendButton && (
              <Button
                type="button"
                variant="outline"
                className="w-full h-11 border-amber-500 text-amber-600 hover:bg-amber-50"
                onClick={handleResendVerification}
                disabled={isResending}
              >
                <Mail className="w-4 h-4 mr-2" />
                {isResending ? "Mengirim..." : "Kirim Ulang Email Verifikasi"}
              </Button>
            )}
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <Separator />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-white px-2 text-gray-500">Atau</span>
              </div>
            </div>
            <Button 
              type="button"
              variant="outline" 
              className="w-full h-11"
              onClick={handleDemoLogin}
              disabled={isLoading}
            >
              Masuk sebagai Demo
            </Button>
          </CardContent>
        </form>
        <CardFooter className="justify-center">
          <p className="text-sm text-gray-600">
            Belum punya akun?{" "}
            <Link
              href="/register"
              className="text-amber-600 hover:text-amber-700 font-medium"
            >
              Daftar sekarang
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}
