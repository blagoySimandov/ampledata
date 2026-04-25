import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/terms")({
  component: TermsPage,
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

function TermsPage() {
  return (
    <div className="max-w-2xl mx-auto px-4 py-16 space-y-10">
      <header className="space-y-2">
        <h1 className="text-3xl font-black text-foreground">
          Terms of Service
        </h1>
        <p className="text-sm text-muted-foreground">
          Last updated: April 2025
        </p>
      </header>

      <Section title="Acceptance">
        <p>
          By using AmpleData you agree to these terms. If you don't agree, do
          not use the service.
        </p>
      </Section>

      <Section title="What AmpleData does">
        <p>
          AmpleData enriches datasets by searching the web and extracting
          structured data using AI. Results are best-effort — accuracy is not
          guaranteed. Always verify critical data before relying on it.
        </p>
      </Section>

      <Section title="Your responsibilities">
        <ul className="list-disc pl-5 space-y-1">
          <li>Only upload data you have the right to process.</li>
          <li>Do not use the service for illegal or harmful purposes.</li>
          <li>
            Do not attempt to reverse-engineer, scrape, or abuse the platform.
          </li>
        </ul>
      </Section>

      <Section title="Data ownership">
        <p>
          You own your uploaded data and results. By uploading, you grant
          AmpleData a limited license to process your data solely to provide the
          service.
        </p>
      </Section>

      <Section title="Limitation of liability">
        <p>
          AmpleData is provided as-is. We are not liable for inaccurate
          enrichment results, data loss, or indirect damages arising from use of
          the service.
        </p>
      </Section>

      <Section title="Termination">
        <p>
          We may suspend accounts that violate these terms. You may delete your
          account at any time.
        </p>
      </Section>

      <Section title="Changes">
        <p>
          We may update these terms. Continued use after changes constitutes
          acceptance.
        </p>
      </Section>

      <Section title="Contact">
        <p>Questions? Reach out via the app.</p>
      </Section>
    </div>
  );
}
