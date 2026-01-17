"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Database } from "lucide-react";
import { useAuth } from "@workos-inc/authkit-react";

export function Navigation() {
	const { user, isLoading, signOut } = useAuth();

	return (
		<nav className="border-b sticky top-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 z-50">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="flex justify-between items-center h-16">
					<div className="flex items-center gap-2">
						<Database className="h-6 w-6 text-primary" />
						<span className="text-xl font-semibold">
							DataEnrich
						</span>
					</div>
					<div className="hidden md:flex items-center gap-8">
						<Link
							href="#features"
							className="text-sm text-muted-foreground hover:text-foreground transition-colors"
						>
							Features
						</Link>
						<Link
							href="#benefits"
							className="text-sm text-muted-foreground hover:text-foreground transition-colors"
						>
							Benefits
						</Link>
						<Link
							href="#pricing"
							className="text-sm text-muted-foreground hover:text-foreground transition-colors"
						>
							Pricing
						</Link>
					</div>
					<div className="flex items-center gap-4">
						{user ? (
							<>
								<Button size="sm" asChild>
									<Link href="/enrich">Enrich</Link>
								</Button>
								<DropdownMenu>
									<DropdownMenuTrigger asChild>
										<Button size="sm" variant="ghost">
											{isLoading
												? "Loading..."
												: user.firstName ||
													user.email ||
													"User"}
										</Button>
									</DropdownMenuTrigger>
									<DropdownMenuContent align="end">
										<DropdownMenuItem
											onClick={() => signOut()}
											variant="destructive"
										>
											Sign Out
										</DropdownMenuItem>
									</DropdownMenuContent>
								</DropdownMenu>
							</>
						) : (
							<Button size="sm" asChild>
								<Link href="/login">Get Started</Link>
							</Button>
						)}
					</div>
				</div>
			</div>
		</nav>
	);
}
