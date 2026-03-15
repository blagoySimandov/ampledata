import { Link } from "@tanstack/react-router";
import { LogOut } from "lucide-react";
import { useAuth } from "@workos-inc/authkit-react";
import logo from "../../../assets/ampledata-logo.png";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { CreditsWidget } from "./credits-widget";

function NavLink({ to, children }: { to: string; children: React.ReactNode }) {
  return (
    <Link
      to={to}
      className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
      activeProps={{ className: "text-foreground font-semibold" }}
    >
      {children}
    </Link>
  );
}

function userInitials(
  firstName: string | null,
  lastName: string | null,
  email: string,
) {
  if (firstName && lastName) return `${firstName[0]}${lastName[0]}`.toUpperCase();
  if (firstName) return firstName[0].toUpperCase();
  return email[0].toUpperCase();
}

function userDisplayName(
  firstName: string | null,
  lastName: string | null,
  email: string,
) {
  if (firstName || lastName) return [firstName, lastName].filter(Boolean).join(" ");
  return email;
}

function UserMenu() {
  const { user, signOut } = useAuth();
  if (!user) return null;

  const initials = userInitials(user.firstName, user.lastName, user.email);
  const displayName = userDisplayName(user.firstName, user.lastName, user.email);

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
            onClick={() => signOut()}
          >
            <LogOut className="size-3.5 mr-2" />
            Sign out
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

export function Header() {
  return (
    <header className="sticky top-0 z-10 w-full border-b bg-background border-border shadow-sm">
      <div className="container mx-auto px-4 h-20 flex items-center justify-between">
        <Link to="/">
          <img src={logo} alt="AmpleData" className="h-12 w-auto" />
        </Link>
        <nav className="flex items-center gap-4">
          <NavLink to="/">Jobs</NavLink>
          <CreditsWidget />
          <UserMenu />
        </nav>
      </div>
    </header>
  );
}
