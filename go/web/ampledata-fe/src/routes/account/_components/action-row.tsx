interface Props {
  label: string;
  description?: string;
  action: React.ReactNode;
}

export function ActionRow({ label, description, action }: Props) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="flex flex-col gap-0.5 min-w-0">
        <span className="text-sm font-medium">{label}</span>
        {description && (
          <span className="text-xs text-muted-foreground">{description}</span>
        )}
      </div>
      <div className="shrink-0">{action}</div>
    </div>
  );
}
