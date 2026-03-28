export function ContactVisual() {
  return (
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
          Let's build<br />something great.
        </h3>
      </div>
    </div>
  );
}
