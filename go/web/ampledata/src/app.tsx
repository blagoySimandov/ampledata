import { RouterProvider } from "react-router-dom";
import { AuthKitProvider } from "@workos-inc/authkit-react";
import { router } from "@/router";

export function App() {
  return (
    <AuthKitProvider
      clientId={import.meta.env.VITE_WORKOS_CLIENT_ID}
      devMode={true}
    >
      <RouterProvider router={router} />
    </AuthKitProvider>
  );
}

export default App;
