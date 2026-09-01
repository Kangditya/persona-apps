import { ApiError } from "@persona-apps/api-client";
import { describe, expect, it } from "vitest";

import { purchaseQueryKeys } from "../features/purchases/queries";
import { createOperationsApi } from "./client";
import {
  createPurchasesApi,
  isCanonicalUUID,
  type PurchaseFilters,
} from "./purchases";

const purchaseId = "abcdefab-cdef-4abc-8def-abcdefabcdef";
const eventId = "22222222-2222-4222-8222-222222222222";
const offeringId = "33333333-3333-4333-8333-333333333333";
const purchaserId = "44444444-4444-4444-8444-444444444444";
const payerId = "55555555-5555-4555-8555-555555555555";

const summaryPayload = {
  id: purchaseId,
  event_id: eventId,
  purchase_ref: "QRB-2026-ABCDEFGH23456789",
  channel: "COMMON",
  purchaser: { id: purchaserId, display_name: "Purchaser" },
  payer: {
    id: payerId,
    display_name: "Payer",
    email: "payer@example.test",
    phone: "+62000000000",
  },
  offering_id: offeringId,
  offering_name_snapshot: "Cow share",
  participant_count: 2,
  total_amount_minor: 5_000_000,
  currency_code: "IDR",
  status: "PENDING_PAYMENT",
  created_at: "2026-08-31T00:00:00Z",
  updated_at: "2026-08-31T00:00:00Z",
};

const detailPayload = {
  ...summaryPayload,
  participants: [
    { sequence_no: 1, display_name: "First participant" },
    { sequence_no: 2, display_name: "Second participant" },
  ],
};

describe("Operations Purchase API", () => {
  it("encodes only contracted filters and opaque cursor on a credentialed GET", async () => {
    let capturedURL = "";
    let capturedInit: RequestInit | undefined;
    const controller = new AbortController();
    const api = createPurchasesApi(
      createOperationsApi({
        baseUrl: "",
        fetch: async (input, init) => {
          capturedURL = String(input);
          capturedInit = init;
          return Response.json({
            data: [summaryPayload],
            page: { limit: 25, next_cursor: "opaque/next" },
          });
        },
      }),
    );

    await expect(
      api.listPurchases({
        cursor: "opaque/current",
        limit: 25,
        eventId,
        status: "PENDING_PAYMENT",
        signal: controller.signal,
      }),
    ).resolves.toMatchObject({ limit: 25, nextCursor: "opaque/next" });

    expect(capturedURL).toBe(
      `/api/operations/v1/purchases?cursor=opaque%2Fcurrent&limit=25&event_id=${eventId}&status=PENDING_PAYMENT`,
    );
    expect(capturedInit?.method).toBeUndefined();
    expect(capturedInit?.credentials).toBe("include");
    expect(capturedInit?.signal).toBeTruthy();
    const headers = new Headers(capturedInit?.headers);
    expect(headers.get("X-CSRF-Token")).toBeNull();
    expect(headers.get("Idempotency-Key")).toBeNull();
  });

  it("parses detail-only participant snapshots and optional Party fields", async () => {
    let capturedURL = "";
    let capturedInit: RequestInit | undefined;
    const api = createPurchasesApi(
      createOperationsApi({
        baseUrl: "https://api.example.test/",
        fetch: async (input, init) => {
          capturedURL = String(input);
          capturedInit = init;
          return Response.json({ data: detailPayload });
        },
      }),
    );

    await expect(api.getPurchase(purchaseId)).resolves.toMatchObject({
      id: purchaseId,
      payer: { displayName: "Payer", email: "payer@example.test" },
      participants: [
        { sequenceNo: 1, displayName: "First participant" },
        { sequenceNo: 2, displayName: "Second participant" },
      ],
    });
    expect(capturedURL).toBe(
      `https://api.example.test/api/operations/v1/purchases/${purchaseId}`,
    );
    expect(capturedInit?.method).toBeUndefined();
    expect(capturedInit?.credentials).toBe("include");
    const headers = new Headers(capturedInit?.headers);
    expect(headers.get("X-CSRF-Token")).toBeNull();
    expect(headers.get("Idempotency-Key")).toBeNull();
  });

  it("preserves an absent payer without inferring the purchaser", async () => {
    const withoutPayer = { ...detailPayload, payer: undefined };
    const api = createPurchasesApi(
      createOperationsApi({
        baseUrl: "",
        fetch: async () => Response.json({ data: withoutPayer }),
      }),
    );
    await expect(api.getPurchase(purchaseId)).resolves.toMatchObject({
      payer: null,
    });
  });

  it("rejects list/detail shape leakage and inconsistent participant data", async () => {
    const invalidPayloads = [
      { data: [{ ...summaryPayload, participants: [] }], page: { limit: 50 } },
      { data: summaryPayload },
      {
        data: {
          ...detailPayload,
          participants: [
            { sequence_no: 2, display_name: "Second participant" },
            { sequence_no: 1, display_name: "First participant" },
          ],
        },
      },
      {
        data: {
          ...detailPayload,
          total_amount_minor: Number.MAX_SAFE_INTEGER + 1,
        },
      },
      { data: { ...detailPayload, access_token: "must-not-be-exposed" } },
    ];

    for (const [index, payload] of invalidPayloads.entries()) {
      const api = createPurchasesApi(
        createOperationsApi({
          baseUrl: "",
          fetch: async () => Response.json(payload),
        }),
      );
      const request =
        index === 0
          ? api.listPurchases({ eventId: "", status: "" })
          : api.getPurchase(purchaseId);
      await expect(request).rejects.toBeInstanceOf(ApiError);
    }
  });

  it("uses canonical UUIDs and isolated private query keys", () => {
    expect(isCanonicalUUID(purchaseId)).toBe(true);
    expect(isCanonicalUUID(purchaseId.toUpperCase())).toBe(false);
    expect(isCanonicalUUID("00000000-0000-0000-0000-000000000000")).toBe(false);

    const filters: PurchaseFilters = { eventId, status: "PAID" };
    expect(purchaseQueryKeys.list(filters, "opaque/cursor")).toEqual([
      "operations",
      "private",
      "purchases",
      "list",
      eventId,
      "PAID",
      "opaque/cursor",
    ]);
    expect(purchaseQueryKeys.detail(purchaseId)).toEqual([
      "operations",
      "private",
      "purchases",
      "detail",
      purchaseId,
    ]);
  });
});
