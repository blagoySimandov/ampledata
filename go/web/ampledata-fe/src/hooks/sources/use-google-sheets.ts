import { useMutation, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import type { CreateGoogleSheetsSourceRequest } from "@/api/types";
import { useApi } from "@/hooks/use-api";

export function useOAuthStatus() {
  const api = useApi();
  return useQuery({
    queryKey: ["oauth", "google", "status"],
    queryFn: () => api.getOAuthStatus(),
  });
}

export function useInitiateGoogleOAuth() {
  const api = useApi();
  return useMutation({
    mutationFn: () => api.initiateGoogleOAuth(),
    onSuccess: ({ url }) => {
      window.location.href = url;
    },
  });
}

export function useListSpreadsheets(enabled: boolean) {
  const api = useApi();
  return useQuery({
    queryKey: ["google-sheets", "spreadsheets"],
    queryFn: () => api.listSpreadsheets(),
    enabled,
  });
}

export function useListSheetTabs(spreadsheetId: string | null) {
  const api = useApi();
  return useQuery({
    queryKey: ["google-sheets", "tabs", spreadsheetId],
    queryFn: () => api.listSheetTabs(spreadsheetId!),
    enabled: !!spreadsheetId,
  });
}

export function useCreateGoogleSheetsSource() {
  const api = useApi();
  const navigate = useNavigate();
  return useMutation({
    mutationFn: (req: CreateGoogleSheetsSourceRequest) =>
      api.createGoogleSheetsSource(req),
    onSuccess: ({ source_id }) => {
      navigate({ to: "/sources/$sourceId", params: { sourceId: source_id } });
    },
  });
}
