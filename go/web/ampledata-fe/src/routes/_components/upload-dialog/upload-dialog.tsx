import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useFileUpload } from "@/hooks/use-file-upload";
import { FileDropZone } from "./file-drop-zone";
import { UploadButton } from "./upload-button";
import { GoogleSheetsDialogContent } from "../google-sheets-dialog";

type Tab = "csv" | "sheets";

export function UploadDialog() {
  const { open, setOpen, file, setFile, handleUpload, isPending, reset } =
    useFileUpload();
  const [activeTab, setActiveTab] = useState<Tab>("csv");

  const handleOpenChange = (val: boolean) => {
    setOpen(val);
    if (!val) {
      reset();
      setActiveTab("csv");
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button className="font-bold gap-2">
          <Plus className="w-4 h-4" />
          NEW SOURCE
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle className="text-2xl font-black">
            Add Data Source
          </DialogTitle>
        </DialogHeader>
        <TabSwitcher active={activeTab} onChange={setActiveTab} />
        <div className="pt-2">
          {activeTab === "csv" ? (
            <div className="space-y-4">
              <FileDropZone file={file} onFileSelect={setFile} />
              <UploadButton
                isPending={isPending}
                disabled={!file}
                onClick={handleUpload}
              />
            </div>
          ) : (
            <GoogleSheetsDialogContent />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function TabSwitcher({
  active,
  onChange,
}: {
  active: Tab;
  onChange: (t: Tab) => void;
}) {
  return (
    <div className="flex gap-1 rounded-lg bg-muted p-1">
      <TabButton label="Upload CSV" value="csv" active={active} onChange={onChange} />
      <TabButton label="Google Sheets" value="sheets" active={active} onChange={onChange} />
    </div>
  );
}

function TabButton({
  label,
  value,
  active,
  onChange,
}: {
  label: string;
  value: Tab;
  active: Tab;
  onChange: (t: Tab) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => onChange(value)}
      className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
        active === value
          ? "bg-background text-foreground shadow-sm"
          : "text-muted-foreground hover:text-foreground"
      }`}
    >
      {label}
    </button>
  );
}
