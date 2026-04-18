import { Database, Loader2, PlayCircle } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { useApi, useCreateSampleSource } from "@/hooks";
import { UploadDialog } from "./upload-dialog";

export function EmptyState() {
  const api = useApi();
  const navigate = useNavigate();
  const createSample = useCreateSampleSource(api);

  const handleLoadSample = async () => {
    try {
      const source = await createSample.mutateAsync();
      navigate({ to: "/sources/$sourceId", params: { sourceId: source.source_id } });
    } catch {
      toast.error("Failed to load sample dataset");
    }
  };

  return (
    <div className="border-2 border-dashed border-slate-200 rounded-3xl p-16 flex flex-col items-center justify-center text-center bg-slate-50/50">
      <div className="w-20 h-20 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100 mb-6">
        <Database className="w-10 h-10 text-slate-400" />
      </div>
      <h2 className="text-2xl font-black text-slate-900 mb-2">
        No Sources Yet
      </h2>
      <p className="text-slate-500 max-w-md mb-8">
        Upload a CSV to create your first source. Or try our sample dataset — 10 SaaS companies ready to enrich in one click.
      </p>
      <div className="flex flex-col sm:flex-row items-center gap-3">
        <button
          onClick={handleLoadSample}
          disabled={createSample.isPending}
          className="inline-flex items-center gap-2 h-11 px-6 rounded-lg bg-primary text-primary-foreground font-black text-sm hover:bg-primary/90 transition-colors disabled:opacity-60"
        >
          {createSample.isPending ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <PlayCircle className="w-4 h-4" />
          )}
          TRY SAMPLE DATASET
        </button>
        <span className="text-xs text-slate-400 font-bold uppercase tracking-widest">or</span>
        <UploadDialog />
      </div>
    </div>
  );
}
