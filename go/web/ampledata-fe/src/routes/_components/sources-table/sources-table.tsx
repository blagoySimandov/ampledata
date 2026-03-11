import { AgGridReact } from "ag-grid-react";
import type { SourceSummary } from "../../../api";
import { useColumnDefs } from "./column-defs";
import { agtheme } from "@/lib/ag-theme";

interface SourcesTableProps {
  sources: SourceSummary[];
}

export function SourcesTable({ sources }: SourcesTableProps) {
  const columnDefs = useColumnDefs();
  return (
    <div className="w-full h-[600px] bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden">
      <AgGridReact
        rowData={sources}
        columnDefs={columnDefs}
        theme={agtheme}
        pagination={true}
        paginationPageSize={10}
        paginationPageSizeSelector={[10, 20, 50]}
        domLayout="normal"
      />
    </div>
  );
}
