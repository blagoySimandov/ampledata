import type { UserResponse } from "@/api";
import { Badge } from "@/components/ui/badge";

interface Props {
  user: UserResponse | undefined;
  tierName: string | null | undefined;
}

function getInitials(first: string, last: string) {
  return `${first.charAt(0)}${last.charAt(0)}`.toUpperCase();
}

function UserAvatar({ first, last }: { first: string; last: string }) {
  return (
    <div className="size-14 rounded-full bg-primary flex items-center justify-center shrink-0">
      <span className="text-lg font-black text-primary-foreground tracking-tight">
        {getInitials(first, last)}
      </span>
    </div>
  );
}

function UserMeta({ user }: { user: UserResponse }) {
  return (
    <div className="flex flex-col gap-0.5 min-w-0">
      <h2 className="text-base font-black truncate">
        {user.first_name} {user.last_name}
      </h2>
      <p className="text-xs text-muted-foreground truncate">{user.email}</p>
    </div>
  );
}

export function ProfileCard({ user, tierName }: Props) {
  if (!user) return null;

  return (
    <div className="flex items-center gap-4 px-4 py-4 rounded-lg bg-card ring-1 ring-foreground/10">
      <UserAvatar first={user.first_name} last={user.last_name} />
      <div className="flex-1 min-w-0">
        <UserMeta user={user} />
      </div>
      <Badge variant="outline">{tierName ?? "Free"}</Badge>
    </div>
  );
}
