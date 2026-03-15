import { createFileRoute, Outlet } from "@tanstack/react-router";

function AccountLayout() {
  return (
    <div className="animate-in fade-in duration-500">
      <div className="mb-6">
        <h1 className="text-2xl font-black">Account</h1>
        <p className="text-xs text-muted-foreground mt-0.5">
          Manage your profile and subscription.
        </p>
      </div>
      <Outlet />
    </div>
  );
}

export const Route = createFileRoute("/account")({
  component: AccountLayout,
});
