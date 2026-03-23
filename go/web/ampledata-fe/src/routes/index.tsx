import { createFileRoute, redirect } from "@tanstack/react-router";
import { LandingPage } from "./_components/landing-page";

export const Route = createFileRoute("/")({
  beforeLoad: ({ context }) => {
    if (!context.auth.isLoading && context.auth.user) {
      throw redirect({ to: "/app" });
    }
  },
  component: LandingPage,
});
