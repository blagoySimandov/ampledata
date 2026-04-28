import type { SourceSummary, Template } from "@/api/types";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useApi, useListSources } from "@/hooks";
import { cn } from "@/lib/utils";
import { useNavigate } from "@tanstack/react-router";
import { ArrowRight, Plus, Sparkles } from "lucide-react";

const ENTITY_BANNER_STYLES: Record<string, { bg: string; text: string }> = {
  "Company Intel": { bg: "bg-primary/10", text: "text-primary" },
  People: {
    bg: "bg-blue-50 dark:bg-blue-950/30",
    text: "text-blue-700 dark:text-blue-400",
  },
  Research: {
    bg: "bg-violet-50 dark:bg-violet-950/30",
    text: "text-violet-700 dark:text-violet-400",
  },
  "Real Estate": {
    bg: "bg-emerald-50 dark:bg-emerald-950/30",
    text: "text-emerald-700 dark:text-emerald-400",
  },
  "E-commerce": {
    bg: "bg-fuchsia-50 dark:bg-fuchsia-950/30",
    text: "text-fuchsia-700 dark:text-fuchsia-400",
  },
  Healthcare: {
    bg: "bg-teal-50 dark:bg-teal-950/30",
    text: "text-teal-700 dark:text-teal-400",
  },
};

const DEFAULT_BANNER_STYLE = { bg: "bg-muted", text: "text-muted-foreground" };

interface SourcePickerModalProps {
  template: Template;
  open: boolean;
  onClose: () => void;
}

function TemplateBanner({ template }: { template: Template }) {
  const style =
    ENTITY_BANNER_STYLES[template.entity_type] ?? DEFAULT_BANNER_STYLE;
  return (
    <div className={cn("flex items-center gap-3 rounded-lg px-3 py-2.5", style.bg)}>
      <Sparkles className={cn("size-4 shrink-0", style.text)} />
      <div>
        <p className="text-sm font-black text-foreground">{template.name}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {template.columns_metadata.length} columns · key:{" "}
          {template.key_columns.join(", ")}
        </p>
      </div>
    </div>
  );
}

function SourceRow({
  source,
  onClick,
}: {
  source: SourceSummary;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className="group flex w-full items-center justify-between rounded-lg border border-border px-4 py-3 text-left transition-colors hover:border-primary hover:bg-primary/5"
    >
      <div>
        <p className="text-sm font-bold text-foreground transition-colors group-hover:text-primary">
          {source.name ?? `Source ${source.source_id.slice(0, 8)}`}
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {source.job_count} run{source.job_count !== 1 ? "s" : ""}
        </p>
      </div>
      <ArrowRight className="size-4 text-muted-foreground transition-colors group-hover:text-primary" />
    </button>
  );
}

export function SourcePickerModal({
  template,
  open,
  onClose,
}: SourcePickerModalProps) {
  const api = useApi();
  const { data } = useListSources(api);
  const navigate = useNavigate();
  const sources = data?.sources ?? [];

  const handlePickSource = (sourceId: string) => {
    onClose();
    navigate({
      to: "/sources/$sourceId",
      params: { sourceId },
      search: { templateId: template.id },
    });
  };

  const handleUploadNew = () => {
    onClose();
    navigate({ to: "/app" });
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="gap-0 overflow-hidden p-0 sm:max-w-[500px]">
        <div className="border-b border-border px-6 py-5">
          <DialogHeader className="mb-3 space-y-0">
            <DialogTitle className="text-lg font-black">
              Apply template to a source
            </DialogTitle>
          </DialogHeader>
          <TemplateBanner template={template} />
        </div>
        <div className="space-y-3 px-6 py-5">
          <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            Choose an existing source
          </p>
          <div className="flex flex-col gap-2">
            {sources.length === 0 ? (
              <p className="text-sm italic text-muted-foreground">
                No sources yet.
              </p>
            ) : (
              sources.map((s) => (
                <SourceRow
                  key={s.source_id}
                  source={s}
                  onClick={() => handlePickSource(s.source_id)}
                />
              ))
            )}
          </div>
          <div className="h-px bg-border" />
          <button
            onClick={handleUploadNew}
            className="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-border bg-muted/40 px-4 py-3 text-sm font-bold text-muted-foreground transition-colors hover:border-primary hover:text-primary"
          >
            <Plus className="size-4" /> Upload new dataset
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
