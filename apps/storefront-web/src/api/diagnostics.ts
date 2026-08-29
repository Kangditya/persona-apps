import { useQuery } from "@tanstack/react-query";

import { storefrontApi } from "./client";

export const storefrontQueryKeys = {
  health: ["storefront", "diagnostics", "health"] as const,
};

export function useStorefrontHealth() {
  return useQuery({
    queryKey: storefrontQueryKeys.health,
    queryFn: ({ signal }) => storefrontApi.diagnostics.health({ signal }),
    staleTime: 30_000,
    retry: false,
  });
}
