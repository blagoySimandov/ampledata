import { createFileRoute } from "@tanstack/react-router";
import { SourceDetail } from "./page";

const SourceDetailPage = () => {
  const { sourceId } = Route.useParams();
  const { templateId } = Route.useSearch();
  return <SourceDetail sourceId={sourceId} templateId={templateId} />;
};

export const Route = createFileRoute("/sources/$sourceId")({
  validateSearch: (search: Record<string, unknown>) => ({
    templateId:
      typeof search.templateId === "string" ? search.templateId : undefined,
  }),
  component: SourceDetailPage,
});
