import {
  createRootRouteWithContext,
  Outlet,
  redirect,
  useRouterState,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/router-devtools";
import { useAuth, type User } from "@workos-inc/authkit-react";
import { Header } from "../components/layout";
import { Loader2 } from "lucide-react";

interface RouterContext {
  auth: {
    isLoading: boolean;
    user: User | null;
  };
}

const PUBLIC_ROUTES = ["/login", "/auth/callback", "/", "/privacy-policy", "/terms"];

function isPublicRoute(pathname: string) {
  return PUBLIC_ROUTES.some((r) =>
    r === "/" ? pathname === "/" : pathname.startsWith(r)
  );
}

export const Route = createRootRouteWithContext<RouterContext>()({
  beforeLoad: ({ context, location }) => {
    if (isPublicRoute(location.pathname)) return;
    if (!context.auth.isLoading && !context.auth.user) {
      throw redirect({ to: "/login" });
    }
  },
  component: RootComponent,
});

function LoadingScreen() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Loader2 className="size-8 animate-spin text-primary" />
    </div>
  );
}

function RootComponent() {
  const { user, isLoading } = useAuth();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isPublic = isPublicRoute(pathname);

  if (isLoading && !isPublic) return <LoadingScreen />;

  return (
    <div className="min-h-screen bg-background flex flex-col font-sans">
      {!isPublic && user && <Header />}
      <main
        className={
          !isPublic && user ? "flex-1 container mx-auto p-4 py-8" : "flex-1"
        }
      >
        <Outlet />
      </main>
      {import.meta.env.DEV && <TanStackRouterDevtools />}
    </div>
  );
}
