export function ContactBlobs() {
  return (
    <div className="absolute inset-0 pointer-events-none overflow-hidden">
      <div className="contact-blob absolute -top-32 -right-32 w-[520px] h-[520px] rounded-full bg-primary/8 blur-3xl" />
      <div className="contact-blob-slow absolute -bottom-24 -left-24 w-[380px] h-[380px] rounded-full bg-primary/5 blur-3xl" />
      <div className="contact-blob absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[300px] h-[300px] rounded-full bg-blue-500/5 blur-3xl" />
    </div>
  );
}
