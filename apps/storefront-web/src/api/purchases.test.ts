import { describe, expect, it, vi } from "vitest";

import { createStorefrontApi } from "./client";
import { createPurchasesApi, type CreatePurchaseRequest } from "./purchases";

const input: CreatePurchaseRequest = {
  offering_id: "11111111-1111-1111-1111-111111111111",
  purchaser: {
    party_ref: "purchaser",
    display_name: "Siti Aminah",
    email: "siti@example.test",
  },
  payer: { party_ref: "purchaser" },
  participants: [{ party_ref: "purchaser" }, { display_name: "Ahmad" }],
};

function successPayload() {
  return {
    data: {
      purchase: {
        id: "22222222-2222-2222-2222-222222222222",
        purchase_ref: "QRB-2026-ABCDEFGHIJKLMNOP",
        channel: "COMMON",
        offering: {
          id: input.offering_id,
          name: "Cattle share",
          kind: "SHARE",
          unit_price_minor: 100,
          participant_capacity: 2,
        },
        participant_count: 2,
        participants: [
          { sequence_no: 1, display_name: "Siti Aminah" },
          { sequence_no: 2, display_name: "Ahmad" },
        ],
        total_amount_minor: 200,
        currency_code: "IDR",
        status: "PENDING_PAYMENT",
        reservation_expires_at: "2026-08-27T10:00:00Z",
        created_at: "2026-08-26T10:00:00Z",
      },
      access_token: "a".repeat(43),
    },
  };
}

function purchasesWithFetch(fetch: typeof globalThis.fetch) {
  return createPurchasesApi(
    createStorefrontApi({ baseUrl: "", fetch, provider: "api" }),
  );
}

describe("Storefront Purchase API", () => {
  it("sends the exact non-credentialed idempotent checkout request", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>().mockResolvedValue(
      new Response(JSON.stringify(successPayload()), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );
    const result = await purchasesWithFetch(fetch).create(
      input,
      "checkout-key-123",
    );

    expect(result.purchase).toMatchObject({
      purchaseRef: "QRB-2026-ABCDEFGHIJKLMNOP",
      participantCount: 2,
      totalAmountMinor: 200,
      status: "PENDING_PAYMENT",
    });
    expect(result.accessToken).toHaveLength(43);
    expect(fetch).toHaveBeenCalledOnce();
    const [url, request] = fetch.mock.calls[0] ?? [];
    expect(url).toBe("/api/public/v1/purchases");
    expect(request).toMatchObject({
      method: "POST",
      credentials: "omit",
    });
    const headers = new Headers(request?.headers);
    expect(headers.get("Idempotency-Key")).toBe("checkout-key-123");
    expect(headers.get("Content-Type")).toBe("application/json");
    expect(JSON.parse(String(request?.body))).toEqual(input);
  });

  it("rejects an invalid retry key before transport", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>();
    await expect(
      purchasesWithFetch(fetch).create(input, "short"),
    ).rejects.toMatchObject({ kind: "validation" });
    expect(fetch).not.toHaveBeenCalled();
  });

  it("rejects malformed or inconsistent successful responses", async () => {
    for (const payload of [
      {
        ...successPayload(),
        data: { ...successPayload().data, access_token: "not-a-token" },
      },
      {
        ...successPayload(),
        data: {
          ...successPayload().data,
          purchase: {
            ...successPayload().data.purchase,
            total_amount_minor: 199,
          },
        },
      },
      {
        ...successPayload(),
        data: {
          ...successPayload().data,
          purchase: {
            ...successPayload().data.purchase,
            internal_field: "must not pass",
          },
        },
      },
    ]) {
      const fetch = vi
        .fn<typeof globalThis.fetch>()
        .mockResolvedValue(
          new Response(JSON.stringify(payload), { status: 201 }),
        );
      await expect(
        purchasesWithFetch(fetch).create(input, "checkout-key-123"),
      ).rejects.toMatchObject({ kind: "unknown" });
    }
  });
});
