import { Badge } from "@/components/ui/badge";

const MOCK_USER = {
  name: "Blagoy Simandov",
  email: "blagoy@ampledata.io",
  plan: "Free",
  initials: "BS",
  memberSince: "January 2025",
};

function UserAvatar() {
  return (
    <div className="size-14 rounded-full bg-primary flex items-center justify-center shrink-0">
      <span className="text-lg font-black text-primary-foreground tracking-tight">
        {MOCK_USER.initials}
      </span>
    </div>
  );
}

function UserMeta() {
  return (
    <div className="flex flex-col gap-0.5 min-w-0">
      <h2 className="text-base font-black truncate">{MOCK_USER.name}</h2>
      <p className="text-xs text-muted-foreground truncate">{MOCK_USER.email}</p>
      <p className="text-xs text-muted-foreground">Member since {MOCK_USER.memberSince}</p>
    </div>
  );
}

export function ProfileCard() {
  return (
    <div className="flex items-center gap-4 px-4 py-4 rounded-lg bg-card ring-1 ring-foreground/10">
      <UserAvatar />
      <div className="flex-1 min-w-0">
        <UserMeta />
      </div>
      <Badge variant="outline">{MOCK_USER.plan}</Badge>
    </div>
  );
}
