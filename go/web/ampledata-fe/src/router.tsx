import { createRouter, RouterProvider } from "@tanstack/react-router";
import { useAuth } from "@workos-inc/authkit-react";
import { useEffect } from "react";
import { routeTree } from "./routeTree.gen";

const router = createRouter({
  routeTree,
  context: { auth: undefined! },
});

export function AppRouter() {
  const auth = useAuth();

  useEffect(() => {
    router.invalidate();
  }, [auth.user, auth.isLoading]);

  return <RouterProvider router={router} context={{ auth }} />;
}
