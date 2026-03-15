import { createFileRoute } from "@tanstack/react-router";
import { BillingPage } from "./page";

export const Route = createFileRoute("/account/billing/")({
  component: BillingPage,
});
