import { useAuth } from "@workos-inc/authkit-react";
import { useEffect } from "react";

export function Login() {
  const { signIn } = useAuth();

  useEffect(() => {
    signIn();
  }, [signIn]);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <p className="text-lg">Redirecting to sign in...</p>
      </div>
    </div>
  );
}
