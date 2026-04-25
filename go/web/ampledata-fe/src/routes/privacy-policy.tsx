import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/privacy-policy")({
  component: PrivacyPolicyPage,
});

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="space-y-2">
      <h2 className="text-lg font-semibold text-foreground">{title}</h2>
      <div className="text-muted-foreground leading-relaxed">{children}</div>
    </section>
  );
}

function PrivacyPolicyPage() {
  return (
    <div className="max-w-2xl mx-auto px-4 py-16 space-y-10">
      <header className="space-y-2">
        <h1 className="text-3xl font-black text-foreground">Privacy Policy</h1>
        <p className="text-sm text-muted-foreground">
          Last updated: April 2025
        </p>
      </header>

      <Section title="What we collect">
        <p>
          We collect your email address and name when you sign in. We store the
          files you upload (CSV/JSON) and the enrichment jobs you run, including
          results and source URLs.
        </p>
      </Section>

      <Section title="How we use it">
        <p>
          Your data is used solely to operate the service — running enrichment
          jobs, displaying results, and maintaining your account. We do not sell
          or share your data with third parties, except the APIs required to run
          the service (search, crawling, LLM inference).
        </p>
      </Section>

      <Section title="Data retention">
        <p>
          Your uploaded files and job results are retained until you delete
          them. Account data is deleted within 30 days of account deletion.
        </p>
      </Section>

      <Section title="Third-party services">
        <p>
          AmpleData uses Serper for web search, Crawl4ai for page crawling, and
          an LLM provider for data extraction. Queries sent to these services
          include row content from your uploaded files.
        </p>
      </Section>

      <Section title="Security">
        <p>
          Data is transmitted over HTTPS. We do not store payment information.
          If you discover a security issue, please contact us immediately.
        </p>
      </Section>

      <Section title="Contact">
        <p>Questions? Reach out via the app.</p>
      </Section>
    </div>
  );
}
