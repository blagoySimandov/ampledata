export function LoginHero() {
  return (
    <div className="hidden lg:flex flex-1 relative overflow-hidden login-hero">
      <div className="login-orb login-orb-1" />
      <div className="login-orb login-orb-2" />
      <div className="login-orb login-orb-3" />
      <div className="login-orb login-orb-4" />
      <div className="login-orb login-orb-5" />
      <div className="login-grid" />
      <div className="absolute inset-0 bg-black/45" />
      <div className="relative z-10 flex flex-col justify-end p-16 text-white">
        <p className="text-4xl font-black leading-tight tracking-tight max-w-xs">
          Turn the web into your database.
        </p>
        <p className="mt-4 text-base text-white/65 max-w-xs">
          Enrich, explore, and understand your data with the power of AmpleData's intelligent pipeline.
        </p>
      </div>
    </div>
  );
}
