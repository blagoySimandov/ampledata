"use client";

import type React from "react";
import { AuthKitProvider } from "@workos-inc/authkit-react";
import { ThemeProvider } from "@/components/theme-provider";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <AuthKitProvider
      clientId={process.env.NEXT_PUBLIC_WORKOS_CLIENT_ID!}
      devMode={true}
    >
      <ThemeProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        storageKey="data-enrichment-theme"
      >
        {children}
      </ThemeProvider>
    </AuthKitProvider>
  );
}
