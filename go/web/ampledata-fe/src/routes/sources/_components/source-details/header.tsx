import { Link, ArrowLeft, PanelRightOpen } from "lucide-react";
import type { SourceJobSummary } from "@/api";
import { AddColumnsDialog } from "./add-column-dialog";

type SourceDetailProps = {
  sourceId: string;
  mostRecentJob?: SourceJobSummary;
  sidebarOpen: boolean;
  onOpenSidebar: () => void;
};
export function SourceDetailHeader({
  sourceId,
  mostRecentJob,
  sidebarOpen,
  onOpenSidebar,
}: SourceDetailProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 shrink-0">
      <div className="flex items-center gap-4">
        <Link to="/" className="...">
          <ArrowLeft className="w-3.5 h-3.5" /> BACK
        </Link>
        <h2 className="text-2xl font-black text-slate-900 tracking-tight">
          Dataset Explorer
        </h2>
      </div>
      <div className="flex items-center gap-2">
        <AddColumnsDialog sourceId={sourceId} mostRecentJob={mostRecentJob} />
        {!sidebarOpen && (
          <button onClick={onOpenSidebar} className="...">
            <PanelRightOpen className="w-4 h-4" /> RUNS
          </button>
        )}
      </div>
    </div>
  );
}
