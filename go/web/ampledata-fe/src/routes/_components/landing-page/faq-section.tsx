import { useState } from "react";
import { FAQS } from "./constants";

function FaqItem({
  faq,
  open,
  onToggle,
}: {
  faq: (typeof FAQS)[number];
  open: boolean;
  onToggle: () => void;
}) {
  return (
    <div className="border-b border-border last:border-0">
      <button
        onClick={onToggle}
        className="w-full flex justify-between items-center py-5 text-left gap-4 bg-transparent border-0 cursor-pointer"
      >
        <span className="text-lg font-bold text-foreground">{faq.q}</span>
        <span
          className="text-lg text-muted-foreground shrink-0 transition-transform duration-200 leading-none"
          style={{ transform: open ? "rotate(45deg)" : "none" }}
        >
          +
        </span>
      </button>
      {open && (
        <div className="pb-5 text-base text-muted-foreground leading-relaxed animate-fade-up">
          {faq.a}
        </div>
      )}
    </div>
  );
}

export function FaqSection() {
  const [openIdx, setOpenIdx] = useState<number | null>(null);
  const toggle = (i: number) => setOpenIdx((prev) => (prev === i ? null : i));

  return (
    <section className="py-14 md:py-20 bg-background border-t border-border">
      <div className="max-w-[920px] mx-auto px-4 md:px-6">
        <div className="text-center mb-12">
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-foreground mb-2.5">
            Frequently asked
          </h2>
          <p className="text-base text-muted-foreground">Real questions. Honest answers.</p>
        </div>
        <div>
          {FAQS.map((faq, i) => (
            <FaqItem key={faq.q} faq={faq} open={openIdx === i} onToggle={() => toggle(i)} />
          ))}
        </div>
      </div>
    </section>
  );
}
