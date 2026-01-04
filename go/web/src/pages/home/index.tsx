import { Hero1 } from "./components/hero1";

export function Home() {
  return (
    <Hero1
      heading="Transform Sparse Data Into Intelligence"
      description="Ample Data uses advanced AI and web scraping to enrich your lead lists with verified emails, company insights, financial metrics, and actionable intelligence processed in bulk."
      buttons={{
        primary: {
          text: "Start Enriching",
          url: "https://ampleddata.com/api",
        },
        secondary: {
          text: "View API Docs",
          url: "https://ampleddata.com/docs",
        },
      }}
      image={{
        src: "https://images.unsplash.com/photo-1763568258343-cdf1306baa2d?q=80&w=2070&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
        alt: "Ample Data enrichment dashboard showing lead transformation and insights",
      }}
    />
  );
}
