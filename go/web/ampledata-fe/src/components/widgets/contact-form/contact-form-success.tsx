import { CheckCircle } from "lucide-react";
import { Button } from "@/components/ui/button";

interface Props {
  onReset: () => void;
}

export function ContactFormSuccess({ onReset }: Props) {
  return (
    <div className="flex flex-col items-center justify-center text-center py-16 gap-5">
      <div className="w-20 h-20 rounded-full bg-emerald-500/10 flex items-center justify-center shadow-lg">
        <CheckCircle className="size-10 text-emerald-600 dark:text-emerald-400" />
      </div>
      <div>
        <h3 className="text-2xl font-black text-foreground mb-2">Message sent! 🎉</h3>
        <p className="text-muted-foreground">
          Thanks for reaching out. We'll get back to you within 24 hours.
        </p>
      </div>
      <Button variant="outline" onClick={onReset}>
        Send another message
      </Button>
    </div>
  );
}
