import type { GridApi } from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import { Download, Maximize2, RefreshCw, Search } from "lucide-react";
import { useRef } from "react";
import { useColumnDefs } from "./columns/";
import type { SourceJobSummary } from "@/api";
import { useMergedData } from "./use-merged";
import { agtheme } from "@/lib/ag-theme";

export function GridToolbar({
  rowCount,
  isFetching,
  onQuickFilter,
  onExport,
  onAutoSize,
}: {
  rowCount: number;
  isFetching: boolean;
  onQuickFilter: (value: string) => void;
  onExport: () => void;
  onAutoSize: () => void;
}) {
  return (
    <div className="flex items-center gap-3 px-4 py-2.5 border-b border-slate-100 bg-slate-50 shrink-0">
      <span className="text-xs font-black uppercase tracking-widest text-slate-400 shrink-0">
        Merged Data View
      </span>
      {isFetching && (
        <RefreshCw className="w-3 h-3 text-blue-600 animate-spin shrink-0" />
      )}
      <div className="flex-1" />
      <div className="relative">
        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
        <input
          type="text"
          placeholder="Search all columns..."
          onChange={(e) => onQuickFilter(e.target.value)}
          className="h-8 pl-8 pr-3 text-xs rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-1 focus:ring-primary/30 focus:border-primary/50 w-52 placeholder:text-slate-400"
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
        <span>Export CSV</span>
      </button>
      <span className="text-xs font-bold text-slate-400 uppercase tracking-widest shrink-0">
        {rowCount} rows
      </span>
    </div>
  );
}

interface DataTableProps {
  sourceId: string;
  jobs: SourceJobSummary[];
}

export function DataTable({ sourceId, jobs }: DataTableProps) {
  const { data: mergedData, isFetching } = useMergedData(sourceId, jobs);
  const gridRef = useRef<AgGridReact>(null);
  const columnDefs = useColumnDefs(
    mergedData.sourceColumns,
    mergedData.enrichedColumns,
  );

  const getApi = (): GridApi | undefined => gridRef.current?.api;

  const handleQuickFilter = (value: string) =>
    getApi()?.setGridOption("quickFilterText", value);

  const handleExport = () => getApi()?.exportDataAsCsv();

  const handleAutoSize = () => getApi()?.autoSizeAllColumns();

  return (
    <div className="bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col flex-1 min-h-0">
      <GridToolbar
        rowCount={mergedData.rows.length}
        isFetching={isFetching}
        onQuickFilter={handleQuickFilter}
        onExport={handleExport}
        onAutoSize={handleAutoSize}
      />
      <div className="w-full flex-1 min-h-0">
        <AgGridReact
          ref={gridRef}
          rowData={mergedData.rows}
          columnDefs={columnDefs}
          theme={agtheme}
          pagination
          paginationPageSize={100}
          paginationPageSizeSelector={[50, 100, 200, 500]}
          animateRows
          enableCellTextSelection
          defaultColDef={{
            resizable: true,
            sortable: true,
            filter: true,
            floatingFilter: true,
          }}
        />
      </div>
    </div>
  );
}
