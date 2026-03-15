import { ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SectionCard } from "./section-card";
import { ActionRow } from "./action-row";

function PasswordAction() {
  return (
    <Button variant="outline" size="sm" disabled>
      Change <ChevronRight />
    </Button>
  );
}

function TwoFactorAction() {
  return (
    <div className="flex items-center gap-2">
      <Badge variant="secondary">Off</Badge>
      <Button variant="outline" size="sm" disabled>
        Enable
      </Button>
    </div>
  );
}

export function SecuritySection() {
  return (
    <SectionCard
      title="Security"
      description="Manage your password and two-factor authentication."
    >
      <ActionRow
        label="Password"
        description="Update your account password."
        action={<PasswordAction />}
      />
      <div className="border-t" />
      <ActionRow
        label="Two-factor authentication"
        description="Add an extra layer of protection to your account."
        action={<TwoFactorAction />}
      />
    </SectionCard>
  );
}
