import type { Template } from "@/api/types";
import { TemplateCard } from "@/components/templates/template-card";
import { Input } from "@/components/ui/input";
import { useListTemplates } from "@/hooks/templates";
import { useApi } from "@/hooks/use-api";
import { cn } from "@/lib/utils";
import {
  LayoutGrid,
  LayoutList,
  Plus,
  Search,
  Sparkles,
  Bookmark,
} from "lucide-react";
import { useState, useMemo } from "react";
//TODO: Split into multiple files

const ALL_FILTER = "All";
const SAVED_FILTER = "Saved";

function useTemplateFilters(templates: Template[]) {
  const [query, setQuery] = useState("");
  const [activeFilter, setActiveFilter] = useState(ALL_FILTER);

  const entityTypes = useMemo(
    () => [...new Set(templates.map((t) => t.entity_type))],
    [templates],
  );

  const filters = [ALL_FILTER, ...entityTypes, SAVED_FILTER];

  const filtered = useMemo(() => {
    return templates.filter((t) => {
      const matchFilter =
        activeFilter === ALL_FILTER ||
        (activeFilter === SAVED_FILTER && t.type === "user_defined_template") ||
        t.entity_type === activeFilter;
      const matchQuery =
        !query ||
        t.name.toLowerCase().includes(query.toLowerCase()) ||
        t.description.toLowerCase().includes(query.toLowerCase());
      return matchFilter && matchQuery;
    });
  }, [templates, activeFilter, query]);

  return { query, setQuery, activeFilter, setActiveFilter, filters, filtered };
}

function FilterChip({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-full px-3 py-1 text-xs font-bold border transition-colors",
        active
          ? "bg-foreground text-background border-foreground"
          : "bg-background text-muted-foreground border-border hover:border-foreground/30 hover:text-foreground",
      )}
    >
      {label}
    </button>
  );
}

function ViewToggle({
  mode,
  onChange,
}: {
  mode: "grid" | "list";
  onChange: (m: "grid" | "list") => void;
}) {
  return (
    <div className="flex overflow-hidden rounded-lg border border-border">
      {(["grid", "list"] as const).map((m) => (
        <button
          key={m}
          onClick={() => onChange(m)}
          className={cn(
            "flex size-9 items-center justify-center transition-colors",
            mode === m
              ? "bg-foreground text-background"
              : "bg-background text-muted-foreground hover:text-foreground",
          )}
        >
          {m === "grid" ? (
            <LayoutGrid className="size-3.5" />
          ) : (
            <LayoutList className="size-3.5" />
          )}
        </button>
      ))}
    </div>
  );
}

function FilterBar({
  query,
  onQueryChange,
  filters,
  activeFilter,
  onFilterChange,
  viewMode,
  onViewModeChange,
}: {
  query: string;
  onQueryChange: (q: string) => void;
  filters: string[];
  activeFilter: string;
  onFilterChange: (f: string) => void;
  viewMode: "grid" | "list";
  onViewModeChange: (m: "grid" | "list") => void;
}) {
  return (
    <div className="mb-7 flex flex-wrap items-center gap-3">
      <div className="relative w-60">
        <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          placeholder="Search templates…"
          className="h-9 pl-8 text-sm"
        />
      </div>
      <div className="flex flex-wrap gap-1.5">
        {filters.map((f) => (
          <FilterChip
            key={f}
            label={f}
            active={activeFilter === f}
            onClick={() => onFilterChange(f)}
          />
        ))}
      </div>
      <div className="ml-auto">
        <ViewToggle mode={viewMode} onChange={onViewModeChange} />
      </div>
    </div>
  );
}

function SectionLabel({
  icon,
  label,
}: {
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <div className="mb-4 flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
      {icon}
      {label}
    </div>
  );
}

function TemplatesGrid({
  templates,
  onApply,
}: {
  templates: Template[];
  onApply: (t: Template) => void;
}) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {templates.map((t) => (
        <TemplateCard key={t.id} template={t} onApply={onApply} />
      ))}
    </div>
  );
}

function TemplatesList({
  templates,
  onApply,
}: {
  templates: Template[];
  onApply: (t: Template) => void;
}) {
  return (
    <div className="flex flex-col gap-0 overflow-hidden rounded-xl border border-border bg-card">
      {templates.map((t) => (
        <div
          key={t.id}
          className="flex items-center justify-between gap-4 border-b border-border px-5 py-3.5 last:border-b-0 hover:bg-muted/30 transition-colors"
        >
          <div className="min-w-0">
            <p className="text-sm font-black text-foreground">{t.name}</p>
            <p className="mt-0.5 truncate text-xs text-muted-foreground">
              {t.description}
            </p>
          </div>
          <div className="flex shrink-0 items-center gap-4">
            <span className="text-xs text-muted-foreground">
              {t.columns_metadata.length} columns
            </span>
            <button
              onClick={() => onApply(t)}
              className="rounded-lg border border-border px-3 py-1.5 text-xs font-black tracking-widest text-muted-foreground transition-colors hover:border-primary hover:bg-primary hover:text-primary-foreground"
            >
              APPLY
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}

function TemplatesSection({
  icon,
  label,
  templates,
  viewMode,
  onApply,
}: {
  icon: React.ReactNode;
  label: string;
  templates: Template[];
  viewMode: "grid" | "list";
  onApply: (t: Template) => void;
}) {
  if (templates.length === 0) return null;
  return (
    <div className="mb-10">
      <SectionLabel icon={icon} label={label} />
      {viewMode === "grid" ? (
        <TemplatesGrid templates={templates} onApply={onApply} />
      ) : (
        <TemplatesList templates={templates} onApply={onApply} />
      )}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="py-20 text-center">
      <p className="text-sm font-bold text-muted-foreground">
        No templates found
      </p>
      <p className="mt-1 text-xs text-muted-foreground">
        Try a different search or category
      </p>
    </div>
  );
}

function GalleryHeader() {
  return (
    <div className="mb-7 flex items-start justify-between">
      <div>
        <h1 className="text-2xl font-black tracking-tight">
          Recipes &amp; Templates
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Predefined column sets and AI prompts. Apply a template to start
          enriching in one click.
        </p>
      </div>
      <button className="flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2.5 text-xs font-black tracking-widest text-primary-foreground transition-colors hover:bg-primary/90">
        <Plus className="size-3.5" />
        NEW TEMPLATE
      </button>
    </div>
  );
}

export function TemplatesPage() {
  const api = useApi();
  const { data, isLoading, isError } = useListTemplates(api);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  const templates = data?.templates ?? [];
  const { query, setQuery, activeFilter, setActiveFilter, filters, filtered } =
    useTemplateFilters(templates);

  const systemTemplates = filtered.filter((t) => t.type === "system_template");
  const userTemplates = filtered.filter(
    (t) => t.type === "user_defined_template",
  );

  function handleApply(template: Template) {
    console.log("Apply template", template.id);
  }

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center text-muted-foreground text-sm">
        Loading templates…
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex h-64 items-center justify-center text-destructive text-sm">
        Failed to load templates.
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl px-7 py-8">
      <GalleryHeader />
      <FilterBar
        query={query}
        onQueryChange={setQuery}
        filters={filters}
        activeFilter={activeFilter}
        onFilterChange={setActiveFilter}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
      />
      <TemplatesSection
        icon={<Sparkles className="size-3" />}
        label="AmpleData Templates"
        templates={systemTemplates}
        viewMode={viewMode}
        onApply={handleApply}
      />
      <TemplatesSection
        icon={<Bookmark className="size-3" />}
        label="Your Saved Templates"
        templates={userTemplates}
        viewMode={viewMode}
        onApply={handleApply}
      />
      {filtered.length === 0 && <EmptyState />}
    </div>
  );
}
