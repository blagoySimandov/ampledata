import { useUser } from "@/hooks";

export function Enrichment() {
  const user = useUser();

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <p className="text-lg">Loading...</p>
      </div>
    );
  }

  return (
    <div className="py-12">
      <h1 className="text-3xl font-bold mb-4">
        Welcome back{user.firstName && `, ${user.firstName}`}
      </h1>
      <p className="text-lg text-muted-foreground">
        This is the enrichment page. Only authenticated users can access this page.
      </p>
    </div>
  );
}
