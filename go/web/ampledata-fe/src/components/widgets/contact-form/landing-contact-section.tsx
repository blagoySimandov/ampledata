import { Mail } from "lucide-react";

interface Props {
  children: React.ReactNode;
}

export function LandingContactSection({ children }: Props) {
  return (
    <section id="contact" className="relative bg-background py-24 overflow-hidden">
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
            {children}
          </div>
        </div>
      </div>
    </section>
  );
}
