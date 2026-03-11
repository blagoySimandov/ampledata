import { Card, CardContent } from "@/components/ui/card";
import { RefreshCw } from "lucide-react";

export function LoadingState() {
  return (
    <div className="flex items-center justify-center py-20 text-gray-500 gap-3 font-medium">
      <RefreshCw className="w-5 h-5 animate-spin" /> Loading source...
    </div>
  );
}

export function ErrorState() {
  return (
    <Card className="border-red-200 bg-red-50">
      <CardContent className="pt-6 text-red-700 font-medium">
        Failed to load source.
      </CardContent>
    </Card>
  );
}
