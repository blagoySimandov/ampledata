import { createBrowserRouter } from "react-router-dom";
import { Layout } from "@/components/layout/layout";
import { Home } from "@/pages/home";
import { Login } from "@/pages/login";
import { ProtectedRoute } from "@/components/auth";
import { AUTH_ROUTES, APP_ROUTES } from "@/constants";

export const router = createBrowserRouter([
  {
    path: APP_ROUTES.HOME,
    element: (
      <Layout>
        <Home />
      </Layout>
    ),
  },
  {
    path: AUTH_ROUTES.LOGIN,
    element: <Login />,
  },
  {
    path: APP_ROUTES.ENRICHMENT,
    element: (
      <ProtectedRoute>
        <Layout>
          <div>Enrichment</div>
        </Layout>
      </ProtectedRoute>
    ),
  },
]);
