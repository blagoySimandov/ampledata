import { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@workos-inc/authkit-react";
import { AUTH_ROUTES, UI_MESSAGES } from "@/constants";

interface ProtectedRouteProps {
  children: ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return <div>{UI_MESSAGES.LOADING}</div>;
  }

  if (!user) {
    return <Navigate to={AUTH_ROUTES.LOGIN} replace />;
  }

  return <>{children}</>;
}
