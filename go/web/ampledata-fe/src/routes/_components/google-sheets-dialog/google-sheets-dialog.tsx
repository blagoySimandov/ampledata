import { useState } from "react";
import { TableIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { GoogleSpreadsheet } from "@/api/types";
import {
  useOAuthStatus,
  useInitiateGoogleOAuth,
  useListSpreadsheets,
  useListSheetTabs,
  useCreateGoogleSheetsSource,
} from "@/hooks/sources/use-google-sheets";
import { SpreadsheetPicker } from "./spreadsheet-picker";

export function GoogleSheetsDialogContent() {
  const { data: status, isLoading: statusLoading } = useOAuthStatus();
  const connected = status?.connected ?? false;

  const initiate = useInitiateGoogleOAuth();
  const { data: spreadsheets, isLoading: spreadsheetsLoading } =
    useListSpreadsheets(connected);

  const [selectedSheet, setSelectedSheet] = useState<GoogleSpreadsheet | null>(
    null,
  );
  const [selectedTab, setSelectedTab] = useState<string>("");

  const { data: tabs } = useListSheetTabs(selectedSheet?.id ?? null);
  const create = useCreateGoogleSheetsSource();

  const handleSheetSelect = (sheet: GoogleSpreadsheet) => {
    setSelectedSheet(sheet);
    setSelectedTab("");
  };

  const handleCreate = () => {
    if (!selectedSheet || !selectedTab) return;
    create.mutate({
      spreadsheet_id: selectedSheet.id,
      spreadsheet_url: `https://docs.google.com/spreadsheets/d/${selectedSheet.id}`,
      sheet_name: selectedTab,
      spreadsheet_name: selectedSheet.name,
    });
  };

  if (statusLoading) return <LoadingState />;

  if (!connected) return <ConnectPrompt onConnect={() => initiate.mutate()} isPending={initiate.isPending} />;

  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <p className="text-sm font-medium">Select spreadsheet</p>
        {spreadsheetsLoading ? (
          <LoadingState />
        ) : (
          <SpreadsheetPicker
            spreadsheets={spreadsheets ?? []}
            selected={selectedSheet}
            onSelect={handleSheetSelect}
          />
        )}
      </div>

      {selectedSheet && (
        <div className="space-y-1.5">
          <p className="text-sm font-medium">Select sheet tab</p>
          <Select value={selectedTab} onValueChange={setSelectedTab}>
            <SelectTrigger>
              <SelectValue placeholder="Choose a tab…" />
            </SelectTrigger>
            <SelectContent>
              {(tabs ?? []).map((tab) => (
                <SelectItem key={tab.id} value={tab.name}>
                  {tab.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <Button
        className="w-full font-bold"
        disabled={!selectedSheet || !selectedTab || create.isPending}
        onClick={handleCreate}
      >
        {create.isPending ? "Connecting…" : "Connect Sheet"}
      </Button>
    </div>
  );
}

function ConnectPrompt({ onConnect, isPending }: { onConnect: () => void; isPending: boolean }) {
  return (
    <div className="flex flex-col items-center gap-4 py-6">
      <TableIcon className="text-muted-foreground h-10 w-10" />
      <div className="space-y-1 text-center">
        <p className="font-semibold">Connect Google Account</p>
        <p className="text-muted-foreground text-sm">
          Grant access to read and write your Google Sheets.
        </p>
      </div>
      <Button onClick={onConnect} disabled={isPending} className="font-bold">
        {isPending ? "Redirecting…" : "Connect Google"}
      </Button>
    </div>
  );
}

function LoadingState() {
  return (
    <div className="flex h-20 items-center justify-center">
      <p className="text-muted-foreground text-sm">Loading…</p>
    </div>
  );
}
