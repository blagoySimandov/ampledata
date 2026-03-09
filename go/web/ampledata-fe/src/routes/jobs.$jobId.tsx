import { createFileRoute, Link } from '@tanstack/react-router';
import { useApi, useJobProgress, useJobRows } from '../hooks';
import { useMemo, useState } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { type ColDef, type ColGroupDef, themeQuartz } from 'ag-grid-community';
import { ArrowLeft, RefreshCw, AlertCircle, Info } from 'lucide-react';
import { 
  Popover, 
  PopoverContent, 
  PopoverTrigger 
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

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

function ConfidenceDataRenderer(params: any) {
  const field = params.colDef.field;
  const confidence = params.data.__confidence?.[field];
  const hasValue = params.value !== undefined && params.value !== null && params.value !== "";
  const stage = params.data.__stage;
  
  // If we have a value, we show it.
  // If we don't have a value, we show a "Missing" badge or a "Pending" indicator.
  const content = hasValue ? (
    <span className="font-medium text-slate-900">{String(params.value)}</span>
  ) : (
    <div className={`flex items-center gap-1.5 px-2 py-0.5 rounded border text-xs font-bold uppercase tracking-tight ${
      stage === 'COMPLETED' || stage === 'FAILED' 
        ? 'bg-amber-50 text-amber-700 border-amber-200' 
        : 'bg-slate-50 text-slate-400 border-slate-200 animate-pulse'
    }`}>
      <AlertCircle className="w-3 h-3" />
      {stage === 'COMPLETED' || stage === 'FAILED' ? 'Missing' : 'Pending'}
    </div>
  );

  // If we have neither a value nor confidence info, and the job is done, it's truly N/A.
  // But usually, we want to allow users to click and see why there's no data.
  const hasConfidence = !!confidence;

  // Color coding based on score (default to neutral if no score)
  let scoreColor = "text-slate-400";
  if (hasConfidence) {
    if (confidence.score >= 0.8) scoreColor = "text-emerald-500";
    else if (confidence.score >= 0.5) scoreColor = "text-amber-500";
    else scoreColor = "text-red-500";
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button className="flex items-center justify-between w-full group cursor-pointer text-left focus:outline-none h-full min-h-[32px]">
          <div className="flex-1 truncate mr-2">
            {content}
          </div>
          {hasConfidence && (
            <div className={`flex items-center gap-1 opacity-40 group-hover:opacity-100 transition-opacity ${scoreColor}`}>
              <span className="text-xs font-black">{(confidence.score * 100).toFixed(0)}%</span>
              <Info className="w-3 h-3" />
            </div>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-80 shadow-2xl border-slate-200 p-0 overflow-hidden">
        <div className={`h-1.5 w-full ${hasConfidence ? (confidence.score >= 0.8 ? 'bg-emerald-500' : confidence.score >= 0.5 ? 'bg-amber-500' : 'bg-red-500') : 'bg-slate-200'}`} />
        <div className="p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="font-black text-xs uppercase tracking-widest text-slate-400">Field Intelligence</h4>
            {hasConfidence ? (
              <Badge variant="outline" className={`text-xs font-black ${scoreColor} border-current bg-white`}>
                {(confidence.score * 100).toFixed(0)}% CONFIDENCE
              </Badge>
            ) : (
              <Badge variant="outline" className="text-xs font-black text-slate-400 border-slate-200 bg-white">
                NO METADATA
              </Badge>
            )}
          </div>
          
          <div className="space-y-1">
            <div className="text-xs font-bold text-slate-900 flex items-center gap-1.5">
               {hasValue ? "Extraction Rationale" : "Reason for Missing Data"}
            </div>
            <p className="text-xs text-slate-500 leading-relaxed italic">
              {hasConfidence 
                ? `"${confidence.reason}"` 
                : hasValue 
                  ? "This field was successfully extracted but no additional rationale was provided by the AI." 
                  : "No data was returned for this field and no specific reason was captured during processing."}
            </p>
          </div>

          <div className="pt-2 border-t border-slate-100 flex items-center justify-between">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-tight">System Status</span>
            <span className="text-xs font-black text-slate-900 uppercase tracking-widest">{stage}</span>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

function JobDetail() {
  const { jobId } = Route.useParams();
  const api = useApi();
  const [page, setPage] = useState(0);
  const pageSize = 50;
  
  const { 
    data: progress, 
    isLoading: isLoadingProgress, 
    isError: isErrorProgress,
    refetch: refetchProgress
  } = useJobProgress(api, jobId);

  const { 
    data: rowsData, 
    isFetching: isFetchingRows
  } = useJobRows(api, jobId, page * pageSize, pageSize);

  // Discover all possible enriched columns
  const enrichedColumns = useMemo(() => {
    const cols = new Set<string>();
    rowsData?.rows.forEach(row => {
      if (row.extracted_data) {
        Object.keys(row.extracted_data).forEach(k => cols.add(k));
      }
      if (row.confidence) {
        Object.keys(row.confidence).forEach(k => cols.add(k));
      }
    });
    return Array.from(cols).sort();
  }, [rowsData]);

  const columnDefs = useMemo<(ColDef | ColGroupDef)[]>(() => {
    const defs: (ColDef | ColGroupDef)[] = [
      { 
        headerName: 'Original Data',
        children: [
          { 
            field: '__key', 
            headerName: 'Source Key', 
            pinned: 'left', 
            width: 200, 
            cellClass: 'bg-slate-50 font-medium' 
          },
        ]
      },
      {
        headerName: 'Enriched Data',
        children: enrichedColumns.map(col => ({
          field: col,
          headerName: col.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' '),
          flex: 1,
          minWidth: 150,
          cellRenderer: ConfidenceDataRenderer
        }))
      },
      {
        headerName: 'System',
        children: [
          { 
            field: '__stage', 
            headerName: 'Stage', 
            width: 120,
            cellRenderer: (params: any) => {
              const stage = params.value;
              let color = "bg-slate-100 text-slate-600";
              if (stage === 'COMPLETED') color = "bg-emerald-50 text-emerald-700 border-emerald-100";
              else if (stage === 'FAILED') color = "bg-red-50 text-red-700 border-red-100";
              
              return (
                <div className="flex items-center h-full">
                  <Badge 
                    variant="secondary" 
                    className={`${color} border font-black text-xs px-2 py-0 h-4.5 min-h-0 leading-none tracking-tighter uppercase`}
                  >
                    {stage}
                  </Badge>
                </div>
              );
            }
          }
        ]
      }
    ];
    return defs;
  }, [enrichedColumns]);

  const rowData = useMemo(() => {
    return rowsData?.rows.map(row => ({
      __key: row.key,
      __stage: row.stage,
      __confidence: row.confidence,
      ...row.extracted_data
    }));
  }, [rowsData]);

  const intelligenceStats = useMemo(() => {
    if (!rowsData?.rows || rowsData.rows.length === 0 || enrichedColumns.length === 0) {
      return { avgConfidence: 0, missingDataPct: 0 };
    }

    let totalScore = 0;
    let scoreCount = 0;
    let missingCount = 0;
    const totalPossibleFields = rowsData.rows.length * enrichedColumns.length;

    rowsData.rows.forEach(row => {
      enrichedColumns.forEach(col => {
        const value = row.extracted_data?.[col];
        const hasValue = value !== undefined && value !== null && value !== "";
        if (!hasValue) missingCount++;

        const score = row.confidence?.[col]?.score;
        if (score !== undefined) {
          totalScore += score;
          scoreCount++;
        }
      });
    });

    return {
      avgConfidence: scoreCount > 0 ? (totalScore / scoreCount) * 100 : 0,
      missingDataPct: (missingCount / totalPossibleFields) * 100
    };
  }, [rowsData, enrichedColumns]);

  if (isLoadingProgress) {
    return <div className="flex items-center justify-center py-20 text-gray-500 gap-3 font-medium">
      <RefreshCw className="w-5 h-5 animate-spin" />
      Loading job session...
    </div>;
  }

  if (isErrorProgress || !progress) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardContent className="pt-6 text-red-700 font-medium">
          Failed to load job details. The session might have expired.
        </CardContent>
      </Card>
    );
  }

  const completedRows = progress.rows_by_stage?.COMPLETED || 0;
  const progressPercent = progress.total_rows > 0 ? (completedRows / progress.total_rows) * 100 : 0;

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <Link to="/" className="inline-flex items-center text-xs font-semibold text-slate-500 hover:text-primary transition-colors mb-4 gap-1">
            <ArrowLeft className="w-3.5 h-3.5" />
            BACK TO JOBS
          </Link>
          <h1 className="text-3xl font-black tracking-tight text-slate-900">
            {jobId.replace(/\.csv$/i, '')}
          </h1>
          <p className="text-sm text-slate-500 mt-1 flex items-center gap-2">
            Started {new Date(progress.started_at).toLocaleDateString()} at {new Date(progress.started_at).toLocaleTimeString()}
          </p>
        </div>
        
        <div className="flex items-center gap-3">
           <Badge variant="outline" className="px-3 py-1 font-bold tracking-widest text-xs uppercase border-slate-300">
              {progress.status}
           </Badge>
           <button 
             onClick={() => refetchProgress()} 
             className="bg-white border border-slate-200 p-2.5 rounded-lg shadow-sm hover:bg-slate-50 transition-all text-slate-600 active:scale-95"
           >
             <RefreshCw className={`w-4 h-4 ${isFetchingRows ? 'animate-spin' : ''}`} />
           </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2 border-slate-200 shadow-sm overflow-hidden border-none bg-slate-100/50 h-full">
          <CardHeader className="pb-3 border-b border-white bg-white/50">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
                <RefreshCw className={`w-3.5 h-3.5 ${progress.status === 'RUNNING' ? 'animate-spin' : ''}`} />
                Enrichment Status
              </CardTitle>
              <span className="text-xs font-black text-slate-900">{completedRows} / {progress.total_rows} ROWS</span>
            </div>
          </CardHeader>
          <CardContent className="pt-6 space-y-6">
            <div className="space-y-2">
              <Progress value={progressPercent} className="h-2.5 bg-white border border-slate-200" />
              <div className="flex justify-end">
                 <span className="text-xs font-black text-primary">{progressPercent.toFixed(0)}% COMPLETE</span>
              </div>
            </div>
            
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              {Object.entries(progress.rows_by_stage || {}).map(([stage, count]) => (
                <div key={stage} className="bg-white p-3 rounded-xl border border-slate-200/60 shadow-sm">
                  <div className="text-xs text-slate-400 font-black uppercase tracking-widest mb-1">{stage}</div>
                  <div className="text-xl font-black text-slate-900 leading-none">{count}</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-slate-200 shadow-sm overflow-hidden border-none bg-slate-100/50 h-full">
          <CardHeader className="pb-3 border-b border-white bg-white/50">
            <CardTitle className="text-sm font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
              <Info className="w-3.5 h-3.5" />
              Intelligence Stats
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6 space-y-6">
            <div className="space-y-4">
              <div className="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex items-center justify-between">
                <div>
                  <div className="text-xs text-slate-400 font-black uppercase tracking-widest mb-1">Avg. Confidence</div>
                  <div className={`text-2xl font-black ${intelligenceStats.avgConfidence >= 80 ? 'text-emerald-500' : intelligenceStats.avgConfidence >= 50 ? 'text-amber-500' : 'text-red-500'}`}>
                    {intelligenceStats.avgConfidence.toFixed(1)}%
                  </div>
                </div>
                <div className="w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center border border-slate-100">
                  <Info className="w-5 h-5 text-slate-400" />
                </div>
              </div>

              <div className="bg-white p-4 rounded-2xl border border-slate-200/60 shadow-sm flex items-center justify-between">
                <div>
                  <div className="text-xs text-slate-400 font-black uppercase tracking-widest mb-1">Data Gaps</div>
                  <div className="text-2xl font-black text-slate-900">
                    {intelligenceStats.missingDataPct.toFixed(1)}%
                  </div>
                </div>
                <div className="w-10 h-10 rounded-full bg-amber-50 flex items-center justify-center border border-amber-100">
                  <AlertCircle className="w-5 h-5 text-amber-500" />
                </div>
              </div>
            </div>
            
            <p className="text-xs text-slate-400 font-medium italic leading-relaxed px-1">
              Stats are derived from the {rowsData?.rows.length || 0} records currently visible in this view.
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-black text-slate-900 tracking-tight">Dataset Explorer</h2>
          {isFetchingRows && <div className="flex items-center gap-2 text-xs font-bold text-blue-600 animate-pulse uppercase tracking-widest">
            <RefreshCw className="w-3 h-3 animate-spin" /> Synchronizing...
          </div>}
        </div>
        
        <div className="bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col">
          <div className="w-full h-[600px]">
            <AgGridReact
              rowData={rowData || []}
              columnDefs={columnDefs}
              theme={myTheme}
              domLayout="normal"
              suppressPaginationPanel={true}
              defaultColDef={{
                resizable: true,
                sortable: true,
              }}
            />
          </div>
          <div className="p-4 border-t border-slate-100 bg-slate-50/50 flex flex-col sm:flex-row justify-between items-center gap-4">
            <div className="text-xs font-bold text-slate-400 uppercase tracking-widest">
              Showing {rowData?.length || 0} of {rowsData?.pagination.total || 0} records
            </div>
            <div className="flex items-center gap-2">
              <button 
                disabled={page === 0}
                onClick={() => setPage(p => p - 1)}
                className="px-4 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-600 shadow-sm hover:bg-slate-50 disabled:opacity-40 transition-all active:translate-y-px"
              >
                Previous
              </button>
              <div className="px-4 py-2 bg-slate-900 text-white rounded-lg text-xs font-bold shadow-lg shadow-slate-900/20">
                Page {page + 1}
              </div>
              <button 
                disabled={!rowsData?.pagination.has_more}
                onClick={() => setPage(p => p + 1)}
                className="px-4 py-2 bg-white border border-slate-200 rounded-lg text-xs font-bold text-slate-600 shadow-sm hover:bg-slate-50 disabled:opacity-40 transition-all active:translate-y-px"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export const Route = createFileRoute('/jobs/$jobId')({
  component: JobDetail,
});
