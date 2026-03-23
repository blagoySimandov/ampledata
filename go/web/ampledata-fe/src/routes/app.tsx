import { createFileRoute } from "@tanstack/react-router";
import { SourcesList } from "./_components";

export const Route = createFileRoute("/app")({
  component: SourcesList,
});
