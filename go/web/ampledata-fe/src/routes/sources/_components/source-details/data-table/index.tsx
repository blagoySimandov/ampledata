import type { GridApi } from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import { useRef } from "react";
import { useColumnDefs } from "./_columns";
import type { SourceJobSummary } from "@/api";
import { useMergedData } from "./_hooks/use-merged-data";
import { agtheme } from "@/lib/ag-theme";
import { GridToolbar } from "./grid-toolbar";

interface DataTableProps {
  sourceId: string;
  jobs: SourceJobSummary[];
  mostRecentJob?: SourceJobSummary;
  sidebarOpen: boolean;
  onToggleSidebar: () => void;
}

export function DataTable({
  sourceId,
  jobs,
  mostRecentJob,
  sidebarOpen,
  onToggleSidebar,
}: DataTableProps) {
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
    <div className="bg-white border border-slate-200 rounded-xl overflow-hidden flex flex-col flex-1 min-h-0">
      <GridToolbar
        rowCount={mergedData.rows.length}
        isFetching={isFetching}
        onQuickFilter={handleQuickFilter}
        onExport={handleExport}
        onAutoSize={handleAutoSize}
        sourceId={sourceId}
        mostRecentJob={mostRecentJob}
        sidebarOpen={sidebarOpen}
        onToggleSidebar={onToggleSidebar}
      />
      <div className="w-full flex-1 min-h-0">
        <AgGridReact
          ref={gridRef}
          rowData={mergedData.rows}
          columnDefs={columnDefs}
          theme={agtheme}
          loading={isFetching && mergedData.rows.length === 0}
          getRowId={(params) => params.data.__index}
          pagination
          paginationPageSize={100}
          paginationPageSizeSelector={[50, 100, 200, 500]}
          animateRows
          enableCellTextSelection
          rowHeight={36}
          defaultColDef={{
            resizable: true,
            sortable: true,
            filter: true,
          }}
        />
      </div>
    </div>
  );
}
