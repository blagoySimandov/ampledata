import { useState } from "react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { GoogleSpreadsheet } from "@/api/types";

interface SpreadsheetPickerProps {
  spreadsheets: GoogleSpreadsheet[];
  selected: GoogleSpreadsheet | null;
  onSelect: (s: GoogleSpreadsheet) => void;
}

export function SpreadsheetPicker({
  spreadsheets,
  selected,
  onSelect,
}: SpreadsheetPickerProps) {
  const [search, setSearch] = useState("");

  const filtered = spreadsheets.filter((s) =>
    s.name.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="space-y-2">
      <Input
        placeholder="Search spreadsheets…"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />
      <ScrollArea className="h-48 rounded-md border">
        <div className="p-1">
          {filtered.length === 0 && (
            <p className="text-muted-foreground py-4 text-center text-sm">
              No spreadsheets found
            </p>
          )}
          {filtered.map((s) => (
            <button
              key={s.id}
              type="button"
              onClick={() => onSelect(s)}
              className={`w-full rounded px-3 py-2 text-left text-sm transition-colors ${
                selected?.id === s.id
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-muted"
              }`}
            >
              {s.name}
            </button>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
