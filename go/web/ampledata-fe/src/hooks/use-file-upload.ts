import { logEvent } from "@/lib/analytics";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useApi, useGetSignedUrl, useUploadFile } from "./index";

function readCsvHeaders(file: File): Promise<string[]> {
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      const firstLine = text.split("\n")[0] ?? "";
      const headers = firstLine.split(",").map((h) => h.trim().replace(/^"|"$/g, ""));
      resolve(headers.filter(Boolean));
    };
    reader.onerror = () => resolve([]);
    reader.readAsText(file);
  });
}

export function useFileUpload() {
  const api = useApi();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);

  const getSignedUrl = useGetSignedUrl(api);
  const uploadFile = useUploadFile(api);

  const reset = () => setFile(null);

  const handleUpload = async () => {
    if (!file) return;
    logEvent("upload_started", { file_size: file.size });
    try {
      const headers = await readCsvHeaders(file);
      const { url, sourceId } = await getSignedUrl.mutateAsync({
        contentType: "text/csv",
        length: file.size,
        headers,
      });
      await uploadFile.mutateAsync({ url, file });
      logEvent("upload_success", { file_size: file.size });
      setOpen(false);
      reset();
      navigate({ to: "/sources/$sourceId", params: { sourceId } });
    } catch (error) {
      logEvent("upload_error");
      console.error("Upload failed", error);
    }
  };

  return {
    open,
    setOpen,
    file,
    setFile,
    handleUpload,
    isPending: getSignedUrl.isPending || uploadFile.isPending,
    reset,
  };
}
