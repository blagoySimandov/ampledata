import { useState } from "react";
import type { UserResponse } from "@/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SectionCard } from "./section-card";

interface Props {
  user: UserResponse | undefined;
}

export function ContactSection({ user }: Props) {
  const [name, setName] = useState(
    user ? `${user.first_name} ${user.last_name}` : "",
  );
  const [email, setEmail] = useState(user?.email ?? "");
  const [message, setMessage] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const subject = encodeURIComponent("Support request from AmpleData");
    const body = encodeURIComponent(
      `Name: ${name}\nEmail: ${email}\n\nMessage:\n${message}`,
    );
    const link = document.createElement("a");
    link.href = `mailto:support@ampledata.io?subject=${subject}&body=${body}`;
    link.click();
  };

  return (
    <SectionCard
      title="Contact Support"
      description="Have a question or need help? Send us a message and we'll get back to you shortly."
    >
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
        <Button type="submit">Send message</Button>
      </form>
    </SectionCard>
  );
}
