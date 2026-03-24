import { Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import logo from "../../../../assets/ampledata-logo.png";

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

interface LoginLeftSidebarProps {
  onSignInWith: (provider: string) => void;
  onSignIn: () => void;
}

export function LoginLeftSidebar({ onSignInWith, onSignIn }: LoginLeftSidebarProps) {
  return (
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
            onClick={() => onSignInWith("GoogleOAuth")}
          >
            <GoogleIcon />
            Continue with Google
          </Button>

          <Button
            variant="outline"
            size="lg"
            className="w-full justify-start gap-3 h-11"
            onClick={() => onSignInWith("GitHubOAuth")}
          >
            <Github size={18} />
            Continue with GitHub
          </Button>

          <Button
            variant="outline"
            size="lg"
            className="w-full justify-start gap-3 h-11"
            onClick={() => onSignInWith("MicrosoftOAuth")}
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
            onClick={onSignIn}
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
  );
}
