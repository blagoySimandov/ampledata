import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useApi, useGetSignedUrl, useUploadFile } from "./index";

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
    try {
      const { url, sourceId } = await getSignedUrl.mutateAsync({
        contentType: "text/csv",
        length: file.size,
      });
      await uploadFile.mutateAsync({ url, file });
      setOpen(false);
      reset();
      navigate({ to: "/sources/$sourceId", params: { sourceId } });
    } catch (error) {
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
