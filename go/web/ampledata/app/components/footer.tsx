import Link from "next/link";
import { Database } from "lucide-react";

const footerLinks = {
	Product: [
		{ href: "#features", label: "Features" },
		{ href: "#pricing", label: "Pricing" },
		{ href: "#", label: "Documentation" },
	],
	Company: [
		{ href: "#", label: "About" },
		{ href: "#", label: "Blog" },
		{ href: "#", label: "Careers" },
	],
	Resources: [
		{ href: "#", label: "Help Center" },
		{ href: "#", label: "API Docs" },
		{ href: "#", label: "Status" },
	],
	Legal: [
		{ href: "#", label: "Privacy" },
		{ href: "#", label: "Terms" },
		{ href: "#", label: "Security" },
	],
};

export function Footer() {
	return (
		<footer className="border-t py-12">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
					{Object.entries(footerLinks).map(([category, links]) => (
						<div key={category}>
							<h3 className="font-semibold mb-4">{category}</h3>
							<ul className="space-y-2">
								{links.map((link) => (
									<li key={link.label}>
										<Link
											href={link.href}
											className="text-sm text-muted-foreground hover:text-foreground"
										>
											{link.label}
										</Link>
									</li>
								))}
							</ul>
						</div>
					))}
				</div>
				<div className="border-t pt-8 flex flex-col sm:flex-row justify-between items-center gap-4">
					<div className="flex items-center gap-2">
						<Database className="h-5 w-5 text-primary" />
						<span className="font-semibold">DataEnrich</span>
					</div>
					<p className="text-sm text-muted-foreground">
						© 2026 DataEnrich. All rights reserved.
					</p>
				</div>
			</div>
		</footer>
	);
}
