import { Link } from "@tanstack/react-router";
import {
  ArrowLeft,
  Download,
  Maximize2,
  PanelRightClose,
  PanelRightOpen,
  RefreshCw,
  Search,
} from "lucide-react";
import type { SourceJobSummary } from "@/api";
import { AddColumnsDialog } from "../add-column-dialog";

interface GridToolbarProps {
  rowCount: number;
  isFetching: boolean;
  onQuickFilter: (value: string) => void;
  onExport: () => void;
  onAutoSize: () => void;
  sourceId: string;
  mostRecentJob?: SourceJobSummary;
  sidebarOpen: boolean;
  onToggleSidebar: () => void;
}

export function GridToolbar({
  rowCount,
  isFetching,
  onQuickFilter,
  onExport,
  onAutoSize,
  sourceId,
  mostRecentJob,
  sidebarOpen,
  onToggleSidebar,
}: GridToolbarProps) {
  return (
    <div className="flex items-center gap-3 px-4 py-2.5 border-b border-slate-100 bg-slate-50 shrink-0 overflow-x-auto">
      <Link
        to="/app"
        className="flex items-center gap-1 text-xs font-bold text-slate-400 hover:text-slate-700 transition-colors uppercase tracking-widest shrink-0"
      >
        <ArrowLeft className="w-3 h-3" />
      </Link>
      <div className="w-px h-4 bg-slate-200 shrink-0" />
      <h2 className="text-xl font-black text-slate-900 tracking-tight shrink-0">
        Dataset Explorer
      </h2>
      {isFetching && (
        <RefreshCw className="w-3 h-3 text-blue-600 animate-spin shrink-0" />
      )}
      <div className="flex-1" />
      <div className="relative">
        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
        <input
          type="text"
          placeholder="Search..."
          onChange={(e) => onQuickFilter(e.target.value)}
          className="h-8 pl-8 pr-3 text-xs rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-1 focus:ring-primary/30 focus:border-primary/50 w-32 sm:w-48 placeholder:text-slate-400"
        />
      </div>
      <button
        onClick={onAutoSize}
        title="Auto-size columns"
        className="h-8 px-2.5 flex items-center gap-1.5 text-xs font-bold text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
      >
        <Maximize2 className="w-3.5 h-3.5" />
      </button>
      <button
        onClick={onExport}
        title="Export CSV"
        className="h-8 px-2.5 flex items-center gap-1.5 text-xs font-bold text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
      >
        <Download className="w-3.5 h-3.5" />
        <span className="hidden sm:inline">Export CSV</span>
      </button>
      <AddColumnsDialog sourceId={sourceId} mostRecentJob={mostRecentJob} />
      <button
        onClick={onToggleSidebar}
        title={sidebarOpen ? "Close jobs panel" : "Open jobs panel"}
        className="h-8 px-2.5 flex items-center gap-1.5 text-xs font-bold text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors shrink-0"
      >
        {sidebarOpen ? (
          <PanelRightClose className="w-3.5 h-3.5" />
        ) : (
          <PanelRightOpen className="w-3.5 h-3.5" />
        )}
      </button>
      <span className="hidden sm:inline text-xs font-bold text-slate-400 uppercase tracking-widest shrink-0">
        {rowCount} rows
      </span>
    </div>
  );
}
