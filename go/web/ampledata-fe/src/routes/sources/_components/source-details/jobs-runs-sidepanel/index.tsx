import { PanelRightClose } from "lucide-react";
import type { SourceJobSummary } from "@/api";
import { JobRunsSidebar } from "./job-runs-sidebar";
import { JobStats } from "./job-stats";

export function JobRunsSidebarPanel({
  open,
  onClose,
  activeJob,
  jobs,
  selectedJobId,
  onSelect,
}: {
  open: boolean;
  onClose: () => void;
  activeJob: SourceJobSummary | undefined;
  jobs: SourceJobSummary[];
  selectedJobId: string | null;
  onSelect: (id: string) => void;
}) {
  return (
    <div
      className={`shrink-0 bg-white border-l border-slate-200 shadow-2xl transition-all duration-300 ease-in-out h-full flex flex-col z-10 relative overflow-hidden pt-4 sm:pt-6 lg:pt-8 ${
        open
          ? "w-[400px] opacity-100"
          : "w-0 opacity-0 border-transparent pointer-events-none"
      }`}
    >
      {open && (
        <>
          <div className="p-3 border-b border-slate-100 bg-white/50 backdrop-blur-sm shrink-0 flex items-center">
            <button
              onClick={onClose}
              className="bg-white border border-slate-200 p-1.5 rounded-md shadow-sm hover:bg-slate-50 transition-all text-slate-400 hover:text-slate-700 active:scale-95"
            >
              <PanelRightClose className="w-3.5 h-3.5" />
            </button>
          </div>
          {activeJob && <JobStats jobId={activeJob.job_id} />}
          <JobRunsSidebar
            jobs={jobs}
            selectedJobId={selectedJobId}
            onSelect={onSelect}
          />
        </>
      )}
    </div>
  );
}
