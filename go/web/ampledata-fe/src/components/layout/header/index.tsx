import { Link } from "@tanstack/react-router";
import logo from "../../../../assets/ampledata-logo.png";
import { CreditsWidget } from "./credits-widget";
import { UserMenu } from "./user-menu";
import { NavLink } from "./nav-link";

export function Header() {
  return (
    <header className="sticky top-0 z-10 w-full border-b bg-background border-border shadow-sm">
      <div className="container mx-auto px-4 h-20 flex items-center justify-between">
        <Link to="/app">
          <img src={logo} alt="AmpleData" className="h-12 w-auto" />
        </Link>
        <nav className="flex items-center gap-4">
          <NavLink to="/app">Jobs</NavLink>
          <CreditsWidget />
          <UserMenu />
        </nav>
      </div>
    </header>
  );
}
