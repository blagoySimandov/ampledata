import { useAuthToken } from "@/hooks/use-auth-token";
import { ApiKeys, WorkOsWidgets } from "@workos-inc/widgets";
import { SectionCard } from "./section-card";

export function ApiKeysContent() {
  const { data: authToken } = useAuthToken();

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
        {authToken && <ApiKeys authToken={authToken} scope="user" />}
      </WorkOsWidgets>
    </SectionCard>
  );
}
