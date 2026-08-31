import { useQuery } from "@tanstack/react-query";

import { purchasesApi, type PurchaseFilters } from "../../api/purchases";

export const purchaseQueryKeys = {
  root: ["operations", "private", "purchases"] as const,
  lists: ["operations", "private", "purchases", "list"] as const,
  list: (filters: PurchaseFilters, cursor = "") =>
    [
      "operations",
      "private",
      "purchases",
      "list",
      filters.eventId,
      filters.status,
      cursor,
    ] as const,
  detail: (purchaseId: string) =>
    ["operations", "private", "purchases", "detail", purchaseId] as const,
};

export function usePurchases(
  filters: PurchaseFilters,
  cursor: string,
  enabled: boolean,
) {
  return useQuery({
    queryKey: purchaseQueryKeys.list(filters, cursor),
    queryFn: ({ signal }) =>
      purchasesApi.listPurchases({
        ...filters,
        cursor: cursor || undefined,
        limit: 10,
        signal,
      }),
    enabled,
    retry: false,
  });
}

export function usePurchase(purchaseId: string, enabled: boolean) {
  return useQuery({
    queryKey: purchaseQueryKeys.detail(purchaseId),
    queryFn: ({ signal }) => purchasesApi.getPurchase(purchaseId, { signal }),
    enabled: enabled && purchaseId !== "",
    retry: false,
  });
}
