import type { Metadata } from "next";
import { Poppins } from "next/font/google";
import { Toaster } from "react-hot-toast";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PostHogProvider } from "@/lib/posthog";
import { Analytics } from "@vercel/analytics/react";
import "./globals.css";

const poppins = Poppins({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-poppins",
});

export const metadata: Metadata = {
  title: "Halora Land - Kalkulasi RAB Konstruksi",
  description:
    "Aplikasi kalkulasi RAB konstruksi berdasarkan AHSP PUPR 2026 untuk kontraktor, mandor, konsultan, dan pemilik proyek.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" className={`${poppins.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col font-sans">
        <PostHogProvider>
          <TooltipProvider>{children}</TooltipProvider>
        </PostHogProvider>
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 3000,
            style: {
              background: "#1f2937",
              color: "#fff",
              borderRadius: "0.5rem",
            },
            success: {
              style: { background: "#10b981" },
            },
            error: {
              style: { background: "#ef4444" },
            },
          }}
        />
        <Analytics />
      </body>
    </html>
  );
}
