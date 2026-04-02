import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { ChevronDown } from "lucide-react";

interface AdvancedSettingsProps {
  keyColumnDescription: string;
  onKeyColumnDescriptionChange: (value: string) => void;
  maxRowsInput: string;
  onMaxRowsInputChange: (value: string) => void;
}

export function AdvancedSettings({
  keyColumnDescription,
  onKeyColumnDescriptionChange,
  maxRowsInput,
  onMaxRowsInputChange,
}: AdvancedSettingsProps) {
  return (
    <Collapsible>
      <CollapsibleTrigger className="flex items-center gap-1 text-xs text-slate-400 font-bold hover:text-slate-600 transition-colors pt-2 group">
        <ChevronDown className="w-3 h-3 transition-transform group-data-[state=open]:rotate-180" />
        Advanced
      </CollapsibleTrigger>
      <CollapsibleContent className="space-y-4 pt-2">
        <div className="space-y-1">
          <Label className="text-xs text-slate-500 font-bold">
            Key Column Definition (AI Context)
          </Label>
          <Input
            placeholder="Optional rules to locate the key columns in raw text..."
            value={keyColumnDescription}
            onChange={(e) => onKeyColumnDescriptionChange(e.target.value)}
            className="text-sm h-9"
          />
        </div>
        <div className="space-y-1">
          <Label className="text-xs text-slate-500 font-bold">
            Max Rows to Enrich
          </Label>
          <p className="text-[10px] text-slate-400 leading-tight">
            Limit how many rows are enriched in this run. Leave blank to enrich
            all rows.
          </p>
          <Input
            type="number"
            min={1}
            placeholder="All rows"
            value={maxRowsInput}
            onChange={(e) => onMaxRowsInputChange(e.target.value)}
            className="text-sm h-9"
          />
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
