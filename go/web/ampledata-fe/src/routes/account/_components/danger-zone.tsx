import { Button } from "@/components/ui/button";
import { SectionCard } from "./section-card";
import { ActionRow } from "./action-row";

function DeleteAction() {
  return (
    <Button variant="destructive" size="sm" disabled>
      Delete account
    </Button>
  );
}

export function DangerZone() {
  return (
    <SectionCard
      title="Danger Zone"
      description="Irreversible and destructive actions."
      className="ring-destructive/30"
    >
      <ActionRow
        label="Delete account"
        description="Permanently remove your account and all associated data."
        action={<DeleteAction />}
      />
    </SectionCard>
  );
}
