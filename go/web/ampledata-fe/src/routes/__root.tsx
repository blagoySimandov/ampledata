import { createRootRoute, Outlet, Link } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/router-devtools';
import { Database } from 'lucide-react';

export const Route = createRootRoute({
  component: () => (
    <div className="min-h-screen bg-gray-50 flex flex-col font-sans">
      <header className="sticky top-0 z-10 w-full border-b bg-white border-gray-200 shadow-sm">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2 text-primary font-bold text-xl">
            <Database className="w-6 h-6" />
            <Link to="/">AmpleData</Link>
          </div>
          <nav>
            <Link
              to="/"
              className="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
              activeProps={{ className: 'text-primary font-semibold' }}
            >
              Jobs
            </Link>
          </nav>
        </div>
      </header>
      <main className="flex-1 container mx-auto p-4 py-8">
        <Outlet />
      </main>
      <TanStackRouterDevtools />
    </div>
  ),
});
