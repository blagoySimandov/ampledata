import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ArrowRight, Sparkles } from "lucide-react";
import { HeroStats } from "./hero-stats";

export function Hero() {
	return (
		<section className="relative overflow-hidden">
			<div className="absolute inset-0 bg-gradient-to-b from-primary/5 to-transparent" />
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-24 lg:pt-32 lg:pb-40 relative">
				<div className="text-center max-w-4xl mx-auto">
					<Badge className="mb-4" variant="outline">
						<Sparkles className="h-3 w-3 mr-1" />
						Transform Your Data
					</Badge>
					<h1 className="text-4xl sm:text-5xl lg:text-7xl font-bold tracking-tight text-balance mb-6">
						Enrich your data.{" "}
						<span className="text-primary">
							Empower your decisions.
						</span>
					</h1>
					<p className="text-lg sm:text-xl text-muted-foreground text-pretty mb-8 max-w-2xl mx-auto">
						Transform raw data into actionable insights with our
						powerful enrichment platform. Upload, enrich, and export
						with confidence.
					</p>
					<div className="flex flex-col sm:flex-row gap-4 justify-center">
						<Button size="lg" asChild className="text-base">
							<Link href="/enrich">
								Start Enriching Free
								<ArrowRight className="ml-2 h-4 w-4" />
							</Link>
						</Button>
						<Button
							size="lg"
							variant="outline"
							asChild
							className="text-base bg-transparent"
						>
							<Link href="#features">View Demo</Link>
						</Button>
					</div>
				</div>

				<HeroStats />
			</div>
		</section>
	);
}
