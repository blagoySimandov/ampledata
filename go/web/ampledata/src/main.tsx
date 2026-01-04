import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AuthKitProvider } from "@workos-inc/authkit-react";

import "./index.css";
import App from "./app";
import { WORKOS_CLIENT_ID } from "./constants";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AuthKitProvider clientId={WORKOS_CLIENT_ID}>
      <App />
    </AuthKitProvider>
  </StrictMode>,
);
