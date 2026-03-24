import { useState } from "react";
import {
  CheckCircle,
  Github,
  Linkedin,
  Mail,
  MapPin,
  Phone,
  Send,
  Twitter,
} from "lucide-react";
import { Button } from "@/components/ui/button";

const CONTACT_ITEMS = [
  {
    icon: Mail,
    label: "Email",
    value: "hello@ampledata.io",
    href: "mailto:hello@ampledata.io",
  },
  {
    icon: Phone,
    label: "Phone",
    value: "+1 (555) 000-0000",
    href: "tel:+15550000000",
  },
  {
    icon: MapPin,
    label: "Address",
    value: "123 Data Street, San Francisco, CA 94107",
    href: null,
  },
] as const;

const SOCIAL_LINKS = [
  {
    icon: Twitter,
    label: "Twitter / X",
    href: "https://twitter.com/ampledata",
  },
  {
    icon: Linkedin,
    label: "LinkedIn",
    href: "https://linkedin.com/company/ampledata",
  },
  {
    icon: Github,
    label: "GitHub",
    href: "https://github.com/ampledata",
  },
] as const;

interface FormState {
  name: string;
  email: string;
  subject: string;
  message: string;
}

const EMPTY_FORM: FormState = { name: "", email: "", subject: "", message: "" };

export function ContactSection() {
  const [formState, setFormState] = useState<FormState>(EMPTY_FORM);
  const [submitted, setSubmitted] = useState(false);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    // TODO: wire up to a real API endpoint or email service when ready
    setSubmitted(true);
    setFormState(EMPTY_FORM);
    setTimeout(() => setSubmitted(false), 4000);
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

        <div className="grid lg:grid-cols-2 gap-12 items-start">
          {/* Left column: photo + contact details + social */}
          <div className="space-y-8 contact-fade-up contact-fade-up-delay-1">
            {/* Hero photo with overlay */}
            <div className="relative rounded-2xl overflow-hidden shadow-xl border border-border group">
              <img
                src="https://images.unsplash.com/photo-1521737711867-e3b97375f902?auto=format&fit=crop&w=900&q=80"
                alt="AmpleData team"
                className="w-full h-64 object-cover transition-transform duration-700 group-hover:scale-105"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/10 to-transparent" />
              <div className="absolute bottom-0 left-0 right-0 p-6">
                <p className="text-white font-bold text-xl leading-snug drop-shadow-md">
                  Let's build something great together
                </p>
                <p className="text-white/70 text-sm mt-1">
                  Enterprise? Startup? We work with teams of all sizes.
                </p>
              </div>
            </div>

            {/* Contact info cards */}
            <div className="space-y-3">
              {CONTACT_ITEMS.map(({ icon: Icon, label, value, href }) => (
                <div
                  key={label}
                  className="flex items-start gap-4 p-4 rounded-xl bg-secondary/40 border border-border hover:border-primary/30 hover:shadow-sm hover:bg-primary/5 transition-all duration-200"
                >
                  <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
                    <Icon className="size-5 text-primary" />
                  </div>
                  <div>
                    <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-0.5">
                      {label}
                    </p>
                    {href ? (
                      <a
                        href={href}
                        className="text-foreground font-medium hover:text-primary transition-colors"
                      >
                        {value}
                      </a>
                    ) : (
                      <p className="text-foreground font-medium">{value}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>

            {/* Response time badge */}
            <div className="flex items-center gap-3 px-4 py-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20">
              <div className="w-2.5 h-2.5 rounded-full bg-emerald-500 shadow-[0_0_6px_2px_rgba(16,185,129,0.4)] shrink-0" />
              <p className="text-sm text-emerald-600 dark:text-emerald-400 font-medium">
                Typical response time: under 24 hours on business days
              </p>
            </div>

            {/* Social links */}
            <div>
              <p className="text-sm font-semibold text-muted-foreground mb-3 uppercase tracking-wider">
                Follow us
              </p>
              <div className="flex gap-3">
                {SOCIAL_LINKS.map(({ icon: Icon, label, href }) => (
                  <a
                    key={label}
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    aria-label={label}
                    className="w-10 h-10 rounded-lg bg-secondary/40 border border-border flex items-center justify-center text-muted-foreground hover:text-primary hover:border-primary/30 hover:bg-primary/5 hover:shadow-sm transition-all duration-200"
                  >
                    <Icon className="size-4" />
                  </a>
                ))}
              </div>
            </div>
          </div>

          {/* Right column: contact form */}
          <div className="bg-card rounded-2xl border border-border shadow-xl p-8 contact-fade-up contact-fade-up-delay-2">
            {submitted ? (
              <div className="flex flex-col items-center justify-center text-center py-16 gap-5">
                <div className="w-20 h-20 rounded-full bg-emerald-100 flex items-center justify-center shadow-lg">
                  <CheckCircle className="size-10 text-emerald-600" />
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

                <Button type="submit" size="lg" className="w-full text-base">
                  Send message <Send className="ml-2 size-4" />
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
