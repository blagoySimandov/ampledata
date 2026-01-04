import { useAuth } from "@workos-inc/authkit-react";
import { Navigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { APP_ROUTES, UI_MESSAGES } from "@/constants";

export function Login() {
  const { user, isLoading, signIn } = useAuth();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div>{UI_MESSAGES.LOADING}</div>
      </div>
    );
  }

  if (user) {
    return <Navigate to={APP_ROUTES.HOME} replace />;
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-background">
      <Card className="p-8 w-full max-w-md">
        <div className="space-y-6">
          <div className="space-y-2 text-center">
            <h1 className="text-3xl font-bold">AmpleData</h1>
            <p className="text-muted-foreground">
              Sign in to access your account
            </p>
          </div>
          <Button onClick={() => signIn()} className="w-full">
            {UI_MESSAGES.SIGN_IN}
          </Button>
        </div>
      </Card>
    </div>
  );
}
