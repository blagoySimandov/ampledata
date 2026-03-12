import { Link } from "@tanstack/react-router";
import logo from "../../../assets/ampledata-logo.png";

export function Header() {
  return (
    <header className="sticky top-0 z-10 w-full border-b bg-white border-gray-200 shadow-sm">
      <div className="container mx-auto px-4 h-20 flex items-center justify-between">
        <Link to="/">
          <img src={logo} alt="AmpleData" className="h-12 w-auto" />
        </Link>
        <nav>
          <Link
            to="/"
            className="text-base font-medium text-gray-600 hover:text-gray-900 transition-colors"
            activeProps={{ className: "text-primary font-semibold" }}
          >
            Jobs
          </Link>
        </nav>
      </div>
    </header>
  );
}
