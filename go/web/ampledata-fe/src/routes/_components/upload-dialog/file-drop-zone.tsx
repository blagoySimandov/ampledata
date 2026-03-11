import { Upload } from "lucide-react";

interface FileDropZoneProps {
  file: File | null;
  onFileSelect: (file: File) => void;
}

export function FileDropZone({ file, onFileSelect }: FileDropZoneProps) {
  return (
    <label className="border-2 border-dashed border-slate-200 rounded-2xl p-10 flex flex-col items-center justify-center gap-4 hover:border-primary/50 transition-colors cursor-pointer bg-slate-50/50">
      <input
        type="file"
        onChange={(e) => e.target.files?.[0] && onFileSelect(e.target.files[0])}
        className="hidden"
        accept=".csv"
      />
      <div className="w-12 h-12 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100">
        <Upload className="w-6 h-6 text-slate-400" />
      </div>
      <div className="text-center">
        <p className="font-bold text-slate-900">
          {file ? file.name : "Select a CSV file"}
        </p>
        <p className="text-xs text-slate-500 mt-1">Maximum size 50MB</p>
      </div>
    </label>
  );
}
