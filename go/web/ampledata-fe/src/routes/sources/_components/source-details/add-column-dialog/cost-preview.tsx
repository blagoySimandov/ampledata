interface CostPreviewProps {
  effectiveRows: number;
  columnsCount: number;
}

export function CostPreview({ effectiveRows, columnsCount }: CostPreviewProps) {
  const estimatedCells = effectiveRows * columnsCount;

  if (columnsCount === 0 || estimatedCells === 0) return null;

  return (
    <p className="text-[11px] text-slate-400 text-center">
      Estimated cost:{" "}
      <span className="font-bold text-slate-600">
        {estimatedCells.toLocaleString()} credit{estimatedCells !== 1 ? "s" : ""}
      </span>{" "}
      ({effectiveRows.toLocaleString()} row{effectiveRows !== 1 ? "s" : ""} ×{" "}
      {columnsCount} column{columnsCount !== 1 ? "s" : ""})
    </p>
  );
}
