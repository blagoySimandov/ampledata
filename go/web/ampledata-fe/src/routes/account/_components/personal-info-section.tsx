import type { UserResponse } from "@/api";
import { SectionCard } from "./section-card";
import { FieldRow } from "./field-row";

interface Props {
  user: UserResponse | undefined;
}

export function PersonalInfoSection({ user }: Props) {
  if (!user) return null;

  return (
    <SectionCard
      title="Personal Information"
      description="Your name and email address on this account."
    >
      <div className="space-y-3">
        <FieldRow label="Full name" value={`${user.first_name} ${user.last_name}`} />
        <FieldRow label="Email address" value={user.email} verified />
      </div>
    </SectionCard>
  );
}
