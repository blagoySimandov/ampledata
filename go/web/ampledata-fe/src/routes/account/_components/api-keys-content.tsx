import { useAuthToken } from "@/hooks/use-auth-token";
import { useAuth } from "@workos-inc/authkit-react";
import { ApiKeys, WorkOsWidgets } from "@workos-inc/widgets";
import { SectionCard } from "./section-card";

export function ApiKeysContent() {
  const { data: authToken } = useAuthToken();
  const { organizationId } = useAuth();

  return (
    <SectionCard
      title="API Keys"
      description="Authenticate with the API using secret keys."
    >
      <WorkOsWidgets
        theme={{
          accentColor: "orange",
          grayColor: "sand",
          radius: "large",
          panelBackground: "solid",
          scaling: "100%",
          fontFamily: '"Figtree Variable", sans-serif',
        }}
      >
        {authToken && organizationId && (
          <ApiKeys authToken={authToken} scope="user" />
        )}
      </WorkOsWidgets>
    </SectionCard>
  );
}
