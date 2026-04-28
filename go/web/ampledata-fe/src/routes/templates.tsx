import { createFileRoute } from "@tanstack/react-router";
import { TemplatesPage } from "./templates/page";

export const Route = createFileRoute("/templates")({
  component: TemplatesPage,
});
