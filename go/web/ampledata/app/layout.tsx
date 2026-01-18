import type React from "react";
import type { Metadata } from "next";
import { Geist } from "next/font/google";
import { Analytics } from "@vercel/analytics/next";
import { Providers } from "./providers";
import "./globals.css";
import { Toaster } from "@/components/ui/toaster";

const geist = Geist({ subsets: ["latin"] });

export const metadata: Metadata = {
	title: "Data Enrichment | Transform Your Data",
	description:
		"Upload, enrich, and export your data with powerful enrichment tools",
	generator: "v0.app",
	icons: {
		icon: [
			{
				url: "/icon-light-32x32.png",
				media: "(prefers-color-scheme: light)",
			},
			{
				url: "/icon-dark-32x32.png",
				media: "(prefers-color-scheme: dark)",
			},
			{
				url: "/icon.svg",
				type: "image/svg+xml",
			},
		],
		apple: "/apple-icon.png",
	},
};

export default function RootLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={`${geist.className} antialiased`}>
				<Toaster position="top-right" />
				<Providers>{children}</Providers>
				<Analytics />
			</body>
		</html>
	);
}
