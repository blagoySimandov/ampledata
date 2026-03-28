import { Mail } from "lucide-react";

export function ContactHeader() {
  return (
    <div className="text-center mb-16 contact-fade-up">
      <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-4 py-1.5 text-sm font-medium mb-6">
        <Mail className="size-4" />
        <span>Get in touch</span>
      </div>
      <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
        We'd love to hear from you
      </h2>
      <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
        Have a question about AmpleData? Want to see a demo or discuss enterprise
        options? Reach out — we'll get back to you within 24 hours.
      </p>
    </div>
  );
}
