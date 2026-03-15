/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useMemo, type ReactNode } from "react";
import { useAuth } from "@workos-inc/authkit-react";
import { ApiClient } from "../api";

const ApiContext = createContext<ApiClient | null>(null);

export function ApiProvider({ children }: { children: ReactNode }) {
  const { getAccessToken, user } = useAuth();
  console.log(user);
  const api = useMemo(() => new ApiClient(getAccessToken), [getAccessToken]);
  return <ApiContext.Provider value={api}>{children}</ApiContext.Provider>;
}

export function useApi(): ApiClient {
  const context = useContext(ApiContext);
  if (!context) throw new Error("useApi must be used within an ApiProvider");
  return context;
}
