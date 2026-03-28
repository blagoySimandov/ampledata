import { useState } from "react";
import { Mail } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function ContactSection() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const subject = encodeURIComponent("Contact from AmpleData");
    const body = encodeURIComponent(
      `Name: ${name}\nEmail: ${email}\n\nMessage:\n${message}`,
    );
    const link = document.createElement("a");
    link.href = `mailto:support@ampledata.io?subject=${subject}&body=${body}`;
    link.click();
  };

  return (
    <section id="contact" className="bg-secondary/30 py-24">
      <div className="container mx-auto px-4">
        <div className="text-center mb-12">
          <div className="inline-flex items-center gap-2 bg-primary/10 text-primary rounded-full px-4 py-1.5 text-sm font-medium mb-6">
            <Mail className="size-4" />
            <span>Get in touch</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-4">
            Have questions? We'd love to hear from you.
          </h2>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            Whether you have a question about features, pricing, or anything
            else, our team is ready to answer all your questions.
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="max-w-xl mx-auto space-y-5"
        >
          <div className="grid sm:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="contact-name-landing">Name</Label>
              <Input
                id="contact-name-landing"
                placeholder="Your name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="contact-email-landing">Email</Label>
              <Input
                id="contact-email-landing"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="contact-message-landing">Message</Label>
            <textarea
              id="contact-message-landing"
              placeholder="Tell us how we can help…"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              required
              rows={5}
              className="w-full rounded-md border border-input bg-input/20 px-3 py-2 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 disabled:pointer-events-none disabled:opacity-50 resize-none dark:bg-input/30"
            />
          </div>
          <Button type="submit" className="w-full">
            Send message
          </Button>
        </form>
      </div>
    </section>
  );
}
