import { Button } from "@/components/ui/button";
import { Settings2, Plus } from "lucide-react";

interface EmptyFieldsPlaceholderProps {
  onAdd: () => void;
}

export function EmptyFieldsPlaceholder({ onAdd }: EmptyFieldsPlaceholderProps) {
  return (
    <div className="bg-slate-50 border border-slate-100 rounded-2xl p-6 flex flex-col items-center text-center gap-3">
      <div className="w-10 h-10 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100">
        <Settings2 className="w-5 h-5 text-slate-400" />
      </div>
      <div className="space-y-1">
        <p className="text-sm font-bold text-slate-900">Define new fields</p>
        <p className="text-xs text-slate-500 max-w-[280px] leading-relaxed">
          Add columns to enrich. Previous runs are not re-processed.
        </p>
      </div>
      <Button
        variant="outline"
        size="sm"
        onClick={onAdd}
        className="mt-1 font-bold h-8 px-4 border-slate-200"
      >
        <Plus className="w-3 h-3 mr-2" /> ADD YOUR FIRST FIELD
      </Button>
    </div>
  );
}
