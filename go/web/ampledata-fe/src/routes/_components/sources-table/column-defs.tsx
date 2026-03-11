import { useNavigate } from "@tanstack/react-router";
import type { ColDef } from "ag-grid-community";
import type { SourceSummary } from "@/api";

export function useColumnDefs(): ColDef<SourceSummary>[] {
  const navigate = useNavigate();
  return [
    {
      field: "source_id",
      headerName: "Source",
      flex: 1,
      cellRenderer: (params: { value: string }) => (
        <span className="font-semibold text-primary cursor-pointer hover:underline">
          {params.value}
        </span>
      ),
      onCellClicked: (params) => {
        navigate({
          to: "/sources/$sourceId",
          params: { sourceId: params.data!.source_id },
        });
      },
    },
    {
      field: "job_count",
      headerName: "Runs",
      width: 100,
      cellRenderer: (params: { value: number }) => (
        <span className="font-mono text-slate-600">{params.value}</span>
      ),
    },
    {
      field: "latest_job_status",
      headerName: "Latest Status",
      width: 160,
      cellRenderer: (params: { value: string | null }) => {
        const status = params.value;
        if (!status)
          return <span className="text-slate-400 text-xs italic">No runs</span>;
        let colorClass = "";
        if (status === "COMPLETED")
          colorClass = "bg-emerald-50 text-emerald-700 border-emerald-200";
        else if (status === "RUNNING")
          colorClass = "bg-blue-50 text-blue-700 border-blue-200 animate-pulse";
        else if (status === "FAILED" || status === "CANCELLED")
          colorClass = "bg-red-50 text-red-700 border-red-200";
        else if (status === "PENDING")
          colorClass = "bg-slate-50 text-slate-600 border-slate-200";
        return (
          <div className="flex items-center h-full">
            <div
              className={`px-2 py-0.5 rounded-full border text-xs font-bold uppercase tracking-wider ${colorClass}`}
            >
              {status}
            </div>
          </div>
        );
      },
    },
    {
      field: "created_at",
      headerName: "Created",
      flex: 1,
      valueFormatter: (params) =>
        new Date(params.value).toLocaleDateString(undefined, {
          month: "short",
          day: "numeric",
          year: "numeric",
          hour: "2-digit",
          minute: "2-digit",
        }),
    },
  ];
}
