import { createBrowserRouter } from "react-router-dom";
import { Layout } from "@/components/layout/layout";
import { Home } from "@/pages/home";
import { Enrichment } from "@/pages/enrichment";
import { Login } from "@/pages/login";

export const router = createBrowserRouter([
  {
    path: "/",
    element: (
      <Layout>
        <Home />
      </Layout>
    ),
  },
  {
    path: "/enrichment",
    element: (
      <Layout>
        <Enrichment />
      </Layout>
    ),
  },
  {
    path: "/login",
    element: <Login />,
  },
]);
