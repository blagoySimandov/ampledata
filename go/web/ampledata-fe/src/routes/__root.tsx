import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/router-devtools";
import { Header } from "../components/layout";

export const Route = createRootRoute({
  component: () => (
    <div className="min-h-screen bg-gray-50 flex flex-col font-sans">
      <Header />
      <main className="flex-1 container mx-auto p-4 py-8">
        <Outlet />
      </main>
      <TanStackRouterDevtools />
    </div>
  ),
});
