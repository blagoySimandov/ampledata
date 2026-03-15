import { Link } from "@tanstack/react-router";
import { UserRound } from "lucide-react";
import logo from "../../../assets/ampledata-logo.png";
import { Button } from "@/components/ui/button";
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
          <Button variant="ghost" size="icon" asChild>
            <Link to="/account" aria-label="Account">
              <UserRound className="size-5" />
            </Link>
          </Button>
        </nav>
      </div>
    </header>
  );
}
