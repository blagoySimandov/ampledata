import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { 
  useApi, 
  useListJobs, 
  useGetSignedUrl, 
  useUploadFile, 
  useSelectKey, 
  useStartJob 
} from '../hooks';
import { useMemo, useState, useRef } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { type ColDef, ModuleRegistry, AllCommunityModule, themeQuartz } from 'ag-grid-community';
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogTrigger
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Plus, Upload, Loader2, CheckCircle2, ChevronRight, Trash2, Settings2 } from 'lucide-react';
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { ColumnMetadata } from '../api';

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

function NewJobDialog() {
  const api = useApi();
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<'upload' | 'configure' | 'starting'>('upload');
  const [file, setFile] = useState<File | null>(null);
  const [jobId, setJobId] = useState<string | null>(null);
  const [allKeys, setAllKeys] = useState<string[]>([]);
  const [keyColumnDescription, setKeyColumnDescription] = useState<string>("");
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [columnsMetadata, setColumnsMetadata] = useState<ColumnMetadata[]>([]);
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  const getSignedUrl = useGetSignedUrl(api);
  const uploadFile = useUploadFile(api);
  const selectKey = useSelectKey(api);
  const startJob = useStartJob(api);

  const navigate = useNavigate();

  const reset = () => {
    setStep('upload');
    setFile(null);
    setJobId(null);
    setAllKeys([]);
    setSelectedKeys([]);
    setColumnsMetadata([]);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
    }
  };

  const handleUpload = async () => {
    if (!file) return;
    
    try {
      const { url, jobId } = await getSignedUrl.mutateAsync({
        contentType: 'text/csv',
        length: file.size
      });
      setJobId(jobId);

      await uploadFile.mutateAsync({ url, file });

      const { selected_key, all_keys } = await selectKey.mutateAsync({ job_id: jobId });
      setAllKeys(all_keys);
      setSelectedKeys([selected_key]);
      
      setStep('configure');
    } catch (error) {
      console.error("Upload failed", error);
    }
  };

  const addColumn = () => {
    setColumnsMetadata([...columnsMetadata, { name: "", type: "string", job_type: "enrichment" }]);
  };

  const removeColumn = (index: number) => {
    setColumnsMetadata(columnsMetadata.filter((_, i) => i !== index));
  };

  const updateColumn = (index: number, updates: Partial<ColumnMetadata>) => {
    const updated = [...columnsMetadata];
    updated[index] = { ...updated[index], ...updates };
    setColumnsMetadata(updated);
  };

  const handleStart = async () => {
    if (!jobId) return;
    
    setStep('starting');
    try {
      // Store column names in localStorage so the job page can immediately render the table headers
      const columnNames = columnsMetadata.map(c => c.name);
      localStorage.setItem(`job_columns_${jobId}`, JSON.stringify(columnNames));

      await startJob.mutateAsync({
        jobId,
        req: {
          key_columns: selectedKeys,
          columns_metadata: columnsMetadata,
          key_column_description: keyColumnDescription
        }
      });
      setOpen(false);
      reset();
      navigate({ to: '/jobs/$jobId', params: { jobId } });
    } catch (error) {
      console.error("Start failed", error);
      setStep('configure');
    }
  };

  return (
    <Dialog open={open} onOpenChange={(val) => { setOpen(val); if (!val) reset(); }}>
      <DialogTrigger asChild>
        <Button className="font-bold gap-2">
          <Plus className="w-4 h-4" />
          NEW ENRICHMENT
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="p-6 pb-0">
          <DialogTitle className="text-2xl font-black flex items-center gap-2">
            {step === 'upload' ? 'Upload Dataset' : 'Configure Enrichment'}
          </DialogTitle>
        </DialogHeader>

        <ScrollArea className="flex-1 px-6">
          <div className="py-6">
            {step === 'upload' && (
              <div className="space-y-4">
                <div 
                  className="border-2 border-dashed border-slate-200 rounded-2xl p-10 flex flex-col items-center justify-center gap-4 hover:border-primary/50 transition-colors cursor-pointer bg-slate-50/50"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <input 
                    type="file" 
                    ref={fileInputRef} 
                    onChange={handleFileChange} 
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
                  disabled={!file || getSignedUrl.isPending || uploadFile.isPending || selectKey.isPending}
                  onClick={handleUpload}
                >
                  {(getSignedUrl.isPending || uploadFile.isPending || selectKey.isPending) ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin mr-2" />
                      PROCESSING...
                    </>
                  ) : (
                    <>
                      CONTINUE
                      <ChevronRight className="w-4 h-4 ml-2" />
                    </>
                  )}
                </Button>
              </div>
            )}

            {step === 'configure' && (
              <div className="space-y-8">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-xs font-black uppercase tracking-widest text-slate-400">Key Column Description</Label>
                    <Input 
                      value={keyColumnDescription}
                      onChange={(e) => setKeyColumnDescription(e.target.value)}
                      className="h-10 rounded-xl"
                      placeholder="e.g. A list of technology companies"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label className="text-xs font-black uppercase tracking-widest text-slate-400">Key Columns</Label>
                    <div className="flex flex-wrap gap-1">
                      {allKeys.map(k => (
                        <Badge 
                          key={k}
                          variant={selectedKeys.includes(k) ? "default" : "outline"}
                          className={`cursor-pointer px-2 py-0 text-xs font-bold ${selectedKeys.includes(k) ? '' : 'text-slate-500 hover:bg-slate-50'}`}
                          onClick={() => {
                            if (selectedKeys.includes(k)) {
                              setSelectedKeys(selectedKeys.filter(x => x !== k));
                            } else {
                              setSelectedKeys([...selectedKeys, k]);
                            }
                          }}
                        >
                          {k}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs font-black uppercase tracking-widest text-slate-400 flex items-center gap-2">
                      <Settings2 className="w-3 h-3" />
                      Enrichment Fields
                    </Label>
                    <Button variant="ghost" size="sm" onClick={addColumn} className="h-7 text-xs font-black px-2 hover:bg-slate-100">
                      <Plus className="w-3 h-3 mr-1" /> ADD FIELD
                    </Button>
                  </div>
                  
                  <div className="space-y-3">
                    {columnsMetadata.length > 0 ? (
                      columnsMetadata.map((col, index) => (
                        <div key={index} className="flex flex-col gap-2 p-3 bg-white border border-slate-100 rounded-xl shadow-sm animate-in slide-in-from-left-2 duration-200">
                          <div className="flex items-center gap-2">
                            <Input 
                              placeholder="Field name" 
                              value={col.name} 
                              onChange={(e) => updateColumn(index, { name: e.target.value })}
                              className="h-9 text-xs font-medium"
                            />
                            <Select 
                              value={col.job_type} 
                              onValueChange={(v: "enrichment" | "imputation") => updateColumn(index, { job_type: v })}
                            >
                              <SelectTrigger className="h-9 w-[120px] text-xs font-medium bg-slate-50 border-transparent">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="enrichment">Enrich</SelectItem>
                                <SelectItem value="imputation">Impute</SelectItem>
                              </SelectContent>
                            </Select>
                            <Select 
                              value={col.type} 
                              onValueChange={(v: "string" | "number" | "boolean" | "date") => updateColumn(index, { type: v })}
                            >
                              <SelectTrigger className="h-9 w-[100px] text-xs font-medium bg-slate-50 border-transparent">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="string">String</SelectItem>
                                <SelectItem value="number">Number</SelectItem>
                                <SelectItem value="boolean">Bool</SelectItem>
                                <SelectItem value="date">Date</SelectItem>
                              </SelectContent>
                            </Select>
                            <Button variant="ghost" size="icon" onClick={() => removeColumn(index)} className="h-9 w-9 text-slate-400 hover:text-red-500 hover:bg-red-50 shrink-0">
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </div>
                          <Input
                            placeholder="Optional instructions for the AI (e.g. 'Extract the person's current job title')"
                            value={col.description || ''}
                            onChange={(e) => updateColumn(index, { description: e.target.value })}
                            className="h-8 text-xs bg-slate-50/50 border-slate-100 placeholder:text-slate-400"
                          />
                        </div>
                      ))
                    ) : (
                      <div className="bg-slate-50 border border-slate-100 rounded-2xl p-6 flex flex-col items-center text-center gap-3 animate-in fade-in zoom-in-95 duration-300">
                        <div className="w-10 h-10 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100">
                          <Settings2 className="w-5 h-5 text-slate-400" />
                        </div>
                        <div className="space-y-1">
                          <p className="text-sm font-bold text-slate-900">Define enrichment fields</p>
                          <p className="text-xs text-slate-500 max-w-[280px] leading-relaxed">
                            Add the fields you want the AI to extract (Enrich) or fill in (Impute) from your data.
                          </p>
                        </div>
                        <Button variant="outline" size="sm" onClick={addColumn} className="mt-1 font-bold h-8 px-4 border-slate-200">
                          <Plus className="w-3 h-3 mr-2" />
                          ADD YOUR FIRST FIELD
                        </Button>
                      </div>
                    )}
                  </div>
                </div>

                <div className="bg-slate-900 text-white p-4 rounded-2xl flex gap-3 items-center">
                  <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
                  <p className="text-xs font-medium leading-relaxed">
                    AI will process your data and fill in the selected fields using high-confidence sources.
                  </p>
                </div>

                <Button 
                  className="w-full font-black h-12 bg-primary hover:bg-primary/90" 
                  onClick={handleStart}
                  disabled={selectedKeys.length === 0 || columnsMetadata.length === 0 || columnsMetadata.some(c => !c.name)}
                >
                  START ENRICHMENT JOB
                </Button>
              </div>
            )}

            {step === 'starting' && (
              <div className="py-12 flex flex-col items-center justify-center gap-4 animate-in fade-in duration-500">
                 <Loader2 className="w-10 h-10 text-primary animate-spin" />
                 <p className="font-black uppercase tracking-widest text-slate-400 text-xs">Initializing Job...</p>
              </div>
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

function JobsList() {
  const api = useApi();
  const navigate = useNavigate();
  const { data, isLoading, isError, error } = useListJobs(api);

  const columnDefs = useMemo<ColDef[]>(() => [
    { 
      field: 'job_id', 
      headerName: 'Job Name', 
      flex: 1,
      cellRenderer: (params: { value: string; data: { job_id: string } }) => {
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
      cellRenderer: (params: { value: string; data: { job_id: string } }) => {
        const status = params.value;
        let colorClass = "";
        
        if (status === 'COMPLETED') colorClass = "bg-emerald-50 text-emerald-700 border-emerald-200";
        else if (status === 'RUNNING') colorClass = "bg-blue-50 text-blue-700 border-blue-200 animate-pulse";
        else if (status === 'FAILED' || status === 'CANCELLED') colorClass = "bg-red-50 text-red-700 border-red-200";
        else if (status === 'PAUSED') colorClass = "bg-amber-50 text-amber-700 border-amber-200";
        
        return (
          <div className="flex items-center h-full">
            <div className={`px-2 py-0.5 rounded-full border text-xs font-bold uppercase tracking-wider text-center leading-none ${colorClass}`}>
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
      cellRenderer: (params: { value: number }) => <span className="font-mono text-gray-600">{params.value}</span>
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

  const hasJobs = data?.jobs && data.jobs.length > 0;

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-black tracking-tight text-slate-900">Enrichment Jobs</h1>
        <NewJobDialog />
      </div>

      {!hasJobs ? (
        <div className="border-2 border-dashed border-slate-200 rounded-3xl p-16 flex flex-col items-center justify-center text-center bg-slate-50/50">
          <div className="w-20 h-20 rounded-full bg-white shadow-sm flex items-center justify-center border border-slate-100 mb-6">
            <Settings2 className="w-10 h-10 text-slate-400" />
          </div>
          <h2 className="text-2xl font-black text-slate-900 mb-2">No Jobs Found</h2>
          <p className="text-slate-500 max-w-md mb-8">
            You haven't run any enrichment jobs yet. Upload your first CSV to start enriching your data with high-confidence sources.
          </p>
          <div className="mt-2">
             <NewJobDialog />
          </div>
        </div>
      ) : (
        <div className="w-full h-[600px] bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden">
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
      )}
    </div>
  );
}

export const Route = createFileRoute('/')({
  component: JobsList,
});
