import { ContactBlobs } from "./contact-blobs";
import { ContactForm } from "./contact-form";
import { ContactHeader } from "./contact-header";
import { ContactVisual } from "./contact-visual";

export function ContactSection() {
  return (
    <section
      id="contact"
      className="relative bg-background py-24 overflow-hidden"
    >
      {/* Animated decorative blobs */}
      <ContactBlobs />

      <div className="container mx-auto px-4 relative">
        {/* Section header */}
        <ContactHeader />

        <div className="grid lg:grid-cols-2 gap-12 items-stretch">
          {/* Left column: Image panel */}
          <ContactVisual />

          {/* Right column: contact form */}
          <ContactForm />
        </div>
      </div>
    </section>
  );
}
