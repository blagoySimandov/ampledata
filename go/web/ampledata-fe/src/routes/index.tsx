import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useApi, useListJobs } from '../hooks';
import { useMemo } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { type ColDef, ModuleRegistry, AllCommunityModule, themeQuartz } from 'ag-grid-community';

ModuleRegistry.registerModules([AllCommunityModule]);

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
  headerBackgroundColor: '#f1f5f9',
  headerFontSize: '14px',
  headerFontWeight: '600',
  headerTextColor: '#475569',
  rowBorder: { color: '#f1f5f9' },
  wrapperBorder: true,
  wrapperBorderRadius: '8px',
});

function JobsList() {
  const api = useApi();
  const navigate = useNavigate();
  const { data, isLoading, isError, error } = useListJobs(api);

  const columnDefs = useMemo<ColDef[]>(() => [
    { 
      field: 'job_id', 
      headerName: 'Job Name', 
      flex: 1,
      cellRenderer: (params: any) => {
        const trimmed = params.value?.replace(/\.csv$/i, '') || params.value;
        return (
          <span className="font-semibold text-primary cursor-pointer hover:underline">
            {trimmed}
          </span>
        );
      },
      onCellClicked: (params) => {
        navigate({ to: '/jobs/$jobId', params: { jobId: params.data.job_id } });
      }
    },
    { 
      field: 'status', 
      headerName: 'Status',
      width: 140,
      cellRenderer: (params: any) => {
        const status = params.value;
        let colorClass = "";
        
        if (status === 'COMPLETED') colorClass = "bg-emerald-50 text-emerald-700 border-emerald-200";
        else if (status === 'RUNNING') colorClass = "bg-blue-50 text-blue-700 border-blue-200 animate-pulse";
        else if (status === 'FAILED' || status === 'CANCELLED') colorClass = "bg-red-50 text-red-700 border-red-200";
        else if (status === 'PAUSED') colorClass = "bg-amber-50 text-amber-700 border-amber-200";
        
        return (
          <div className="flex items-center h-full">
            <div className={`px-2 py-0.5 rounded-full border text-[10px] font-bold uppercase tracking-wider text-center leading-none ${colorClass}`}>
              {status}
            </div>
          </div>
        );
      }
    },
    { 
      field: 'total_rows', 
      headerName: 'Rows', 
      width: 100,
      cellRenderer: (params: any) => <span className="font-mono text-gray-600">{params.value}</span>
    },
    { 
      field: 'created_at', 
      headerName: 'Date Created', 
      flex: 1,
      valueFormatter: (params) => new Date(params.value).toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    },
  ], [navigate]);

  if (isLoading) {
    return <div className="text-center py-10 text-gray-500">Loading jobs...</div>;
  }

  if (isError) {
    return (
      <div className="bg-red-50 text-red-700 p-4 rounded-md">
        Failed to load jobs: {error?.message}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold tracking-tight">Enrichment Jobs</h1>
        {/* Placeholder for future action button */}
        {/* <button className="bg-primary text-primary-foreground hover:bg-primary/90 px-4 py-2 rounded-md font-medium text-sm">
          New Job
        </button> */}
      </div>

      <div className="w-full h-[600px] shadow-sm">
        <AgGridReact
          rowData={data?.jobs || []}
          columnDefs={columnDefs}
          theme={myTheme}
          pagination={true}
          paginationPageSize={10}
          paginationPageSizeSelector={[10, 20, 50]}
          domLayout="normal"
        />
      </div>
    </div>
  );
}

export const Route = createFileRoute('/')({
  component: JobsList,
});
