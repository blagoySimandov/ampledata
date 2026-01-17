import { Card, CardContent } from "@/components/ui/card";
import {
	Database,
	Zap,
	Shield,
	TrendingUp,
	CheckCircle2,
	ArrowRight,
} from "lucide-react";

const features = [
	{
		icon: Database,
		title: "Universal Import",
		description:
			"Upload CSV, JSON, and other common data formats with drag-and-drop simplicity.",
	},
	{
		icon: Zap,
		title: "Lightning Fast",
		description:
			"Process thousands of rows in seconds with our optimized enrichment engine.",
	},
	{
		icon: Shield,
		title: "Secure by Default",
		description:
			"Enterprise-grade security ensures your data is always protected and private.",
	},
	{
		icon: TrendingUp,
		title: "Real-time Progress",
		description:
			"Watch your data transform with live progress tracking and detailed insights.",
	},
	{
		icon: CheckCircle2,
		title: "Editable Grid",
		description:
			"Edit cells, columns, and rows on the fly with our intuitive data grid interface.",
	},
	{
		icon: ArrowRight,
		title: "Seamless Export",
		description:
			"Export enriched data instantly in your preferred format with a single click.",
	},
];

export function Features() {
	return (
		<section id="features" className="py-20 lg:py-32">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="text-center max-w-3xl mx-auto mb-16">
					<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold mb-4 text-balance">
						Everything you need to enrich your data
					</h2>
					<p className="text-lg text-muted-foreground text-pretty">
						Powerful features designed to make data enrichment fast,
						accurate, and effortless.
					</p>
				</div>

				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
					{features.map((feature, index) => (
						<Card
							key={index}
							className="border-2 hover:border-primary/50 transition-colors"
						>
							<CardContent className="p-6">
								<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
									<feature.icon className="h-6 w-6 text-primary" />
								</div>
								<h3 className="text-xl font-semibold mb-2">
									{feature.title}
								</h3>
								<p className="text-muted-foreground text-pretty">
									{feature.description}
								</p>
							</CardContent>
						</Card>
					))}
				</div>
			</div>
		</section>
	);
}
