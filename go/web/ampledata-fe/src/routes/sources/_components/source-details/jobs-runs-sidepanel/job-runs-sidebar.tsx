import type { SourceJobSummary } from "@/api";
import { Settings2 } from "lucide-react";
import { JobRunCard } from "./job-run-card";

interface JobRunsSidebarProps {
  jobs: SourceJobSummary[];
  selectedJobId: string | null;
  onSelect: (jobId: string) => void;
}

export function JobRunsSidebar({
  jobs,
  selectedJobId,
  onSelect,
}: JobRunsSidebarProps) {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-slate-50/50">
      <div className="p-4 border-b border-slate-100 flex items-center justify-between bg-white/50 backdrop-blur-sm sticky top-0 z-10">
        <h2 className="text-xs font-black uppercase tracking-widest text-slate-500 flex items-center gap-2">
          <Settings2 className="w-3.5 h-3.5" /> All Runs
        </h2>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2.5">
        {jobs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center text-slate-400 py-12">
            <Settings2 className="w-8 h-8 mb-3 opacity-40" />
            <p className="font-bold text-sm">No enrichment runs yet</p>
            <p className="text-xs mt-1">
              Click "ADD COLUMNS" to start your first enrichment.
            </p>
          </div>
        ) : (
          jobs.map((job) => (
            <JobRunCard
              key={job.job_id}
              job={job}
              isSelected={job.job_id === selectedJobId}
              onSelect={() => onSelect(job.job_id)}
            />
          ))
        )}
      </div>
    </div>
  );
}
