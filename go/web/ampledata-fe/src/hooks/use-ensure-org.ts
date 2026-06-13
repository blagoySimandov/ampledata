import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@workos-inc/authkit-react";
import { useApi } from "./use-api";

const STAGING_ORG_ID = "org_01KE21C5J2552WY8B64GBXR1NK";
const PRODUCTION_ORG_ID = "org_01KS087ZBHZ5WZSDW3QZMJTHNF";
const GLOBAL_ORG_ID = import.meta.env.PROD ? PRODUCTION_ORG_ID : STAGING_ORG_ID;

export function useEnsureGlobalOrg() {
  const { user, organizationId, switchToOrganization } = useAuth();
  const api = useApi();
  return useQuery({
    queryKey: ["ensure-org", user?.id],
    enabled: !!user && !organizationId,
    retry: false,
    queryFn: async () => {
      await api.getMe();
      await switchToOrganization({ organizationId: GLOBAL_ORG_ID });
      return true;
    },
  });
}
