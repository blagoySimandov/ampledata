import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { Trash2 } from "lucide-react";
import type { ColumnEditorProps } from "./types";

export function ColumnEditor({
  column,
  sourceColumns,
  onUpdate,
  onRemove,
}: ColumnEditorProps) {
  const isImputation = column.job_type === "imputation";
  return (
    <div className="flex flex-col gap-1.5 p-2.5 bg-white border border-slate-100 rounded-xl shadow-sm">
      <div className="flex items-center gap-1.5">
        {isImputation ? (
          <Select
            value={column.name}
            onValueChange={(v) => onUpdate({ name: v })}
          >
            <SelectTrigger className="h-8 flex-1 text-[10px] font-medium">
              <SelectValue placeholder="Select column to impute" />
            </SelectTrigger>
            <SelectContent>
              {sourceColumns.map((col) => (
                <SelectItem key={col} value={col}>
                  {col}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <Input
            placeholder="Field name"
            value={column.name}
            onChange={(e) => onUpdate({ name: e.target.value })}
            className="h-8 text-[10px] font-medium"
          />
        )}
        <Select
          value={column.job_type}
          onValueChange={(v: "enrichment" | "imputation") =>
            onUpdate({ job_type: v })
          }
        >
          <SelectTrigger className="h-8 w-[110px] text-[10px] font-medium bg-slate-50 border-transparent">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="enrichment">Enrich</SelectItem>
            <SelectItem value="imputation">Impute</SelectItem>
          </SelectContent>
        </Select>
        <Select
          value={column.type}
          onValueChange={(v: "string" | "number" | "boolean" | "date") =>
            onUpdate({ type: v })
          }
        >
          <SelectTrigger className="h-8 w-[90px] text-[10px] font-medium bg-slate-50 border-transparent">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="string">String</SelectItem>
            <SelectItem value="number">Number</SelectItem>
            <SelectItem value="boolean">Bool</SelectItem>
            <SelectItem value="date">Date</SelectItem>
          </SelectContent>
        </Select>
        <Button
          variant="ghost"
          size="icon"
          onClick={onRemove}
          className="h-8 w-8 text-slate-400 hover:text-red-500 hover:bg-red-50 shrink-0"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </Button>
      </div>
      <Input
        placeholder="Optional AI instructions"
        value={column.description ?? ""}
        onChange={(e) => onUpdate({ description: e.target.value })}
        className="h-7 text-[10px] bg-slate-50/50 border-slate-100 placeholder:text-slate-400"
      />
    </div>
  );
}
