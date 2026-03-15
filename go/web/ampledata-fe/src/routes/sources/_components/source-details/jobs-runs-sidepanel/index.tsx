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
      className={`bg-white shadow-2xl transition-all duration-300 ease-in-out flex flex-col z-20
        absolute inset-0
        lg:relative lg:inset-auto lg:shrink-0 lg:h-full lg:border-l lg:border-slate-200
        ${open
          ? "translate-x-0 lg:w-[320px] lg:opacity-100"
          : "translate-x-full pointer-events-none lg:translate-x-0 lg:w-0 lg:opacity-0 lg:border-transparent"
        }`}
    >
      <div className="w-full lg:w-[320px] flex flex-col h-full overflow-hidden pt-4 sm:pt-6 lg:pt-8">
        <div className="px-3 pb-2 shrink-0 flex items-center justify-end lg:hidden">
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
      </div>
    </div>
  );
}
