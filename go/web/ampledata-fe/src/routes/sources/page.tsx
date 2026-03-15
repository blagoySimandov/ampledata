import { SourceDetailHeader } from "./_components/source-details/header";
import { JobRunsSidebarPanel } from "./_components/source-details/jobs-runs-sidepanel";
import { LoadingState, ErrorState } from "./_components/source-details/state";
import { DataTable } from "./_components/source-details/data-table";
import { useSourceDetail } from "./_hooks";

export function SourceDetail({ sourceId }: { sourceId: string }) {
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

  if (isLoading) return <LoadingState />;
  if (isError || !source) return <ErrorState />;

  return (
    <div className="fixed left-0 right-0 top-16 bottom-0 flex bg-slate-50/50 animate-in fade-in duration-500 z-0 overflow-hidden">
      <div className="flex-1 flex flex-col min-w-0 transition-all duration-300 h-full p-4 sm:p-6 lg:p-8">
        <SourceDetailHeader
          sourceId={sourceId}
          mostRecentJob={mostRecentJob}
          sidebarOpen={sidebarOpen}
          onOpenSidebar={() => setSidebarOpen(true)}
        />
        <DataTable sourceId={sourceId} jobs={source.jobs} />
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
