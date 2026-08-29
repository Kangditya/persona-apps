import { ApiError } from "@persona-apps/api-client";
import { useQuery } from "@tanstack/react-query";

import { catalogueApi } from "../../api/catalogue";

export const catalogueQueryKeys = {
  root: ["storefront", "catalogue"] as const,
  activeEvent: ["storefront", "catalogue", "active-event"] as const,
  offerings: (eventId: string) =>
    ["storefront", "catalogue", "events", eventId, "offerings"] as const,
  offering: (offeringId: string) =>
    ["storefront", "catalogue", "offerings", offeringId] as const,
};

export function retryPublicRead(failureCount: number, error: unknown): boolean {
  return (
    failureCount < 2 &&
    error instanceof ApiError &&
    (error.kind === "service_unavailable" || error.kind === "unavailable")
  );
}

export function useActiveEvent() {
  return useQuery({
    queryKey: catalogueQueryKeys.activeEvent,
    queryFn: ({ signal }) => catalogueApi.getActiveEvent({ signal }),
    retry: retryPublicRead,
  });
}

export function usePublicOfferings(eventId: string) {
  return useQuery({
    queryKey: catalogueQueryKeys.offerings(eventId),
    queryFn: ({ signal }) => catalogueApi.listOfferings(eventId, { signal }),
    enabled: eventId !== "",
    retry: retryPublicRead,
  });
}

export function usePublicOffering(offeringId: string, enabled: boolean) {
  return useQuery({
    queryKey: catalogueQueryKeys.offering(offeringId),
    queryFn: ({ signal }) => catalogueApi.getOffering(offeringId, { signal }),
    enabled,
    retry: retryPublicRead,
  });
}
