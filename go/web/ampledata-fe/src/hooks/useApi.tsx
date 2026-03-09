// src/hooks/useApi.tsx
import { createContext, useContext, useState, type ReactNode } from 'react';
import { ApiClient } from '../api';

const ApiContext = createContext<ApiClient | null>(null);

export function ApiProvider({ children }: { children: ReactNode }) {
  const [api] = useState(() => new ApiClient());
  return <ApiContext.Provider value={api}>{children}</ApiContext.Provider>;
}

export function useApi(): ApiClient {
  const context = useContext(ApiContext);
  if (!context) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return context;
}
