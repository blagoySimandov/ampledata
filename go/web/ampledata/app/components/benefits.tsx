import { Card, CardContent } from "@/components/ui/card";
import { CheckCircle2 } from "lucide-react";

const benefits = [
	"Reduce data preparation time by 90%",
	"Improve decision-making with richer datasets",
	"Scale enrichment operations effortlessly",
	"Maintain complete control over your data",
];

const stats = [
	{
		value: "300%",
		label: "Increase in data completeness",
	},
	{
		value: "20 hours",
		label: "Saved per week on average",
	},
	{
		value: "99.9%",
		label: "Data accuracy rate",
	},
];

export function Benefits() {
	return (
		<section id="benefits" className="py-20 lg:py-32 bg-muted/50">
			<div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div className="grid grid-cols-1 lg:grid-cols-2 gap-12 lg:gap-16 items-center">
					<div>
						<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold mb-6 text-balance">
							Turn data chaos into clarity
						</h2>
						<p className="text-lg text-muted-foreground mb-8 text-pretty">
							Our platform empowers teams to transform incomplete
							datasets into comprehensive, actionable information.
							Stop wasting time on manual data entry and let
							automation do the heavy lifting.
						</p>
						<div className="space-y-4">
							{benefits.map((benefit, index) => (
								<div
									key={index}
									className="flex items-start gap-3"
								>
									<CheckCircle2 className="h-6 w-6 text-primary shrink-0 mt-0.5" />
									<span className="text-lg">{benefit}</span>
								</div>
							))}
						</div>
					</div>
					<div className="space-y-4">
						{stats.map((stat, index) => (
							<Card key={index} className="border-2">
								<CardContent className="p-6">
									<div className="text-5xl font-bold text-primary mb-2">
										{stat.value}
									</div>
									<div className="text-muted-foreground">
										{stat.label}
									</div>
								</CardContent>
							</Card>
						))}
					</div>
				</div>
			</div>
		</section>
	);
}
