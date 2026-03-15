import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { userInitials, userDisplayName } from "@/lib/utils";
import { useAuth } from "@workos-inc/authkit-react";
import { Link, LogOut } from "lucide-react";

export function UserMenu() {
  const { user, signOut } = useAuth();
  if (!user) return null;

  const initials = userInitials(user.firstName, user.lastName, user.email);
  const displayName = userDisplayName(
    user.firstName,
    user.lastName,
    user.email,
  );

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          aria-label="User menu"
          className="size-9 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-bold hover:opacity-90 transition-opacity focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {initials}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-56 p-2">
        <div className="px-2 py-1.5 mb-1">
          <p className="text-sm font-semibold truncate">{displayName}</p>
          <p className="text-xs text-muted-foreground truncate">{user.email}</p>
        </div>
        <div className="border-t pt-1">
          <Link to="/account" className="block">
            <Button
              variant="ghost"
              size="sm"
              className="w-full justify-start text-sm"
            >
              Account settings
            </Button>
          </Link>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start text-sm text-destructive hover:text-destructive"
            onClick={() => signOut({ returnTo: "/" })}
          >
            <LogOut className="size-3.5 mr-2" />
            Sign out
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
