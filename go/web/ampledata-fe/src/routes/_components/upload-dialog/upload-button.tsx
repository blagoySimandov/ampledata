import { Button } from "@/components/ui/button";
import { Loader2, ChevronRight } from "lucide-react";

interface UploadButtonProps {
  isPending: boolean;
  disabled: boolean;
  onClick: () => void;
}

export function UploadButton({
  isPending,
  disabled,
  onClick,
}: UploadButtonProps) {
  return (
    <Button
      className="w-full font-black h-12"
      disabled={disabled || isPending}
      onClick={onClick}
    >
      {isPending ? (
        <>
          <Loader2 className="w-4 h-4 animate-spin mr-2" />
          UPLOADING...
        </>
      ) : (
        <>
          CONTINUE
          <ChevronRight className="w-4 h-4 ml-2" />
        </>
      )}
    </Button>
  );
}
