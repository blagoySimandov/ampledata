import { useAuthToken } from "@/hooks/use-auth-token";
import { useAuth } from "@workos-inc/authkit-react";
import { ApiKeys, WorkOsWidgets } from "@workos-inc/widgets";

export function ApiKeysContent() {
  const { isLoading, user } = useAuth();
  const { data: authToken } = useAuthToken();

  if (isLoading) {
    return "...";
  }
  if (!user) {
    return "Logged in user is required";
  }

  return (
    <WorkOsWidgets theme={{ accentColor: "orange" }}>
      {authToken && <ApiKeys authToken={authToken} scope="user" />}
    </WorkOsWidgets>
  );
}
