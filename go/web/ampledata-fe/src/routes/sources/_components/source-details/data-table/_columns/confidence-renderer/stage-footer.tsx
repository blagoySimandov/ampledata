export function StageFooter({ stage }: { stage: string }) {
  return (
    <div className="pt-2 border-t border-slate-100 flex items-center justify-between">
      <span className="text-xs font-bold text-slate-400 uppercase tracking-tight">
        Stage
      </span>
      <span className="text-xs font-black text-slate-900 uppercase tracking-widest">
        {stage}
      </span>
    </div>
  );
}
