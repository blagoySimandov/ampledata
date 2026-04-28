import { Link } from "@tanstack/react-router";

interface NavLinkProps {
  to: string;
  children: React.ReactNode;
  exact?: boolean;
}

export function NavLink({ to, children, exact }: NavLinkProps) {
  return (
    <Link
      to={to}
      activeOptions={{ exact }}
      className="flex h-full items-center px-3 text-base font-bold tracking-wide text-muted-foreground border-b-2 border-b-transparent transition-colors hover:text-foreground data-[status=active]:text-primary data-[status=active]:border-b-primary"
    >
      {children}
    </Link>
  );
}
