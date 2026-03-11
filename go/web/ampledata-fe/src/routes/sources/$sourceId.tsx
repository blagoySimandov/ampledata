import { createFileRoute } from "@tanstack/react-router";
import { SourceDetail } from "./page";

const SourceDetailPage = () => {
  const { sourceId } = Route.useParams();
  return <SourceDetail sourceId={sourceId} />;
};

export const Route = createFileRoute("/sources/$sourceId")({
  component: SourceDetailPage,
});
