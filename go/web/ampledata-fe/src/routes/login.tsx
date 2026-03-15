import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuth } from "@workos-inc/authkit-react";
import { Button } from "@/components/ui/button";
import logo from "../../assets/ampledata-logo.png";

export const Route = createFileRoute("/login")({
  beforeLoad: ({ context }) => {
    if (context.auth.user) {
      throw redirect({ to: "/" });
    }
  },
  component: LoginPage,
});

function LoginPage() {
  const { signIn } = useAuth();

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-8 p-8">
        <img src={logo} alt="AmpleData" className="h-14 w-auto" />
        <div className="text-center space-y-2">
          <h1 className="text-2xl font-black text-foreground">
            Welcome to AmpleData
          </h1>
          <p className="text-sm text-muted-foreground">
            Sign in to enrich and explore your datasets
          </p>
        </div>
        <Button
          size="lg"
          className="w-full max-w-xs"
          onClick={() => signIn()}
        >
          Sign in
        </Button>
      </div>
    </div>
  );
}
