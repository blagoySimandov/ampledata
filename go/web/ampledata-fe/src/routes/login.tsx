import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuth } from "@workos-inc/authkit-react";
import { LoginLeftSidebar, LoginHero } from "./_components/login";

export const Route = createFileRoute("/login")({
  beforeLoad: ({ context }) => {
    if (context.auth.user) {
      throw redirect({ to: "/app" });
    }
  },
  component: LoginPage,
});

function LoginPage() {
  const { signIn, getSignInUrl } = useAuth();

  async function signInWith(provider: string) {
    const url = await getSignInUrl();
    const parsed = new URL(url);
    parsed.searchParams.set("provider", provider);
    window.location.assign(parsed.toString());
  }

  return (
    <div className="min-h-screen flex">
      <LoginLeftSidebar onSignInWith={signInWith} onSignIn={() => signIn()} />
      <LoginHero />
    </div>
  );
}
