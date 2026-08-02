import { cn } from "@/lib/utils";
import { Card, CardContent } from "@/components/ui/card";
import { TrendingUp, TrendingDown } from "lucide-react";

interface StatCardProps {
  title: string;
  value: string | number;
  description?: string;
  icon: React.ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  className?: string;
}

export function StatCard({
  title,
  value,
  description,
  icon,
  trend,
  className,
}: StatCardProps) {
  return (
    <Card className={cn("", className)}>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-500">{title}</p>
            <p className="text-2xl font-bold text-gray-900">{value}</p>
            {trend && (
              <div className="flex items-center gap-1 text-xs">
                {trend.isPositive ? (
                  <TrendingUp className="w-3.5 h-3.5 text-emerald-500" />
                ) : (
                  <TrendingDown className="w-3.5 h-3.5 text-red-500" />
                )}
                <span
                  className={cn(
                    "font-medium",
                    trend.isPositive ? "text-emerald-600" : "text-red-600"
                  )}
                >
                  {trend.value}%
                </span>
                {description && (
                  <span className="text-gray-500">{description}</span>
                )}
              </div>
            )}
            {!trend && description && (
              <p className="text-xs text-gray-500">{description}</p>
            )}
          </div>
          <div className="w-12 h-12 bg-amber-50 rounded-lg flex items-center justify-center text-amber-600">
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
