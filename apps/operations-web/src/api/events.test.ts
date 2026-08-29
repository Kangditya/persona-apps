import { ApiError } from "@persona-apps/api-client";
import { describe, expect, it } from "vitest";

import { createOperationsApi } from "./client";
import { createEventsApi } from "./events";
import { createSessionApi, operationsLoginHref } from "./session";

const eventPayload = {
  id: "11111111-1111-1111-1111-111111111111",
  event_year: 2026,
  name: "Qurban 2026",
  status: "DRAFT",
  registration_opens_at: null,
  registration_closes_at: null,
  participant_quota: null,
  version: 1,
  created_at: "2026-08-14T00:00:00Z",
  updated_at: "2026-08-14T00:00:00Z",
};
const offeringPayload = {
  id: "33333333-3333-3333-3333-333333333333",
  event_id: eventPayload.id,
  code: "COW-7",
  name: "Cow share",
  offering_kind: "SHARE",
  description: null,
  price_minor: 2_500_000,
  currency_code: "IDR",
  participant_capacity: 7,
  participant_quota: 7,
  available_participant_units: 7,
  status: "DRAFT",
  published_at: null,
  version: 1,
  created_at: "2026-08-14T00:00:00Z",
  updated_at: "2026-08-14T00:00:00Z",
};

describe("Operations Event and Offering API", () => {
  it("loads the cookie session and logs out with only the CSRF header", async () => {
    const calls: Array<{ url: string; init?: RequestInit }> = [];
    const client = createOperationsApi({
      baseUrl: "",
      fetch: async (input, init) => {
        calls.push({ url: String(input), init });
        if (String(input).endsWith("/logout"))
          return new Response(null, { status: 204 });
        return Response.json({
          data: {
            operator: {
              id: "22222222-2222-2222-2222-222222222222",
              display_name: "Operator",
            },
            permissions: ["event.read", "event.manage"],
            expires_at: "2026-08-15T00:00:00Z",
            csrf_token: "c".repeat(43),
          },
        });
      },
    });
    const api = createSessionApi(client);

    await expect(api.get()).resolves.toMatchObject({
      operator: { displayName: "Operator" },
      permissions: ["event.read", "event.manage"],
    });
    await api.logout("c".repeat(43));

    expect(calls.map((call) => call.url)).toEqual([
      "/api/operations/v1/auth/session",
      "/api/operations/v1/auth/logout",
    ]);
    expect(calls[0]?.init?.credentials).toBe("include");
    expect(calls[1]?.init?.method).toBe("POST");
    const headers = new Headers(calls[1]?.init?.headers);
    expect(headers.get("X-CSRF-Token")).toBe("c".repeat(43));
    expect(headers.get("Idempotency-Key")).toBeNull();
    expect(operationsLoginHref("//attacker.invalid")).toBe(
      "/api/operations/v1/auth/login?return_to=%2Fevents",
    );
  });

  it("sends exact create, patch, and lifecycle command contracts", async () => {
    const calls: Array<{ url: string; init?: RequestInit }> = [];
    const client = createOperationsApi({
      baseUrl: "https://api.example.test/",
      fetch: async (input, init) => {
        calls.push({ url: String(input), init });
        return Response.json({
          data: String(input).includes("offerings")
            ? offeringPayload
            : eventPayload,
        });
      },
    });
    const api = createEventsApi(client);
    const csrf = "c".repeat(43);

    await api.createEvent(
      {
        event_year: 2026,
        name: "Qurban 2026",
        registration_opens_at: null,
        registration_closes_at: null,
        participant_quota: null,
      },
      csrf,
      "event.intent.1",
    );
    await api.patchEvent(
      eventPayload.id,
      { expected_version: 1, participant_quota: null },
      csrf,
    );
    await api.transitionEvent(
      eventPayload.id,
      "publish",
      1,
      csrf,
      "event.intent.2",
    );
    await api.createOffering(
      eventPayload.id,
      {
        code: "COW-7",
        name: "Cow share",
        offering_kind: "SHARE",
        description: null,
        price_minor: 2_500_000,
        currency_code: "IDR",
        participant_capacity: 7,
        participant_quota: 7,
      },
      csrf,
      "offering.intent.1",
    );
    await api.patchOffering(
      offeringPayload.id,
      { expected_version: 1, description: null, participant_quota: null },
      csrf,
    );
    await api.transitionOffering(
      offeringPayload.id,
      "publish",
      1,
      csrf,
      "offering.intent.2",
    );

    expect(calls.map((call) => call.url)).toEqual([
      "https://api.example.test/api/operations/v1/events",
      `https://api.example.test/api/operations/v1/events/${eventPayload.id}`,
      `https://api.example.test/api/operations/v1/events/${eventPayload.id}/publish`,
      `https://api.example.test/api/operations/v1/events/${eventPayload.id}/offerings`,
      `https://api.example.test/api/operations/v1/offerings/${offeringPayload.id}`,
      `https://api.example.test/api/operations/v1/offerings/${offeringPayload.id}/publish`,
    ]);
    expect(calls.map((call) => call.init?.credentials)).toEqual([
      "include",
      "include",
      "include",
      "include",
      "include",
      "include",
    ]);
    expect(new Headers(calls[0]?.init?.headers).get("Idempotency-Key")).toBe(
      "event.intent.1",
    );
    expect(calls[1]?.init?.method).toBe("PATCH");
    expect(
      new Headers(calls[1]?.init?.headers).get("Idempotency-Key"),
    ).toBeNull();
    expect(calls[1]?.init?.body).toBe(
      '{"expected_version":1,"participant_quota":null}',
    );
    expect(calls[2]?.init?.body).toBe('{"expected_version":1}');
    expect(new Headers(calls[2]?.init?.headers).get("Idempotency-Key")).toBe(
      "event.intent.2",
    );
    expect(new Headers(calls[3]?.init?.headers).get("Idempotency-Key")).toBe(
      "offering.intent.1",
    );
    expect(calls[4]?.init?.method).toBe("PATCH");
    expect(calls[4]?.init?.body).toBe(
      '{"expected_version":1,"description":null,"participant_quota":null}',
    );
    expect(
      new Headers(calls[4]?.init?.headers).get("Idempotency-Key"),
    ).toBeNull();
    expect(calls[5]?.init?.body).toBe('{"expected_version":1}');
    expect(new Headers(calls[5]?.init?.headers).get("Idempotency-Key")).toBe(
      "offering.intent.2",
    );
  });

  it("encodes cursors, forwards cancellation, and rejects malformed success data", async () => {
    let capturedURL = "";
    let capturedSignal: AbortSignal | null | undefined;
    const controller = new AbortController();
    const client = createOperationsApi({
      baseUrl: "",
      fetch: async (input, init) => {
        capturedURL = String(input);
        capturedSignal = init?.signal;
        return Response.json({
          data: [eventPayload],
          page: { limit: 50, next_cursor: "next/cursor" },
        });
      },
    });

    await expect(
      createEventsApi(client).listEvents({
        cursor: "opaque/cursor",
        limit: 50,
        signal: controller.signal,
      }),
    ).resolves.toMatchObject({ limit: 50, nextCursor: "next/cursor" });
    expect(capturedURL).toBe(
      "/api/operations/v1/events?cursor=opaque%2Fcursor&limit=50",
    );
    expect(capturedSignal).toBeTruthy();

    const malformed = createEventsApi(
      createOperationsApi({
        baseUrl: "",
        fetch: async () => Response.json({ data: { id: "only-an-id" } }),
      }),
    );
    await expect(malformed.getEvent(eventPayload.id)).rejects.toBeInstanceOf(
      ApiError,
    );
  });
});
