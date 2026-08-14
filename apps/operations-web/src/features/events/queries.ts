import { useQuery } from "@tanstack/react-query";

import { eventsApi } from "../../api/events";
import { sessionApi } from "../../api/session";

export const operationsQueryKeys = {
  privateRoot: ["operations", "private"] as const,
  session: ["operations", "private", "session"] as const,
  events: ["operations", "private", "events"] as const,
  eventLists: ["operations", "private", "events", "list"] as const,
  eventList: (cursor = "") =>
    ["operations", "private", "events", "list", cursor] as const,
  eventDetail: (eventId: string) =>
    ["operations", "private", "events", "detail", eventId] as const,
  offeringLists: (eventId: string) =>
    ["operations", "private", "events", eventId, "offerings", "list"] as const,
  offeringList: (eventId: string, cursor = "") =>
    [
      "operations",
      "private",
      "events",
      eventId,
      "offerings",
      "list",
      cursor,
    ] as const,
  offeringDetail: (offeringId: string) =>
    ["operations", "private", "offerings", "detail", offeringId] as const,
};

export function useOperationsSession() {
  return useQuery({
    queryKey: operationsQueryKeys.session,
    queryFn: ({ signal }) => sessionApi.get({ signal }),
    retry: false,
    staleTime: Number.POSITIVE_INFINITY,
    refetchOnMount: false,
    refetchOnReconnect: false,
    refetchOnWindowFocus: false,
  });
}

export function useEvents(cursor: string, enabled: boolean) {
  return useQuery({
    queryKey: operationsQueryKeys.eventList(cursor),
    queryFn: ({ signal }) =>
      eventsApi.listEvents({ cursor: cursor || undefined, signal }),
    enabled,
  });
}

export function useEvent(eventId: string, enabled: boolean) {
  return useQuery({
    queryKey: operationsQueryKeys.eventDetail(eventId),
    queryFn: ({ signal }) => eventsApi.getEvent(eventId, { signal }),
    enabled: enabled && eventId !== "",
  });
}

export function useOfferings(
  eventId: string,
  cursor: string,
  enabled: boolean,
) {
  return useQuery({
    queryKey: operationsQueryKeys.offeringList(eventId, cursor),
    queryFn: ({ signal }) =>
      eventsApi.listOfferings(eventId, { cursor: cursor || undefined, signal }),
    enabled: enabled && eventId !== "",
  });
}

export function useOffering(offeringId: string, enabled: boolean) {
  return useQuery({
    queryKey: operationsQueryKeys.offeringDetail(offeringId),
    queryFn: ({ signal }) => eventsApi.getOffering(offeringId, { signal }),
    enabled: enabled && offeringId !== "",
  });
}
