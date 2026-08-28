import { ApiError } from "@persona-apps/api-client";
import { describe, expect, it } from "vitest";

import {
  buildCheckoutRequest,
  checkoutErrorState,
  CheckoutValidationError,
  createCheckoutIntent,
  type CheckoutDraft,
} from "./checkout";

const offeringId = "11111111-1111-1111-1111-111111111111";

function draft(overrides: Partial<CheckoutDraft> = {}): CheckoutDraft {
  return {
    purchaser: {
      displayName: " Siti Aminah ",
      email: " siti@example.test ",
      phone: "",
    },
    payerMode: "same",
    payer: { displayName: "", email: "", phone: "" },
    participants: [
      { rowId: "participant-1", kind: "purchaser", displayName: "" },
      { rowId: "participant-2", kind: "name", displayName: " Ahmad " },
    ],
    ...overrides,
  };
}

describe("checkout contract mapping", () => {
  it("maps a shared purchaser and payer with ordered participants", () => {
    expect(buildCheckoutRequest(offeringId, 2, draft())).toEqual({
      offering_id: offeringId,
      purchaser: {
        party_ref: "purchaser",
        display_name: "Siti Aminah",
        email: "siti@example.test",
      },
      payer: { party_ref: "purchaser" },
      participants: [{ party_ref: "purchaser" }, { display_name: "Ahmad" }],
    });
  });

  it("maps a distinct payer and explicit payer participant", () => {
    expect(
      buildCheckoutRequest(
        offeringId,
        2,
        draft({
          payerMode: "different",
          payer: {
            displayName: " Budi Santoso ",
            email: "",
            phone: " +628100000000 ",
          },
          participants: [
            { rowId: "participant-1", kind: "payer", displayName: "" },
          ],
        }),
      ),
    ).toEqual({
      offering_id: offeringId,
      purchaser: {
        party_ref: "purchaser",
        display_name: "Siti Aminah",
        email: "siti@example.test",
      },
      payer: {
        party_ref: "payer",
        display_name: "Budi Santoso",
        phone: "+628100000000",
      },
      participants: [{ party_ref: "payer" }],
    });
  });

  it("rejects ambiguous payer participants and capacity overflow", () => {
    expect(() =>
      buildCheckoutRequest(
        offeringId,
        2,
        draft({
          participants: [
            { rowId: "participant-1", kind: "payer", displayName: "" },
          ],
        }),
      ),
    ).toThrow(CheckoutValidationError);
    expect(() => buildCheckoutRequest(offeringId, 1, draft())).toThrow(
      /between 1 and 1/,
    );
  });

  it("trims values and rejects missing or oversized required names", () => {
    expect(() =>
      buildCheckoutRequest(
        offeringId,
        1,
        draft({
          purchaser: { displayName: " ", email: "", phone: "" },
          participants: [
            { rowId: "participant-1", kind: "purchaser", displayName: "" },
          ],
        }),
      ),
    ).toThrow(/Purchaser name/);
    expect(() =>
      buildCheckoutRequest(
        offeringId,
        1,
        draft({
          purchaser: {
            displayName: "Siti",
            email: "a".repeat(321),
            phone: "",
          },
          participants: [
            { rowId: "participant-1", kind: "purchaser", displayName: "" },
          ],
        }),
      ),
    ).toThrow(/at most 320/);
  });
});

describe("checkout retry and error presentation", () => {
  it("freezes one generated key with one request", () => {
    const input = buildCheckoutRequest(offeringId, 2, draft());
    expect(createCheckoutIntent(input, () => "fixed-key")).toEqual({
      key: "fixed-key",
      input,
    });
  });

  it("allows same-intent retry only for uncertain service failures", () => {
    expect(
      checkoutErrorState(
        new ApiError("unavailable", "network", 500, {
          requestId: "req-1",
        }),
      ),
    ).toEqual({
      message:
        "The Purchase service is temporarily unavailable. Your entries are preserved.",
      requestId: "req-1",
      retryOriginal: true,
    });
    expect(
      checkoutErrorState(
        new ApiError("conflict", "quota", 409, {
          code: "quota_unavailable",
        }),
      ),
    ).toMatchObject({
      message: "Participant quota is no longer available for this Offering.",
      retryOriginal: false,
    });
  });
});
