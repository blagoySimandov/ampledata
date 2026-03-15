import type { ColDef, ColGroupDef } from "ag-grid-community";
import { useMemo } from "react";
import { ConfidenceDataRenderer } from "./confidence-renderer";

export function useColumnDefs(
  sourceColumns: string[],
  enrichedColumns: string[],
): (ColDef | ColGroupDef)[] {
  return useMemo<(ColDef | ColGroupDef)[]>(() => {
    if (sourceColumns.length === 0) return [];

    return [
      {
        headerName: "Original Data",
        openByDefault: true,
        children: sourceColumns.map(
          (col, idx): ColDef => ({
            field: col,
            headerName: col,
            columnGroupShow: idx === 0 ? undefined : "open",
            minWidth: 150,
            cellClass: "bg-slate-50 font-medium",
          }),
        ),
      },
      {
        headerName: "Enriched Data",
        children: enrichedColumns.map(
          (col): ColDef => ({
            field: col,
            headerName: col,
            flex: 1,
            minWidth: 150,
            cellRenderer: ConfidenceDataRenderer,
          }),
        ),
      },
    ];
  }, [sourceColumns, enrichedColumns]);
}
