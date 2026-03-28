import { useState } from "react";
import { AlertCircle, CheckCircle, Loader2, Send } from "lucide-react";
import type { UserResponse } from "@/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SectionCard } from "./section-card";

const FORMSPREE_ENDPOINT = "https://formspree.io/f/xeerypze";

type SubmitStatus = "idle" | "loading" | "success" | "error";

interface Props {
  user: UserResponse | undefined;
}

export function ContactSection({ user }: Props) {
  const [name, setName] = useState(
    user ? `${user.first_name} ${user.last_name}` : "",
  );
  const [email, setEmail] = useState(user?.email ?? "");
  const [message, setMessage] = useState("");
  const [status, setStatus] = useState<SubmitStatus>("idle");
  const [errorMessage, setErrorMessage] = useState("");

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
        body: JSON.stringify({ name, email, message }),
      });

      if (res.ok) {
        setStatus("success");
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

  return (
    <SectionCard
      title="Contact Support"
      description="Have a question or need help? Send us a message and we'll get back to you shortly."
    >
      {status === "success" ? (
        <div className="flex flex-col items-center text-center gap-4 py-6">
          <div className="w-14 h-14 rounded-full bg-emerald-500/10 flex items-center justify-center">
            <CheckCircle className="size-7 text-emerald-600 dark:text-emerald-400" />
          </div>
          <div>
            <p className="font-semibold text-foreground">Message sent! 🎉</p>
            <p className="text-sm text-muted-foreground mt-1">
              Thanks for reaching out. We'll get back to you within 24 hours.
            </p>
          </div>
          <Button variant="outline" size="sm" onClick={() => setStatus("idle")}>
            Send another message
          </Button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid sm:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="contact-name-account">Name</Label>
              <Input
                id="contact-name-account"
                placeholder="Your name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="contact-email-account">Email</Label>
              <Input
                id="contact-email-account"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="contact-message-account">Message</Label>
            <textarea
              id="contact-message-account"
              placeholder="Tell us how we can help…"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              required
              rows={4}
              className="w-full rounded-md border border-input bg-input/20 px-3 py-2 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 disabled:pointer-events-none disabled:opacity-50 resize-none dark:bg-input/30"
            />
          </div>

          {status === "error" && (
            <div className="flex items-start gap-3 px-4 py-3 rounded-xl bg-destructive/10 border border-destructive/20 text-sm text-destructive">
              <AlertCircle className="size-4 mt-0.5 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          <Button type="submit" disabled={status === "loading"}>
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
        </form>
      )}
    </SectionCard>
  );
}
