import { Database } from "lucide-react";
import { UploadDialog } from "./upload-dialog";

export function EmptyState() {
  return (
    <div className="border-2 border-dashed border-slate-200 rounded-3xl p-16 flex flex-col items-center justify-center text-center bg-slate-50/50">
      <div className="w-20 h-20 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100 mb-6">
        <Database className="w-10 h-10 text-slate-400" />
      </div>
      <h2 className="text-2xl font-black text-slate-900 mb-2">
        No Sources Yet
      </h2>
      <p className="text-slate-500 max-w-md mb-8">
        Upload a CSV to create your first source. You can then add columns and
        run enrichments at any time.
      </p>
      <UploadDialog />
    </div>
  );
}
