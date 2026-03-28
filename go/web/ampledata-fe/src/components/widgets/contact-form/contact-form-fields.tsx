import { AlertCircle, Loader2, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import type { SubmitStatus } from "./types";

interface Props {
  variant: "landing" | "account";
  name: string;
  email: string;
  subject: string;
  message: string;
  status: SubmitStatus;
  errorMessage: string;
  onNameChange: (value: string) => void;
  onEmailChange: (value: string) => void;
  onSubjectChange: (value: string) => void;
  onMessageChange: (value: string) => void;
  onSubmit: (e: React.FormEvent) => void;
}

export function ContactFormFields({
  variant,
  name,
  email,
  subject,
  message,
  status,
  errorMessage,
  onNameChange,
  onEmailChange,
  onSubjectChange,
  onMessageChange,
  onSubmit,
}: Props) {
  const isLanding = variant === "landing";

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <div>
        <h3
          className={cn(
            "text-foreground mb-1",
            isLanding ? "text-xl font-black" : "text-base font-semibold",
          )}
        >
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
            onChange={(e) => onNameChange(e.target.value)}
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
            onChange={(e) => onEmailChange(e.target.value)}
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
            onChange={(e) => onSubjectChange(e.target.value)}
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
          onChange={(e) => onMessageChange(e.target.value)}
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
  );
}
