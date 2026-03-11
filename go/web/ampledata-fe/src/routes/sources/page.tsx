import { Card, CardContent } from "@/components/ui/card";
import { useApi, useSource } from "@/hooks";
import {
  RefreshCw,
  Link,
  ArrowLeft,
  PanelRightOpen,
  PanelRightClose,
} from "lucide-react";
import { useState } from "react";
import { AddColumnsDialog } from "./add-column-dialog";
import { DataTable } from "./data-table";
import { JobStats } from "./job-stats";
import { JobRunsSidebar } from "./job-sidebar";

export function SourceDetail({ sourceId }: { sourceId: string }) {
  const api = useApi();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const { data: source, isLoading, isError } = useSource(api, sourceId);

  const mostRecentJob = source?.jobs[0];
  const activeJobId = selectedJobId ?? mostRecentJob?.job_id ?? null;
  const activeJob = source?.jobs.find((j) => j.job_id === activeJobId);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 text-gray-500 gap-3 font-medium">
        <RefreshCw className="w-5 h-5 animate-spin" /> Loading source...
      </div>
    );
  }

  if (isError || !source) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardContent className="pt-6 text-red-700 font-medium">
          Failed to load source.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="fixed left-0 right-0 top-16 bottom-0 flex bg-slate-50/50 animate-in fade-in duration-500 z-0 overflow-hidden">
      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0 transition-all duration-300 h-full p-4 sm:p-6 lg:p-8">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 shrink-0">
          <div className="flex items-center gap-4">
            <Link
              to="/"
              className="inline-flex items-center text-xs font-bold text-slate-500 hover:text-primary transition-colors gap-1 uppercase tracking-widest bg-white px-3 py-2 rounded-lg border border-slate-200 shadow-sm active:scale-95"
            >
              <ArrowLeft className="w-3.5 h-3.5" /> BACK
            </Link>
            <h2 className="text-2xl font-black text-slate-900 tracking-tight">
              Dataset Explorer
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <AddColumnsDialog
              sourceId={sourceId}
              mostRecentJob={mostRecentJob}
            />
            {!sidebarOpen && (
              <button
                onClick={() => setSidebarOpen(true)}
                className="bg-white border border-slate-200 text-slate-600 px-4 py-2 rounded-lg shadow-sm hover:bg-slate-50 transition-all active:scale-95 flex items-center gap-2 font-bold text-xs uppercase tracking-widest"
              >
                <PanelRightOpen className="w-4 h-4" /> RUNS
              </button>
            )}
          </div>
        </div>

        <DataTable sourceId={sourceId} jobs={source.jobs} />
      </div>

      {/* Sidebar */}
      <div
        className={`shrink-0 bg-white border-l border-slate-200 shadow-2xl transition-all duration-300 ease-in-out h-full flex flex-col z-10 relative overflow-hidden ${
          sidebarOpen
            ? "w-[400px] opacity-100"
            : "w-0 opacity-0 border-transparent pointer-events-none"
        }`}
      >
        {sidebarOpen && (
          <>
            <div className="p-3 border-b border-slate-100 bg-white/50 backdrop-blur-sm shrink-0 flex items-center">
              <button
                onClick={() => setSidebarOpen(false)}
                className="bg-white border border-slate-200 p-1.5 rounded-md shadow-sm hover:bg-slate-50 transition-all text-slate-400 hover:text-slate-700 active:scale-95"
              >
                <PanelRightClose className="w-3.5 h-3.5" />
              </button>
            </div>
            {activeJob && <JobStats jobId={activeJob.job_id} />}
            <JobRunsSidebar
              jobs={source.jobs}
              selectedJobId={activeJobId}
              onSelect={setSelectedJobId}
              onClose={() => setSidebarOpen(false)}
            />
          </>
        )}
      </div>
    </div>
  );
}
