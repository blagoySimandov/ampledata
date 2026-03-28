import { useState } from "react";
import { AlertCircle, CheckCircle, Loader2, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  EMPTY_FORM,
  FORMSPREE_ENDPOINT,
  type FormState,
  type SubmitStatus,
} from "./constants";
import { FormField } from "./form-field";

export function ContactForm() {
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
              Thanks for reaching out. We'll get back to you within 24 hours.
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
  );
}
