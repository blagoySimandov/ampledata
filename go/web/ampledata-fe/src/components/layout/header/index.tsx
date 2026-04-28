import { Link } from "@tanstack/react-router";
import logo from "../../../../assets/ampledata-logo.png";
import { CreditsWidget } from "./credits-widget";
import { UserMenu } from "./user-menu";
import { NavLink } from "./nav-link";

export function Header() {
  return (
    <header className="sticky top-0 z-20 w-full border-b border-border bg-background">
      <div className="flex h-18 items-center justify-between px-7">
        <div className="flex h-full items-center gap-8">
          <Link to="/app">
            <img src={logo} alt="AmpleData" className="h-10 w-auto" />
          </Link>
          <nav className="flex h-full items-center gap-0.5">
            <NavLink to="/app" exact>Sources</NavLink>
            <NavLink to="/templates">Templates</NavLink>
          </nav>
        </div>
        <div className="flex items-center gap-3">
          <CreditsWidget />
          <UserMenu />
        </div>
      </div>
    </header>
  );
}
