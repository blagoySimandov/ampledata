import { createFileRoute } from "@tanstack/react-router";
import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";

const ApiDocs = lazy(() => import("./_components/api-docs"));

function DocsFallback() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Loader2 className="size-8 animate-spin text-primary" />
    </div>
  );
}

function DocsPage() {
  return (
    <Suspense fallback={<DocsFallback />}>
      <ApiDocs />
    </Suspense>
  );
}

export const Route = createFileRoute("/docs")({
  component: DocsPage,
});
