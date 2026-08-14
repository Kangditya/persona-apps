import { ApiError } from "@persona-apps/api-client";
import { describe, expect, it } from "vitest";

import {
  catalogueQueryKeys,
  retryPublicRead,
} from "../features/catalogue/queries";
import { createCatalogueApi, isRFC3339 } from "./catalogue";
import { createStorefrontApi } from "./client";

const eventID = "11111111-1111-1111-1111-111111111111";
const offeringID = "22222222-2222-2222-2222-222222222222";
const publicOffering = {
  id: offeringID,
  event_id: eventID,
  code: "COW-7",
  name: "Seven-person cow",
  offering_kind: "CATTLE_SHARE",
  description: "A public description",
  price_minor: 2_600_000,
  currency_code: "IDR",
  participant_capacity: 7,
  available_participant_units: 14,
  status: "PUBLISHED",
};

describe("public catalogue API", () => {
  it("loads the active Event without credentials or private headers", async () => {
    let url = "";
    let request: RequestInit | undefined;
    const api = testApi(async (input, init) => {
      url = String(input);
      request = init;
      return Response.json({ data: null });
    });

    await expect(api.getActiveEvent()).resolves.toBeNull();
    expect(url).toBe("/api/public/v1/events/active");
    expect(request?.credentials).toBe("omit");
    const headers = new Headers(request?.headers);
    expect(headers.get("Authorization")).toBeNull();
    expect(headers.get("X-CSRF-Token")).toBeNull();
    expect(headers.get("Idempotency-Key")).toBeNull();
  });

  it("parses the bounded published Offering page and exact fields", async () => {
    let url = "";
    const api = testApi(async (input) => {
      url = String(input);
      return Response.json({
        data: [publicOffering],
        page: { limit: 100, next_cursor: "next" },
      });
    });

    await expect(api.listOfferings(eventID)).resolves.toEqual({
      data: [
        {
          id: offeringID,
          eventId: eventID,
          code: "COW-7",
          name: "Seven-person cow",
          offeringKind: "CATTLE_SHARE",
          description: "A public description",
          priceMinor: 2_600_000,
          currencyCode: "IDR",
          participantCapacity: 7,
          availableParticipantUnits: 14,
          status: "PUBLISHED",
        },
      ],
      limit: 100,
      nextCursor: "next",
    });
    expect(url).toBe(`/api/public/v1/events/${eventID}/offerings?limit=100`);
  });

  it("rejects malformed, private, and invalid-identifier responses safely", async () => {
    let calls = 0;
    const api = testApi(async () => {
      calls += 1;
      return Response.json({ data: { ...publicOffering, version: 9 } });
    });

    await expect(api.getOffering(offeringID)).rejects.toMatchObject({
      kind: "unknown",
    } satisfies Partial<ApiError>);
    await expect(api.getOffering("not-a-uuid")).rejects.toMatchObject({
      kind: "not_found",
      code: "not_found",
    } satisfies Partial<ApiError>);
    expect(calls).toBe(1);
  });

  it("forwards cancellation and keeps retries bounded to transient errors", async () => {
    const controller = new AbortController();
    const api = testApi(
      (_input, init) =>
        new Promise((_resolve, reject) => {
          init?.signal?.addEventListener(
            "abort",
            () => reject(new DOMException("Aborted", "AbortError")),
            { once: true },
          );
        }),
    );
    const pending = api.getActiveEvent({ signal: controller.signal });
    controller.abort();
    await expect(pending).rejects.toMatchObject({
      kind: "cancelled",
    } satisfies Partial<ApiError>);

    expect(
      retryPublicRead(0, new ApiError("service_unavailable", "temporary", 503)),
    ).toBe(true);
    expect(retryPublicRead(2, new ApiError("unavailable", "temporary"))).toBe(
      false,
    );
    expect(
      retryPublicRead(0, new ApiError("rate_limited", "slow down", 429)),
    ).toBe(false);
  });

  it("keeps query keys deterministic and validates RFC 3339 instants", () => {
    expect(catalogueQueryKeys.offerings(eventID)).toEqual([
      "storefront",
      "catalogue",
      "events",
      eventID,
      "offerings",
    ]);
    expect(catalogueQueryKeys.offering(offeringID)).toEqual([
      "storefront",
      "catalogue",
      "offerings",
      offeringID,
    ]);
    expect(isRFC3339("2027-05-01T00:00:00Z")).toBe(true);
    expect(isRFC3339("2027-02-30T00:00:00Z")).toBe(false);
    expect(isRFC3339("May 1, 2027")).toBe(false);
  });
});

function testApi(fetch: typeof globalThis.fetch) {
  return createCatalogueApi(createStorefrontApi({ baseUrl: "", fetch }));
}
