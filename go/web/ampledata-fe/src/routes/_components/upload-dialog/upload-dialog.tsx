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

export function UploadDialog() {
  const { open, setOpen, file, setFile, handleUpload, isPending, reset } =
    useFileUpload();

  return (
    <Dialog
      open={open}
      onOpenChange={(val) => {
        setOpen(val);
        if (!val) reset();
      }}
    >
      <DialogTrigger asChild>
        <Button className="font-bold gap-2">
          <Plus className="w-4 h-4" />
          NEW SOURCE
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle className="text-2xl font-black">
            Upload Dataset
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-2">
          <FileDropZone file={file} onFileSelect={setFile} />
          <UploadButton
            isPending={isPending}
            disabled={!file}
            onClick={handleUpload}
          />
        </div>
      </DialogContent>
    </Dialog>
  );
}
