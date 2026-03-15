import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/account/billing/")({
  beforeLoad: () => {
    throw redirect({ to: "/account" });
  },
});
