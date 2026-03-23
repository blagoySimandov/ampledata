import { renderToString } from "react-dom/server";
import {
  RouterProvider,
  createMemoryHistory,
  createRouter,
  createRootRoute,
  createRoute,
  Outlet,
} from "@tanstack/react-router";
import { LandingPage } from "@/routes/_components/landing-page";

/*
 *
 * File is only used for SSR for the landing page to better the SEO
 *
 * **/

function buildSSRRouter() {
  const rootRoute = createRootRoute({ component: Outlet });
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: LandingPage,
  });
  const routeTree = rootRoute.addChildren([indexRoute]);
  const memoryHistory = createMemoryHistory({ initialEntries: ["/"] });
  return createRouter({ routeTree, history: memoryHistory });
}

export async function render() {
  const router = buildSSRRouter();
  await router.load();
  return renderToString(<RouterProvider router={router} />);
}
