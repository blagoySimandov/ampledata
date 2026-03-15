import { createFileRoute, Link, Outlet } from "@tanstack/react-router";

const TAB_BASE =
  "pb-3 text-sm font-medium border-b-2 transition-colors -mb-px";
const TAB_INACTIVE =
  "text-muted-foreground border-transparent hover:text-foreground hover:border-border";
const TAB_ACTIVE = "text-foreground border-primary";

function AccountTabLink({
  to,
  exact,
  children,
}: {
  to: string;
  exact?: boolean;
  children: React.ReactNode;
}) {
  return (
    <Link
      to={to}
      activeOptions={{ exact }}
      className={`${TAB_BASE} ${TAB_INACTIVE}`}
      activeProps={{ className: `${TAB_BASE} ${TAB_ACTIVE}` }}
    >
      {children}
    </Link>
  );
}

function AccountHeader() {
  return (
    <div className="mb-4">
      <h1 className="text-2xl font-black">Account</h1>
      <p className="text-xs text-muted-foreground mt-0.5">
        Manage your profile and account settings.
      </p>
    </div>
  );
}

function AccountTabNav() {
  return (
    <nav className="flex gap-6 border-b mb-6">
      <AccountTabLink to="/account" exact>
        Profile
      </AccountTabLink>
      <AccountTabLink to="/account/billing">Billing</AccountTabLink>
    </nav>
  );
}

function AccountLayout() {
  return (
    <div className="animate-in fade-in duration-500">
      <AccountHeader />
      <AccountTabNav />
      <Outlet />
    </div>
  );
}

export const Route = createFileRoute("/account")({
  component: AccountLayout,
});
