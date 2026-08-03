import { useQuery } from "@tanstack/react-query";

import { operationsApi } from "./client";

export const operationsQueryKeys = {
  health: ["operations", "diagnostics", "health"] as const,
};

export function useOperationsHealth() {
  return useQuery({
    queryKey: operationsQueryKeys.health,
    queryFn: ({ signal }) => operationsApi.diagnostics.health({ signal }),
    staleTime: 30_000,
    retry: false,
  });
}
