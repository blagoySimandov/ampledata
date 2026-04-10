import { createRouter, RouterProvider } from "@tanstack/react-router";
import { useAuth } from "@workos-inc/authkit-react";
import { useEffect } from "react";
import { logEvent } from "./lib/analytics";
import { routeTree } from "./routeTree.gen";

const router = createRouter({
  routeTree,
  context: { auth: undefined! },
});

router.subscribe("onResolved", () => {
  logEvent("page_view", { page_path: router.state.location.pathname });
});

export function AppRouter() {
  const auth = useAuth();

  useEffect(() => {
    router.invalidate();
  }, [auth.user, auth.isLoading]);

  return <RouterProvider router={router} context={{ auth }} />;
}
