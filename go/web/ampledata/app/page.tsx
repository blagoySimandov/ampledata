"use client";

import {
	Navigation,
	Hero,
	Features,
	Benefits,
	CTA,
	Footer,
} from "./components";

export default function LandingPage() {
	return (
		<div className="min-h-screen bg-background">
			<Navigation />
			<Hero />
			<Features />
			<Benefits />
			<CTA />
			<Footer />
		</div>
	);
}
