import { Link } from "@tanstack/react-router";

export function NavLink({
  to,
  children,
}: {
  to: string;
  children: React.ReactNode;
}) {
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
