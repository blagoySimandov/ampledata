import { useState } from "react";
import { AlertCircle, CheckCircle, Loader2, Mail, Send } from "lucide-react";
import type { UserResponse } from "@/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

const FORMSPREE_ENDPOINT = "https://formspree.io/f/xeerypze";

type SubmitStatus = "idle" | "loading" | "success" | "error";

interface Props {
  variant: "landing" | "account";
  user?: UserResponse;
}

export function ContactFormWidget({ variant, user }: Props) {
  const [name, setName] = useState(
    user ? `${user.first_name} ${user.last_name}`.trim() : "",
  );
  const [email, setEmail] = useState(user?.email ?? "");
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [status, setStatus] = useState<SubmitStatus>("idle");
  const [errorMessage, setErrorMessage] = useState("");

  const isLanding = variant === "landing";

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("loading");
    setErrorMessage("");

    const payload = isLanding
      ? { name, email, subject, message }
      : { name, email, message };

    try {
      const res = await fetch(FORMSPREE_ENDPOINT, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setStatus("success");
        setSubject("");
        setMessage("");
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

  const formBody = (
    <>
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
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <h3 className={isLanding ? "text-xl font-black text-foreground mb-1" : "text-base font-semibold text-foreground mb-1"}>
              Send us a message
            </h3>
            <p className="text-sm text-muted-foreground">
              Fill out the form and we'll be in touch soon.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor={`contact-name-${variant}`}>Name</Label>
              <Input
                id={`contact-name-${variant}`}
                type="text"
                placeholder="Jane Smith"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor={`contact-email-${variant}`}>Email</Label>
              <Input
                id={`contact-email-${variant}`}
                type="email"
                placeholder="jane@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>

          {isLanding && (
            <div className="space-y-1.5">
              <Label htmlFor="contact-subject-landing">Subject</Label>
              <Input
                id="contact-subject-landing"
                type="text"
                placeholder="How can we help?"
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
              />
            </div>
          )}

          <div className="space-y-1.5">
            <Label htmlFor={`contact-message-${variant}`}>Message</Label>
            <textarea
              id={`contact-message-${variant}`}
              rows={isLanding ? 5 : 4}
              placeholder="Tell us about your project, use case, or question…"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
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
            size={isLanding ? "lg" : "default"}
            className={isLanding ? "w-full text-base" : ""}
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

          <p
            className={cn(
              "text-xs text-muted-foreground",
              isLanding && "text-center",
            )}
          >
            We respect your privacy. Your information is never shared.
          </p>
        </form>
      )}
    </>
  );

  if (!isLanding) {
    return (
      <Card className="gap-0">
        <CardHeader className="border-b">
          <CardTitle className="text-sm font-semibold">Contact Support</CardTitle>
          <p className="text-sm text-muted-foreground">
            Have a question or need help? Send us a message and we'll get back
            to you shortly.
          </p>
        </CardHeader>
        <CardContent className="pt-4 space-y-4">{formBody}</CardContent>
      </Card>
    );
  }

  return (
    <section
      id="contact"
      className="relative bg-background py-24 overflow-hidden"
    >
      <div className="absolute inset-0 pointer-events-none overflow-hidden">
        <div className="contact-blob absolute -top-32 -right-32 w-[520px] h-[520px] rounded-full bg-primary/8 blur-3xl" />
        <div className="contact-blob-slow absolute -bottom-24 -left-24 w-[380px] h-[380px] rounded-full bg-primary/5 blur-3xl" />
        <div className="contact-blob absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[300px] h-[300px] rounded-full bg-blue-500/5 blur-3xl" />
      </div>

      <div className="container mx-auto px-4 relative">
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
          <div className="relative rounded-2xl overflow-hidden shadow-xl min-h-[480px] contact-fade-up contact-fade-up-delay-1">
            <img
              src="https://images.unsplash.com/photo-1551288049-bebda4e38f71?q=80&w=1600&auto=format&fit=crop"
              alt="Data and communication"
              className="absolute inset-0 w-full h-full object-cover"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-background/80 via-background/20 to-transparent" />

            <div className="absolute bottom-0 left-0 p-10">
              <div className="inline-flex items-center gap-2 bg-primary/90 text-primary-foreground rounded-full px-3 py-1 text-sm font-medium mb-4 backdrop-blur-sm">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-white opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-white"></span>
                </span>
                We are online
              </div>
              <h3 className="text-3xl md:text-4xl font-black text-white leading-tight drop-shadow-md">
                Let's build
                <br />
                something great.
              </h3>
            </div>
          </div>

          <div className="bg-card rounded-2xl border border-border shadow-xl p-8 contact-fade-up contact-fade-up-delay-2">
            {formBody}
          </div>
        </div>
      </div>
    </section>
  );
}
