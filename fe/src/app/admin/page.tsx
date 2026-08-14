import { Users } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { PageHeader } from "@/components/shared/page-header";
import Link from "next/link";

const adminMenus = [
  {
    href: "/admin/users",
    title: "Kelola User",
    description: "Manajemen pengguna dan role",
    icon: Users,
  },
];

export default function AdminPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Admin Center"
        description="Panel administrasi sistem"
      />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {adminMenus.map((menu) => (
          <Link key={menu.href} href={menu.href}>
            <Card className="hover:shadow-md transition-shadow cursor-pointer h-full">
              <CardContent className="p-6 text-center">
                <div className="w-14 h-14 bg-amber-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <menu.icon className="w-7 h-7 text-amber-600" />
                </div>
                <h3 className="font-semibold text-gray-900 mb-1">
                  {menu.title}
                </h3>
                <p className="text-sm text-gray-500">{menu.description}</p>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
