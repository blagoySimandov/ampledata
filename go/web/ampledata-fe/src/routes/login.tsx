import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuth } from "@workos-inc/authkit-react";
import { Button } from "@/components/ui/button";
import logo from "../../assets/ampledata-logo.png";

export const Route = createFileRoute("/login")({
  beforeLoad: ({ context }) => {
    if (context.auth.user) {
      throw redirect({ to: "/app" });
    }
  },
  component: LoginPage,
});

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M17.64 9.205c0-.639-.057-1.252-.164-1.841H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z" fill="#4285F4"/>
      <path d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18z" fill="#34A853"/>
      <path d="M3.964 10.71A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.042l3.007-2.332z" fill="#FBBC05"/>
      <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z" fill="#EA4335"/>
    </svg>
  );
}

function GitHubIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"/>
    </svg>
  );
}

function MicrosoftIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 21 21" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <rect x="1" y="1" width="9" height="9" fill="#F25022"/>
      <rect x="11" y="1" width="9" height="9" fill="#7FBA00"/>
      <rect x="1" y="11" width="9" height="9" fill="#00A4EF"/>
      <rect x="11" y="11" width="9" height="9" fill="#FFB900"/>
    </svg>
  );
}

function LoginPage() {
  const { signIn } = useAuth();

  function signInWith(provider: string) {
    signIn({ provider } as Parameters<typeof signIn>[0]);
  }

  return (
    <div className="min-h-screen flex">
      {/* Left sidebar */}
      <div className="w-full lg:w-[440px] shrink-0 flex flex-col justify-between px-10 py-12 bg-background">
        <div>
          <img src={logo} alt="AmpleData" className="h-9 w-auto" />
        </div>

        <div className="flex flex-col gap-8">
          <div className="space-y-2">
            <h1 className="text-3xl font-black tracking-tight text-foreground">
              Welcome back
            </h1>
            <p className="text-sm text-muted-foreground">
              Sign in to your account to continue
            </p>
          </div>

          <div className="flex flex-col gap-3">
            <Button
              variant="outline"
              size="lg"
              className="w-full justify-start gap-3 h-11"
              onClick={() => signInWith("GoogleOAuth")}
            >
              <GoogleIcon />
              Continue with Google
            </Button>

            <Button
              variant="outline"
              size="lg"
              className="w-full justify-start gap-3 h-11"
              onClick={() => signInWith("GitHubOAuth")}
            >
              <GitHubIcon />
              Continue with GitHub
            </Button>

            <Button
              variant="outline"
              size="lg"
              className="w-full justify-start gap-3 h-11"
              onClick={() => signInWith("MicrosoftOAuth")}
            >
              <MicrosoftIcon />
              Continue with Microsoft
            </Button>

            <div className="relative my-1">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs">
                <span className="bg-background px-3 text-muted-foreground">or</span>
              </div>
            </div>

            <Button
              size="lg"
              className="w-full h-11"
              onClick={() => signIn()}
            >
              Sign in with SSO / Email
            </Button>
          </div>
        </div>

        <p className="text-xs text-muted-foreground">
          By signing in, you agree to our{" "}
          <a href="#" className="underline underline-offset-2 hover:text-foreground transition-colors">
            Terms of Service
          </a>{" "}
          and{" "}
          <a href="#" className="underline underline-offset-2 hover:text-foreground transition-colors">
            Privacy Policy
          </a>
          .
        </p>
      </div>

      {/* Right illustration panel */}
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
    </div>
  );
}
