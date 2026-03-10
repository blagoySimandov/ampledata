import { createFileRoute, Link } from '@tanstack/react-router';
import { useApi, useSource, useEnrich, useJobRows } from '../hooks';
import { useMemo, useState } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { type ColDef, type ColGroupDef, themeQuartz } from 'ag-grid-community';
import {
  ArrowLeft, RefreshCw, AlertCircle, Info, PanelRightOpen, PanelRightClose,
  Link2, ExternalLink, Plus, Trash2, Settings2, Loader2,
} from 'lucide-react';
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { ColumnMetadata, SourceJobSummary } from '../api';

const myTheme = themeQuartz.withParams({
  accentColor: '#0f172a',
  backgroundColor: '#ffffff',
  borderColor: '#e2e8f0',
  borderRadius: '8px',
  browserColorScheme: 'light',
  chromeBackgroundColor: '#f8fafc',
  fontFamily: 'inherit',
  fontSize: '14px',
  foregroundColor: '#0f172a',
  headerBackgroundColor: '#f8fafc',
  headerFontSize: '14px',
  headerFontWeight: '600',
  headerTextColor: '#64748b',
  rowBorder: { color: '#f1f5f9' },
  wrapperBorder: true,
  wrapperBorderRadius: '12px',
});

function SourcesRenderer(params: { value: string[] }) {
  const sources: string[] = params.value;
  if (!sources || sources.length === 0) {
    return <div className="text-slate-400 text-xs italic mt-1.5 opacity-50">-</div>;
  }
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button className="flex items-center gap-1.5 h-full px-2 hover:bg-slate-100 rounded text-slate-500 hover:text-primary transition-colors focus:outline-none">
          <Link2 className="w-3.5 h-3.5" />
          <span className="text-xs font-bold">{sources.length}</span>
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0 shadow-xl border-slate-200 overflow-hidden z-50">
        <div className="bg-slate-50 border-b border-slate-100 px-3 py-2 flex items-center justify-between">
          <span className="text-xs font-black uppercase tracking-widest text-slate-500 flex items-center gap-1.5">
            <Link2 className="w-3 h-3" /> Data Sources
          </span>
          <Badge variant="secondary" className="text-[10px] font-bold px-1.5 py-0 h-4 min-h-0">
            {sources.length} LINKS
          </Badge>
        </div>
        <div className="max-h-[300px] overflow-y-auto p-2 space-y-1">
          {sources.map((src, i) => {
            let domain = src;
            try { domain = new URL(src).hostname.replace(/^www\./, ''); } catch { /* ignore */ }
            return (
              <a key={i} href={src} target="_blank" rel="noopener noreferrer"
                className="flex items-center justify-between p-2 rounded-lg hover:bg-slate-50 border border-transparent hover:border-slate-100 transition-colors group">
                <div className="flex items-center gap-2 truncate min-w-0">
                  <div className="w-6 h-6 rounded bg-slate-100 flex items-center justify-center shrink-0 border border-slate-200">
                    <img src={`https://icon.horse/icon/${domain}`} alt="" className="w-3 h-3 opacity-70"
                      onError={(e) => { e.currentTarget.style.display = 'none'; }} />
                  </div>
                  <span className="text-xs font-medium text-slate-700 truncate group-hover:text-primary transition-colors">{domain}</span>
                </div>
                <ExternalLink className="w-3 h-3 text-slate-300 group-hover:text-primary shrink-0 ml-2 transition-colors" />
              </a>
            );
          })}
        </div>
      </PopoverContent>
    </Popover>
  );
}

function getConfidenceConfig(score: number) {
  if (score >= 0.8) return { label: "Very High", color: "text-emerald-500", bg: "bg-emerald-500", borderColor: "border-emerald-200" };
  if (score >= 0.7) return { label: "High", color: "text-green-500", bg: "bg-green-500", borderColor: "border-green-200" };
  if (score >= 0.5) return { label: "Medium", color: "text-amber-500", bg: "bg-amber-500", borderColor: "border-amber-200" };
  return { label: "Low", color: "text-red-500", bg: "bg-red-500", borderColor: "border-red-200" };
}

function ConfidenceDataRenderer(params: {
  colDef: { field: string };
  data: { __confidence?: Record<string, { score: number; reason: string }>; __stage: string };
  value: unknown;
}) {
  const field = params.colDef.field;
  const confidence = params.data.__confidence?.[field];
  const hasValue = params.value !== undefined && params.value !== null && params.value !== "";
  const stage = params.data.__stage;
  const confConfig = confidence
    ? getConfidenceConfig(confidence.score)
    : { label: "Unknown", color: "text-slate-400", bg: "bg-slate-200", borderColor: "border-slate-200" };

  let content;
  if (hasValue) {
    content = <span className="font-medium text-slate-900">{String(params.value)}</span>;
  } else if (stage === 'COMPLETED' || stage === 'FAILED') {
    content = (
      <div className="flex items-center gap-1.5 px-2 py-0.5 rounded border text-xs font-bold uppercase tracking-tight bg-amber-50 text-amber-700 border-amber-200">
        <AlertCircle className="w-3 h-3" /> Missing
      </div>
    );
  } else {
    content = <div className="h-2.5 bg-slate-200/80 rounded w-full max-w-[80%] animate-pulse" />;
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button className="flex items-center justify-between w-full group cursor-pointer text-left focus:outline-none h-full min-h-[32px]">
          <div className="flex-1 truncate mr-2">{content}</div>
          {confidence && (
            <div className={`flex items-center gap-1 opacity-40 group-hover:opacity-100 transition-opacity ${confConfig.color}`}>
              <span className="text-xs font-black uppercase tracking-wider">{confConfig.label}</span>
              <Info className="w-3 h-3" />
            </div>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-80 shadow-2xl border-slate-200 p-0 overflow-hidden">
        <div className={`h-1.5 w-full ${confConfig.bg}`} />
        <div className="p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="font-black text-xs uppercase tracking-widest text-slate-400">Field Intelligence</h4>
            <Badge variant="outline" className={`text-xs font-black ${confConfig.color} ${confConfig.borderColor} bg-white`}>
              {confConfig.label} CONFIDENCE
            </Badge>
          </div>
          <p className="text-xs text-slate-500 leading-relaxed italic">
            {confidence ? `"${confidence.reason}"` : "No rationale captured."}
          </p>
          <div className="pt-2 border-t border-slate-100 flex items-center justify-between">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-tight">Stage</span>
            <span className="text-xs font-black text-slate-900 uppercase tracking-widest">{stage}</span>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

function AddColumnsDialog({ sourceId, hasExistingJob }: { sourceId: string; hasExistingJob: boolean }) {
  const api = useApi();
  const enrich = useEnrich(api, sourceId);
  const [open, setOpen] = useState(false);
  const [columnsMetadata, setColumnsMetadata] = useState<ColumnMetadata[]>([]);

  const addColumn = () => setColumnsMetadata(prev => [...prev, { name: '', type: 'string', job_type: 'enrichment' }]);
  const removeColumn = (i: number) => setColumnsMetadata(prev => prev.filter((_, idx) => idx !== i));
  const updateColumn = (i: number, updates: Partial<ColumnMetadata>) =>
    setColumnsMetadata(prev => { const u = [...prev]; u[i] = { ...u[i], ...updates }; return u; });

  const handleEnrich = async () => {
    try {
      await enrich.mutateAsync({ columns_metadata: columnsMetadata });
      setOpen(false);
      setColumnsMetadata([]);
    } catch (e) { console.error('Enrich failed', e); }
  };

  const canStart = columnsMetadata.length > 0 && columnsMetadata.every(c => c.name) && hasExistingJob;

  return (
    <Dialog open={open} onOpenChange={(val) => { setOpen(val); if (!val) setColumnsMetadata([]); }}>
      <DialogTrigger asChild>
        <Button className="font-bold gap-2" disabled={!hasExistingJob}>
          <Plus className="w-4 h-4" /> ADD COLUMNS
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[560px] max-h-[90vh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="p-6 pb-0">
          <DialogTitle className="text-2xl font-black">Add Enrichment Columns</DialogTitle>
        </DialogHeader>
        <ScrollArea className="flex-1 px-6">
          <div className="py-6 space-y-4">
            <div className="flex items-center justify-between">
              <Label className="text-xs font-black uppercase tracking-widest text-slate-400 flex items-center gap-2">
                <Settings2 className="w-3 h-3" /> New Fields
              </Label>
              <Button variant="ghost" size="sm" onClick={addColumn} className="h-7 text-xs font-black px-2 hover:bg-slate-100">
                <Plus className="w-3 h-3 mr-1" /> ADD FIELD
              </Button>
            </div>
            {columnsMetadata.length === 0 ? (
              <div className="bg-slate-50 border border-slate-100 rounded-2xl p-6 flex flex-col items-center text-center gap-3">
                <div className="w-10 h-10 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100">
                  <Settings2 className="w-5 h-5 text-slate-400" />
                </div>
                <div className="space-y-1">
                  <p className="text-sm font-bold text-slate-900">Define new fields</p>
                  <p className="text-xs text-slate-500 max-w-[280px] leading-relaxed">
                    Add columns to enrich. Previous runs are not re-processed.
                  </p>
                </div>
                <Button variant="outline" size="sm" onClick={addColumn} className="mt-1 font-bold h-8 px-4 border-slate-200">
                  <Plus className="w-3 h-3 mr-2" /> ADD YOUR FIRST FIELD
                </Button>
              </div>
            ) : (
              <div className="space-y-3">
                {columnsMetadata.map((col, index) => (
                  <div key={index} className="flex flex-col gap-2 p-3 bg-white border border-slate-100 rounded-xl shadow-sm">
                    <div className="flex items-center gap-2">
                      <Input placeholder="Field name" value={col.name}
                        onChange={(e) => updateColumn(index, { name: e.target.value })}
                        className="h-9 text-xs font-medium" />
                      <Select value={col.job_type} onValueChange={(v: 'enrichment' | 'imputation') => updateColumn(index, { job_type: v })}>
                        <SelectTrigger className="h-9 w-[120px] text-xs font-medium bg-slate-50 border-transparent"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="enrichment">Enrich</SelectItem>
                          <SelectItem value="imputation">Impute</SelectItem>
                        </SelectContent>
                      </Select>
                      <Select value={col.type} onValueChange={(v: 'string' | 'number' | 'boolean' | 'date') => updateColumn(index, { type: v })}>
                        <SelectTrigger className="h-9 w-[100px] text-xs font-medium bg-slate-50 border-transparent"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="string">String</SelectItem>
                          <SelectItem value="number">Number</SelectItem>
                          <SelectItem value="boolean">Bool</SelectItem>
                          <SelectItem value="date">Date</SelectItem>
                        </SelectContent>
                      </Select>
                      <Button variant="ghost" size="icon" onClick={() => removeColumn(index)}
                        className="h-9 w-9 text-slate-400 hover:text-red-500 hover:bg-red-50 shrink-0">
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                    <Input placeholder="Optional AI instructions"
                      value={col.description || ''}
                      onChange={(e) => updateColumn(index, { description: e.target.value })}
                      className="h-8 text-xs bg-slate-50/50 border-slate-100 placeholder:text-slate-400" />
                  </div>
                ))}
              </div>
            )}
            <Button className="w-full font-black h-12" onClick={handleEnrich} disabled={!canStart || enrich.isPending}>
              {enrich.isPending ? <><Loader2 className="w-4 h-4 animate-spin mr-2" />STARTING...</> : 'START ENRICHMENT'}
            </Button>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

function JobRunsSidebar({ jobs, selectedJobId, onSelect, onClose }: {
  jobs: SourceJobSummary[];
  selectedJobId: string | null;
  onSelect: (jobId: string) => void;
  onClose: () => void;
}) {
  return (
    <div className="w-[400px] h-full flex flex-col min-h-0 absolute top-0 left-0 bottom-0">
      <div className="p-6 border-b border-slate-100 bg-slate-50/50 shrink-0 flex items-center justify-between">
        <h2 className="text-sm font-black uppercase tracking-widest text-slate-500">Enrichment Runs</h2>
        <button onClick={onClose} className="bg-white border border-slate-200 p-2 rounded-lg shadow-sm hover:bg-slate-50 transition-all text-slate-400 hover:text-slate-700 active:scale-95">
          <PanelRightClose className="w-4 h-4" />
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {jobs.map((job) => {
          const isSelected = job.job_id === selectedJobId;
          const cols = job.columns_metadata?.map(c => c.name).join(', ') || '—';
          let statusColor = 'bg-slate-100 text-slate-600 border-slate-200';
          if (job.status === 'COMPLETED') statusColor = 'bg-emerald-50 text-emerald-700 border-emerald-200';
          else if (job.status === 'RUNNING') statusColor = 'bg-blue-50 text-blue-700 border-blue-200';
          else if (job.status === 'CANCELLED') statusColor = 'bg-red-50 text-red-700 border-red-200';
          return (
            <button key={job.job_id} onClick={() => onSelect(job.job_id)}
              className={`w-full text-left p-3 rounded-xl border transition-all ${isSelected ? 'border-primary bg-primary/5 shadow-sm' : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50'}`}>
              <div className="flex items-center justify-between mb-1">
                <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border uppercase tracking-wider ${statusColor}`}>
                  {job.status}
                </span>
                <span className="text-[11px] text-slate-400">{new Date(job.created_at).toLocaleDateString()}</span>
              </div>
              <p className="text-xs text-slate-500 truncate mt-1.5">
                <span className="font-semibold text-slate-700">Columns: </span>{cols}
              </p>
              <p className="text-xs text-slate-400">{job.total_rows} rows</p>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function DataTable({ jobId }: { jobId: string }) {
  const api = useApi();
  const [page, setPage] = useState(0);
  const pageSize = 50;
  const { data: rowsData, isFetching } = useJobRows(api, jobId, page * pageSize, pageSize);

  const enrichedColumns = useMemo(() => {
    const cols = new Set<string>();
    rowsData?.rows.forEach(row => {
      if (row.extracted_data) Object.keys(row.extracted_data).forEach(k => cols.add(k));
      if (row.confidence) Object.keys(row.confidence).forEach(k => cols.add(k));
    });
    return Array.from(cols).sort();
  }, [rowsData]);

  const columnDefs = useMemo<(ColDef | ColGroupDef)[]>(() => [
    {
      headerName: 'Original Data',
      children: [{ field: '__key', headerName: 'Source Key', pinned: 'left', width: 200, cellClass: 'bg-slate-50 font-medium' }],
    },
    {
      headerName: 'Enriched Data',
      children: enrichedColumns.map(col => ({
        field: col,
        headerName: col.split('_').map((w: string) => w.charAt(0).toUpperCase() + w.slice(1)).join(' '),
        flex: 1,
        minWidth: 150,
        cellRenderer: ConfidenceDataRenderer,
      })),
    },
    {
      headerName: 'System',
      children: [
        {
          field: '__stage', headerName: 'Stage', width: 120,
          cellRenderer: (params: { value: string }) => {
            let color = "bg-slate-100 text-slate-600";
            if (params.value === 'COMPLETED') color = "bg-emerald-50 text-emerald-700 border-emerald-100";
            else if (params.value === 'FAILED') color = "bg-red-50 text-red-700 border-red-100";
            return (
              <div className="flex items-center h-full">
                <Badge variant="secondary" className={`${color} border font-black text-xs px-2 py-0 h-4.5 min-h-0 leading-none tracking-tighter uppercase`}>
                  {params.value}
                </Badge>
              </div>
            );
          },
        },
        { field: '__sources', headerName: 'Sources', width: 100, cellRenderer: SourcesRenderer },
      ],
    },
  ], [enrichedColumns]);

  const rowData = useMemo(() => rowsData?.rows.map(row => ({
    __key: row.key,
    __stage: row.stage,
    __confidence: row.confidence,
    __sources: row.sources,
    ...row.extracted_data,
  })), [rowsData]);

  return (
    <div className="bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col flex-1 min-h-0">
      <div className="flex items-center gap-2 px-4 py-2 border-b border-slate-100 bg-slate-50 shrink-0">
        <span className="text-xs font-black uppercase tracking-widest text-slate-400">Data</span>
        {isFetching && <RefreshCw className="w-3 h-3 text-blue-600 animate-spin" />}
      </div>
      <div className="w-full flex-1 min-h-0">
        <AgGridReact rowData={rowData || []} columnDefs={columnDefs} theme={myTheme}
          suppressPaginationPanel={true} defaultColDef={{ resizable: true, sortable: true }} />
      </div>
      <div className="p-4 border-t border-slate-100 bg-slate-50 flex justify-between items-center shrink-0">
        <div className="text-xs font-bold text-slate-400 uppercase tracking-widest">
          {rowData?.length || 0} of {rowsData?.pagination.total || 0} records
        </div>
        <div className="flex items-center gap-2">
          <button disabled={page === 0} onClick={() => setPage(p => p - 1)}
            className="px-4 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-600 shadow-sm hover:bg-slate-50 disabled:opacity-40 transition-all">
            Previous
          </button>
          <div className="px-4 py-2 bg-slate-900 text-white rounded-lg text-xs font-bold">Page {page + 1}</div>
          <button disabled={!rowsData?.pagination.has_more} onClick={() => setPage(p => p + 1)}
            className="px-4 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-600 shadow-sm hover:bg-slate-50 disabled:opacity-40 transition-all">
            Next
          </button>
        </div>
      </div>
    </div>
  );
}

function SourceDetailPage() {
  const { sourceId } = Route.useParams();
  const api = useApi();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const { data: source, isLoading, isError } = useSource(api, sourceId);

  const mostRecentJob = source?.jobs[0];
  const activeJobId = selectedJobId ?? mostRecentJob?.job_id ?? null;
  const activeJob = source?.jobs.find(j => j.job_id === activeJobId);
  const hasCompletedJob = source?.jobs.some(j => j.status === 'COMPLETED') ?? false;

  const completedRows = activeJob?.status === 'COMPLETED' ? activeJob.total_rows : 0;
  const progressPercent = activeJob && activeJob.total_rows > 0 ? (completedRows / activeJob.total_rows) * 100 : 0;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 text-gray-500 gap-3 font-medium">
        <RefreshCw className="w-5 h-5 animate-spin" /> Loading source...
      </div>
    );
  }

  if (isError || !source) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardContent className="pt-6 text-red-700 font-medium">Failed to load source.</CardContent>
      </Card>
    );
  }

  return (
    <div className="fixed left-0 right-0 top-16 bottom-0 flex bg-slate-50/50 animate-in fade-in duration-500 z-0 overflow-hidden">
      <div className="flex-1 flex flex-col min-w-0 transition-all duration-300 h-full p-4 sm:p-6 lg:p-8">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 shrink-0">
          <div className="flex items-center gap-4">
            <Link to="/" className="inline-flex items-center text-xs font-bold text-slate-500 hover:text-primary transition-colors gap-1 uppercase tracking-widest bg-white px-3 py-2 rounded-lg border border-slate-200 shadow-sm active:scale-95">
              <ArrowLeft className="w-3.5 h-3.5" /> BACK
            </Link>
            <h2 className="text-2xl font-black text-slate-900 tracking-tight">Dataset Explorer</h2>
          </div>
          <div className="flex items-center gap-2">
            <AddColumnsDialog sourceId={sourceId} hasExistingJob={hasCompletedJob} />
            {!sidebarOpen && (
              <button onClick={() => setSidebarOpen(true)}
                className="bg-white border border-slate-200 text-slate-600 px-4 py-2 rounded-lg shadow-sm hover:bg-slate-50 transition-all active:scale-95 flex items-center gap-2 font-bold text-xs uppercase tracking-widest">
                <PanelRightOpen className="w-4 h-4" /> RUNS
              </button>
            )}
          </div>
        </div>

        {activeJobId ? (
          <DataTable jobId={activeJobId} />
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center text-slate-400">
              <Settings2 className="w-12 h-12 mx-auto mb-3 opacity-40" />
              <p className="font-bold">No enrichment runs yet</p>
              <p className="text-sm mt-1">Click "ADD COLUMNS" to start your first enrichment.</p>
            </div>
          </div>
        )}
      </div>

      <div className={`shrink-0 bg-white border-l border-slate-200 shadow-2xl transition-all duration-300 ease-in-out h-full flex flex-col z-10 relative overflow-hidden ${
        sidebarOpen ? 'w-[400px] opacity-100' : 'w-0 opacity-0 border-transparent pointer-events-none'
      }`}>
        {sidebarOpen && (
          <>
            {activeJob && (
              <div className="p-4 border-b border-slate-100 bg-slate-50/50 shrink-0 space-y-3">
                <div className="flex items-center justify-between">
                  <Badge variant="outline" className="px-3 py-1.5 font-black tracking-widest text-xs uppercase border-slate-300 bg-white">
                    {activeJob.status}
                  </Badge>
                  <span className="text-xs font-bold text-slate-400">{activeJob.total_rows} rows</span>
                </div>
                {activeJob.status === 'RUNNING' && (
                  <div className="space-y-1">
                    <Progress value={progressPercent} className="h-2 bg-slate-200/60" />
                    <div className="flex justify-end">
                      <span className="text-xs font-black text-primary">{progressPercent.toFixed(0)}% COMPLETE</span>
                    </div>
                  </div>
                )}
              </div>
            )}
            <JobRunsSidebar
              jobs={source.jobs}
              selectedJobId={activeJobId}
              onSelect={setSelectedJobId}
              onClose={() => setSidebarOpen(false)}
            />
          </>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute('/sources/$sourceId')({
  component: SourceDetailPage,
});
