import type { LucideIcon } from "lucide-react";
import { type BadgeVariant, BADGE_VARIANTS } from "./badge-variants";

export function CellBadge({
  variant,
  icon: Icon,
  spin = false,
  children,
}: {
  variant: BadgeVariant;
  icon: LucideIcon;
  spin?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div
      className={`flex items-center gap-1.5 px-2 py-0.5 rounded border text-xs font-bold uppercase tracking-tight w-fit ${BADGE_VARIANTS[variant]}`}
    >
      <Icon className={`w-3 h-3 ${spin ? "animate-spin" : ""}`} />
      {children}
    </div>
  );
}
