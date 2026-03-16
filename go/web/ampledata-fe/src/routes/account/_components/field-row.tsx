interface Props {
  label: string;
  value: string;
  verified?: boolean;
}

export function FieldRow({ label, value }: Props) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col gap-0.5">
        <span className="text-xs font-medium text-muted-foreground">
          {label}
        </span>
        <span className="text-sm font-medium">{value}</span>
      </div>
    </div>
  );
}
