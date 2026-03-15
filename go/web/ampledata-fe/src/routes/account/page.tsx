import { DangerZone } from "./_components/danger-zone";
import { PersonalInfoSection } from "./_components/personal-info-section";
import { ProfileCard } from "./_components/profile-card";
import { SecuritySection } from "./_components/security-section";

export function AccountPage() {
  return (
    <div className="max-w-2xl space-y-4">
      <ProfileCard />
      <PersonalInfoSection />
      <SecuritySection />
      <DangerZone />
    </div>
  );
}
