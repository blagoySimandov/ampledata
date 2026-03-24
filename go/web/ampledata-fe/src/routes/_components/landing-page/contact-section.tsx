import { useState } from "react";
import {
  AlertCircle,
  CheckCircle,
  Loader2,
  Mail,
  Send,
  Zap,
} from "lucide-react";
import { Button } from "@/components/ui/button";

/**
 * Formspree sends emails to ampledata.io
 */
const FORMSPREE_ENDPOINT = "https://formspree.io/f/xeerypze";

const STATS = [
  { label: "Avg. response time", value: "< 2 hrs" },
  { label: "Satisfaction rate", value: "98.7%" },
  { label: "Enterprise clients", value: "50+" },
] as const;

interface FormState {
  name: string;
  email: string;
  subject: string;
  message: string;
}

const EMPTY_FORM: FormState = { name: "", email: "", subject: "", message: "" };

type SubmitStatus = "idle" | "loading" | "success" | "error";

export function ContactSection() {
  const [formState, setFormState] = useState<FormState>(EMPTY_FORM);
  const [status, setStatus] = useState<SubmitStatus>("idle");
  const [errorMessage, setErrorMessage] = useState<string>("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("loading");
    setErrorMessage("");

    try {
      const res = await fetch(FORMSPREE_ENDPOINT, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(formState),
      });

      if (res.ok) {
        setStatus("success");
        setFormState(EMPTY_FORM);
      } else {
        const data = (await res.json().catch(() => ({}))) as { error?: string };
        setErrorMessage(
          data.error ?? "Something went wrong. Please try again.",
        );
        setStatus("error");
      }
    } catch {
      setErrorMessage(
        "Network error. Please check your connection and try again.",
      );
      setStatus("error");
    }
  }

  function field(key: keyof FormState) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setFormState((s) => ({ ...s, [key]: e.target.value }));
  }

  return (
    <section
      id="contact"
      className="relative bg-background py-24 overflow-hidden"
    >
      {/* Animated decorative blobs */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div className="contact-blob absolute -top-32 -right-32 w-[520px] h-[520px] rounded-full bg-primary/8 blur-3xl" />
        <div className="contact-blob-slow absolute -bottom-24 -left-24 w-[380px] h-[380px] rounded-full bg-primary/5 blur-3xl" />
        <div className="contact-blob absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[300px] h-[300px] rounded-full bg-blue-500/5 blur-3xl" />
      </div>

      <div className="container mx-auto px-4 relative">
        {/* Section header */}
        <div className="text-center mb-16 contact-fade-up">
          <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-4 py-1.5 text-sm font-medium mb-6">
            <Mail className="size-4" />
            <span>Get in touch</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
            We'd love to hear from you
          </h2>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            Have a question about AmpleData? Want to see a demo or discuss
            enterprise options? Reach out — we'll get back to you within 24
            hours.
          </p>
        </div>

        <div className="grid lg:grid-cols-2 gap-12 items-stretch">
          {/* Left column: animated visual panel */}
          <div className="relative rounded-2xl overflow-hidden border border-border shadow-xl bg-gradient-to-br from-primary/20 via-card to-primary/5 min-h-[480px] flex flex-col justify-center contact-fade-up contact-fade-up-delay-1">
            {/* Dot-grid background */}
            <div
              className="absolute inset-0 opacity-40"
              style={{
                backgroundImage:
                  "radial-gradient(circle, oklch(var(--primary-raw, 0.553 0.195 38.402) / 0.25) 1px, transparent 1px)",
                backgroundSize: "26px 26px",
              }}
            />

            {/* Animated orbs */}
            <div className="contact-blob absolute -top-16 -right-16 w-64 h-64 rounded-full bg-primary/20 blur-3xl" />
            <div className="contact-blob-slow absolute -bottom-12 -left-12 w-52 h-52 rounded-full bg-blue-500/15 blur-3xl" />
            <div
              className="contact-blob absolute top-1/3 right-1/4 w-28 h-28 rounded-full bg-violet-500/10 blur-2xl"
              style={{ animationDelay: "3s" }}
            />

            {/* Center content */}
            <div className="relative z-10 flex flex-col items-center text-center px-10 py-14 gap-10">
              {/* Pulsing icon */}
              <div className="relative flex items-center justify-center w-24 h-24">
                <span className="contact-pulse-ring absolute inset-0 rounded-full bg-primary/30" />
                <span
                  className="contact-pulse-ring absolute inset-0 rounded-full bg-primary/20"
                  style={{ animationDelay: "0.8s" }}
                />
                <div className="relative w-20 h-20 rounded-full bg-primary/15 border border-primary/30 flex items-center justify-center shadow-lg backdrop-blur-sm">
                  <Mail className="size-9 text-primary" />
                </div>
              </div>

              {/* Headline */}
              <div>
                <h3 className="text-2xl font-black text-foreground mb-2 leading-snug">
                  We're just a message away
                </h3>
                <p className="text-muted-foreground text-sm max-w-xs leading-relaxed">
                  Quick questions, enterprise demos, or just saying hi — we read
                  every message personally.
                </p>
              </div>

              {/* Floating stat chips */}
              <div className="flex flex-col gap-3 w-full max-w-xs">
                {STATS.map(({ label, value }, i) => (
                  <div
                    key={label}
                    className="contact-blob flex items-center justify-between px-4 py-3 rounded-xl bg-background/60 border border-border backdrop-blur-sm shadow-sm"
                    style={{ animationDelay: `${i * 1.4}s` }}
                  >
                    <span className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Zap className="size-3 text-primary shrink-0" />
                      {label}
                    </span>
                    <span className="text-sm font-bold text-primary">
                      {value}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Right column: contact form */}
          <div className="bg-card rounded-2xl border border-border shadow-xl p-8 contact-fade-up contact-fade-up-delay-2">
            {status === "success" ? (
              <div className="flex flex-col items-center justify-center text-center py-16 gap-5">
                <div className="w-20 h-20 rounded-full bg-emerald-500/10 flex items-center justify-center shadow-lg">
                  <CheckCircle className="size-10 text-emerald-600 dark:text-emerald-400" />
                </div>
                <div>
                  <h3 className="text-2xl font-black text-foreground mb-2">
                    Message sent! 🎉
                  </h3>
                  <p className="text-muted-foreground">
                    Thanks for reaching out. We'll get back to you within 24
                    hours.
                  </p>
                </div>
                <Button variant="outline" onClick={() => setStatus("idle")}>
                  Send another message
                </Button>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-5">
                <div>
                  <h3 className="text-xl font-black text-foreground mb-1">
                    Send us a message
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Fill out the form and we'll be in touch soon.
                  </p>
                </div>

                <div className="grid sm:grid-cols-2 gap-4">
                  <FormField
                    id="contact-name"
                    label="Name"
                    type="text"
                    placeholder="Jane Smith"
                    value={formState.name}
                    onChange={field("name")}
                    required
                  />
                  <FormField
                    id="contact-email"
                    label="Email"
                    type="email"
                    placeholder="jane@company.com"
                    value={formState.email}
                    onChange={field("email")}
                    required
                  />
                </div>

                <FormField
                  id="contact-subject"
                  label="Subject"
                  type="text"
                  placeholder="How can we help?"
                  value={formState.subject}
                  onChange={field("subject")}
                />

                <div className="space-y-1.5">
                  <label
                    htmlFor="contact-message"
                    className="text-sm font-medium text-foreground"
                  >
                    Message
                  </label>
                  <textarea
                    id="contact-message"
                    rows={5}
                    placeholder="Tell us about your project, use case, or question…"
                    value={formState.message}
                    onChange={field("message")}
                    required
                    className="w-full px-3.5 py-2.5 rounded-lg border border-border bg-background text-foreground placeholder:text-muted-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary transition-all resize-none"
                  />
                </div>

                {status === "error" && (
                  <div className="flex items-start gap-3 px-4 py-3 rounded-xl bg-destructive/10 border border-destructive/20 text-sm text-destructive">
                    <AlertCircle className="size-4 mt-0.5 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                <Button
                  type="submit"
                  size="lg"
                  className="w-full text-base"
                  disabled={status === "loading"}
                >
                  {status === "loading" ? (
                    <>
                      <Loader2 className="mr-2 size-4 animate-spin" />
                      Sending…
                    </>
                  ) : (
                    <>
                      Send message <Send className="ml-2 size-4" />
                    </>
                  )}
                </Button>

                <p className="text-xs text-center text-muted-foreground">
                  We respect your privacy. Your information is never shared.
                </p>
              </form>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}

interface FormFieldProps {
  id: string;
  label: string;
  type: string;
  placeholder: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  required?: boolean;
}

function FormField({
  id,
  label,
  type,
  placeholder,
  value,
  onChange,
  required,
}: FormFieldProps) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium text-foreground">
        {label}
      </label>
      <input
        id={id}
        type={type}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        required={required}
        className="w-full px-3.5 py-2.5 rounded-lg border border-border bg-background text-foreground placeholder:text-muted-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary transition-all"
      />
    </div>
  );
}
