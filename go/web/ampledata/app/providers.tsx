"use client";

import type React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthKitProvider } from "@workos-inc/authkit-react";
import { ThemeProvider } from "@/components/theme-provider";

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 1000 * 60 * 5,
			refetchOnWindowFocus: false,
		},
	},
});

export function Providers({ children }: { children: React.ReactNode }) {
	return (
		<QueryClientProvider client={queryClient}>
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
		</QueryClientProvider>
	);
}
