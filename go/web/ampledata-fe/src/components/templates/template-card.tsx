import type { Template, TemplateColumnMetadata } from "@/api/types";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

const ENTITY_TYPE_STYLES: Record<
  string,
  { bg: string; text: string; dot: string }
> = {
  "Company Intel": {
    bg: "bg-primary/10",
    text: "text-primary",
    dot: "bg-primary",
  },
  People: {
    bg: "bg-blue-50 dark:bg-blue-950/30",
    text: "text-blue-700 dark:text-blue-400",
    dot: "bg-blue-500",
  },
  Research: {
    bg: "bg-violet-50 dark:bg-violet-950/30",
    text: "text-violet-700 dark:text-violet-400",
    dot: "bg-violet-500",
  },
  "Real Estate": {
    bg: "bg-emerald-50 dark:bg-emerald-950/30",
    text: "text-emerald-700 dark:text-emerald-400",
    dot: "bg-emerald-500",
  },
  "E-commerce": {
    bg: "bg-fuchsia-50 dark:bg-fuchsia-950/30",
    text: "text-fuchsia-700 dark:text-fuchsia-400",
    dot: "bg-fuchsia-500",
  },
  Healthcare: {
    bg: "bg-teal-50 dark:bg-teal-950/30",
    text: "text-teal-700 dark:text-teal-400",
    dot: "bg-teal-500",
  },
};

const DEFAULT_STYLE = {
  bg: "bg-muted",
  text: "text-muted-foreground",
  dot: "bg-muted-foreground",
};

const VISIBLE_COLUMN_COUNT = 4;

interface TemplateCardProps {
  template: Template;
  onApply: (template: Template) => void;
}

function EntityTypeBadge({ entityType }: { entityType: string }) {
  const style = ENTITY_TYPE_STYLES[entityType] ?? DEFAULT_STYLE;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-bold whitespace-nowrap",
        style.bg,
        style.text,
      )}
    >
      <span className={cn("size-1.5 shrink-0 rounded-full", style.dot)} />
      {entityType}
    </span>
  );
}

function ColumnPills({ columns }: { columns: TemplateColumnMetadata[] }) {
  const visible = columns.slice(0, VISIBLE_COLUMN_COUNT);
  const overflow = columns.length - VISIBLE_COLUMN_COUNT;
  return (
    <div className="flex flex-wrap gap-1.5">
      {visible.map((col) => (
        <span
          key={col.name}
          className="rounded-md bg-muted px-1.5 py-0.5 font-mono text-[11px] font-semibold text-muted-foreground"
        >
          {col.name}
        </span>
      ))}
      {overflow > 0 && (
        <span className="rounded-md bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
          +{overflow} more
        </span>
      )}
    </div>
  );
}

function ColumnsSection({ columns }: { columns: TemplateColumnMetadata[] }) {
  return (
    <div className="space-y-2">
      <p className="text-[9px] font-black uppercase tracking-widest text-muted-foreground">
        Columns · {columns.length}
      </p>
      <ColumnPills columns={columns} />
    </div>
  );
}

export function TemplateCard({ template, onApply }: TemplateCardProps) {
  return (
    <Card className="rounded-xl pb-0 transition-shadow duration-200 hover:shadow-lg cursor-pointer" onClick={() => onApply(template)}>
      <CardHeader>
        <CardTitle className="text-base font-black tracking-tight">
          {template.name}
        </CardTitle>
        <CardAction>
          <EntityTypeBadge entityType={template.entity_type} />
        </CardAction>
        <CardDescription className="line-clamp-2">{template.description}</CardDescription>
      </CardHeader>
      <CardContent>
        <ColumnsSection columns={template.columns_metadata} />
      </CardContent>
      <div className="mt-auto flex justify-end border-t border-border px-4 py-3">
        <Button
          onClick={(e) => { e.stopPropagation(); onApply(template); }}
          variant="outline"
          className="h-9 px-5 text-xs font-black tracking-widest text-muted-foreground hover:bg-primary hover:text-primary-foreground hover:border-primary group-hover/card:bg-primary group-hover/card:text-primary-foreground group-hover/card:border-primary"
        >
          APPLY
        </Button>
      </div>
    </Card>
  );
}
