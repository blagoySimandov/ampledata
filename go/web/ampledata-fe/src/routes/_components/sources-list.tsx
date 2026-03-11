import { useApi, useListSources } from "../../hooks";
import { EmptyState } from "./empty-state";
import { SourcesTable } from "./sources-table";
import { UploadDialog } from "./upload-dialog";

export function SourcesList() {
  const api = useApi();
  const { data, isLoading, isError, error } = useListSources(api);

  if (isLoading)
    return (
      <div className="text-center py-10 text-gray-500">Loading sources...</div>
    );

  if (isError) {
    return (
      <div className="bg-red-50 text-red-700 p-4 rounded-md">
        Failed to load sources: {error?.message}
      </div>
    );
  }

  const hasSources = data?.sources && data.sources.length > 0;

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-black tracking-tight text-slate-900">
          My Sources
        </h1>
        <UploadDialog />
      </div>

      {!hasSources ? <EmptyState /> : <SourcesTable sources={data.sources} />}
    </div>
  );
}
