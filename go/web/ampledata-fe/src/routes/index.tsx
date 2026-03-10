import { createFileRoute, useNavigate } from '@tanstack/react-router';
import {
  useApi,
  useListSources,
  useGetSignedUrl,
  useUploadFile,
} from '../hooks';
import { useRef, useState, useMemo } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { type ColDef, ModuleRegistry, AllCommunityModule, themeQuartz } from 'ag-grid-community';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Plus, Upload, Loader2, ChevronRight, Database } from 'lucide-react';
import type { SourceSummary } from '../api';

ModuleRegistry.registerModules([AllCommunityModule]);

const myTheme = themeQuartz.withParams({
  accentColor: '#0f172a',
  backgroundColor: '#ffffff',
  borderColor: '#e2e8f0',
  borderRadius: '8px',
  browserColorScheme: 'light',
  chromeBackgroundColor: '#f8fafc',
  fontFamily: 'inherit',
  fontSize: '15px',
  foregroundColor: '#0f172a',
  headerBackgroundColor: '#f1f5f9',
  headerFontSize: '15px',
  headerFontWeight: '600',
  headerTextColor: '#475569',
  rowBorder: { color: '#f1f5f9' },
  wrapperBorder: true,
  wrapperBorderRadius: '8px',
});

function UploadDialog() {
  const api = useApi();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const getSignedUrl = useGetSignedUrl(api);
  const uploadFile = useUploadFile(api);

  const reset = () => setFile(null);

  const handleUpload = async () => {
    if (!file) return;
    try {
      const { url, sourceId } = await getSignedUrl.mutateAsync({
        contentType: 'text/csv',
        length: file.size,
      });
      await uploadFile.mutateAsync({ url, file });
      setOpen(false);
      reset();
      navigate({ to: '/sources/$sourceId', params: { sourceId } });
    } catch (error) {
      console.error('Upload failed', error);
    }
  };

  const isPending = getSignedUrl.isPending || uploadFile.isPending;

  return (
    <Dialog open={open} onOpenChange={(val) => { setOpen(val); if (!val) reset(); }}>
      <DialogTrigger asChild>
        <Button className="font-bold gap-2">
          <Plus className="w-4 h-4" />
          NEW SOURCE
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle className="text-2xl font-black">Upload Dataset</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-2">
          <div
            className="border-2 border-dashed border-slate-200 rounded-2xl p-10 flex flex-col items-center justify-center gap-4 hover:border-primary/50 transition-colors cursor-pointer bg-slate-50/50"
            onClick={() => fileInputRef.current?.click()}
          >
            <input
              type="file"
              ref={fileInputRef}
              onChange={(e) => e.target.files?.[0] && setFile(e.target.files[0])}
              className="hidden"
              accept=".csv"
            />
            <div className="w-12 h-12 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100">
              <Upload className="w-6 h-6 text-slate-400" />
            </div>
            <div className="text-center">
              <p className="font-bold text-slate-900">{file ? file.name : 'Select a CSV file'}</p>
              <p className="text-xs text-slate-500 mt-1">Maximum size 50MB</p>
            </div>
          </div>
          <Button
            className="w-full font-black h-12"
            disabled={!file || isPending}
            onClick={handleUpload}
          >
            {isPending ? (
              <><Loader2 className="w-4 h-4 animate-spin mr-2" />UPLOADING...</>
            ) : (
              <>CONTINUE<ChevronRight className="w-4 h-4 ml-2" /></>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function SourcesList() {
  const api = useApi();
  const navigate = useNavigate();
  const { data, isLoading, isError, error } = useListSources(api);

  const columnDefs = useMemo<ColDef<SourceSummary>[]>(() => [
    {
      field: 'source_id',
      headerName: 'Source',
      flex: 1,
      cellRenderer: (params: { value: string }) => (
        <span className="font-semibold text-primary cursor-pointer hover:underline">
          {params.value}
        </span>
      ),
      onCellClicked: (params) => {
        navigate({ to: '/sources/$sourceId', params: { sourceId: params.data!.source_id } });
      },
    },
    {
      field: 'job_count',
      headerName: 'Runs',
      width: 100,
      cellRenderer: (params: { value: number }) => (
        <span className="font-mono text-slate-600">{params.value}</span>
      ),
    },
    {
      field: 'latest_job_status',
      headerName: 'Latest Status',
      width: 160,
      cellRenderer: (params: { value: string | null }) => {
        const status = params.value;
        if (!status) return <span className="text-slate-400 text-xs italic">No runs</span>;
        let colorClass = '';
        if (status === 'COMPLETED') colorClass = 'bg-emerald-50 text-emerald-700 border-emerald-200';
        else if (status === 'RUNNING') colorClass = 'bg-blue-50 text-blue-700 border-blue-200 animate-pulse';
        else if (status === 'FAILED' || status === 'CANCELLED') colorClass = 'bg-red-50 text-red-700 border-red-200';
        else if (status === 'PENDING') colorClass = 'bg-slate-50 text-slate-600 border-slate-200';
        return (
          <div className="flex items-center h-full">
            <div className={`px-2 py-0.5 rounded-full border text-xs font-bold uppercase tracking-wider ${colorClass}`}>
              {status}
            </div>
          </div>
        );
      },
    },
    {
      field: 'created_at',
      headerName: 'Created',
      flex: 1,
      valueFormatter: (params) => new Date(params.value).toLocaleDateString(undefined, {
        month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit',
      }),
    },
  ], [navigate]);

  if (isLoading) return <div className="text-center py-10 text-gray-500">Loading sources...</div>;

  if (isError) {
    return (
      <div className="bg-red-50 text-red-700 p-4 rounded-md">
        Failed to load sources: {error?.message}
      </div>
    );
  }

  const hasSources = data?.sources && data.sources.length > 0;

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-black tracking-tight text-slate-900">My Sources</h1>
        <UploadDialog />
      </div>

      {!hasSources ? (
        <div className="border-2 border-dashed border-slate-200 rounded-3xl p-16 flex flex-col items-center justify-center text-center bg-slate-50/50">
          <div className="w-20 h-20 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100 mb-6">
            <Database className="w-10 h-10 text-slate-400" />
          </div>
          <h2 className="text-2xl font-black text-slate-900 mb-2">No Sources Yet</h2>
          <p className="text-slate-500 max-w-md mb-8">
            Upload a CSV to create your first source. You can then add columns and run enrichments at any time.
          </p>
          <UploadDialog />
        </div>
      ) : (
        <div className="w-full h-[600px] bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden">
          <AgGridReact
            rowData={data?.sources || []}
            columnDefs={columnDefs}
            theme={myTheme}
            pagination={true}
            paginationPageSize={10}
            paginationPageSizeSelector={[10, 20, 50]}
            domLayout="normal"
          />
        </div>
      )}
    </div>
  );
}

export const Route = createFileRoute('/')({
  component: SourcesList,
});
