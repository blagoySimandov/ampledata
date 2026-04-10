import { Button } from "@/components/ui/button";
import logo from "../../../../assets/ampledata-logo.png";
import { GoogleIcon, AppleIcon, GithubIcon } from "@/components/icons";
import { Link } from "@tanstack/react-router";

interface LoginLeftSidebarProps {
  onSignInWith: (provider: string) => void;
  onSignIn: () => void;
}

export function LoginLeftSidebar({
  onSignInWith,
  onSignIn,
}: LoginLeftSidebarProps) {
  return (
    <div className="w-full lg:w-[440px] shrink-0 flex flex-col justify-between px-10 py-12 bg-background">
      <div>
        <Link to="/">
          <img src={logo} alt="AmpleData" className="h-12 w-auto" />
        </Link>
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
            <GithubIcon />
            Continue with GitHub
          </Button>

          <Button
            variant="outline"
            size="lg"
            className="w-full justify-start gap-3 h-11"
            onClick={() => onSignInWith("AppleOAuth")}
          >
            <AppleIcon />
            Continue with Apple
          </Button>

          <div className="relative my-1">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs">
              <span className="bg-background px-3 text-muted-foreground">
                or
              </span>
            </div>
          </div>

          <Button size="lg" className="w-full h-11" onClick={onSignIn}>
            Sign in with SSO / Email
          </Button>
        </div>
      </div>

      <p className="text-xs text-muted-foreground">
        By signing in, you agree to our{" "}
        <a
          href="#"
          className="underline underline-offset-2 hover:text-foreground transition-colors"
        >
          Terms of Service
        </a>{" "}
        and{" "}
        <a
          href="#"
          className="underline underline-offset-2 hover:text-foreground transition-colors"
        >
          Privacy Policy
        </a>
        .
      </p>
    </div>
  );
}
