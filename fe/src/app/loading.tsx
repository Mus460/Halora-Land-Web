import { LoadingSpinner } from "@/components/shared/loading-spinner";

export default function RootLoading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <LoadingSpinner size="lg" />
    </div>
  );
}
