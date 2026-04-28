import type { Template } from "@/api/types";
import { useApi } from "@/hooks";
import { useListTemplates } from "@/hooks/templates";
import { JobRunsSidebarPanel } from "./_components/source-details/jobs-runs-sidepanel";
import { LoadingState, ErrorState } from "./_components/source-details/state";
import { DataTable } from "./_components/source-details/data-table";
import { useSourceDetail } from "./_hooks";

interface SourceDetailProps {
  sourceId: string;
  templateId?: string;
}

export function SourceDetail({ sourceId, templateId }: SourceDetailProps) {
  const api = useApi();
  const { data: templateData } = useListTemplates(api);
  const {
    source,
    isLoading,
    isError,
    sidebarOpen,
    setSidebarOpen,
    setSelectedJobId,
    mostRecentJob,
    activeJobId,
    activeJob,
  } = useSourceDetail(sourceId);

  const initialTemplate: Template | undefined = templateId
    ? templateData?.templates.find((t) => t.id === templateId)
    : undefined;

  if (isLoading) return <LoadingState />;
  if (isError || !source) return <ErrorState />;

  return (
    <div className="fixed left-0 right-0 top-16 bottom-0 flex bg-slate-50/50 animate-in fade-in duration-500 z-0 overflow-hidden">
      <div className="flex-1 flex flex-col min-w-0 transition-all duration-300 h-full p-3 sm:p-4">
        <DataTable
          sourceId={sourceId}
          jobs={source.jobs}
          mostRecentJob={mostRecentJob}
          sidebarOpen={sidebarOpen}
          onToggleSidebar={() => setSidebarOpen((prev) => !prev)}
          initialTemplate={initialTemplate}
        />
      </div>

      <JobRunsSidebarPanel
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        activeJob={activeJob}
        jobs={source.jobs}
        selectedJobId={activeJobId}
        onSelect={setSelectedJobId}
      />
    </div>
  );
}
