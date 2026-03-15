import { Lock } from "lucide-react";
import { SectionCard } from "./section-card";
import { FieldRow } from "./field-row";

const MOCK_FIELDS = [
  { label: "Full name", value: "Blagoy Simandov" },
  { label: "Email address", value: "blagoy@ampledata.io", verified: true },
] as const;

function WorkOsNote() {
  return (
    <div className="flex items-center gap-1.5 pt-2 border-t text-xs text-muted-foreground">
      <Lock className="size-3 shrink-0" />
      <span>Profile editing will be available once WorkOS is integrated.</span>
    </div>
  );
}

export function PersonalInfoSection() {
  return (
    <SectionCard
      title="Personal Information"
      description="Your name and email address on this account."
    >
      <div className="space-y-3">
        {MOCK_FIELDS.map((f) => (
          <FieldRow key={f.label} {...f} />
        ))}
      </div>
      <WorkOsNote />
    </SectionCard>
  );
}
