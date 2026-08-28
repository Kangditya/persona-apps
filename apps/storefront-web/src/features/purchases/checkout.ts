import { ApiError } from "@persona-apps/api-client";

import { isCatalogueID } from "../../api/catalogue";
import type {
  CreatePurchaseRequest,
  PartyDeclaration,
} from "../../api/purchases";

export type CheckoutPartyDraft = {
  displayName: string;
  email: string;
  phone: string;
};

export type CheckoutParticipantDraft = {
  rowId: string;
  kind: "purchaser" | "payer" | "name";
  displayName: string;
};

export type CheckoutDraft = {
  purchaser: CheckoutPartyDraft;
  payerMode: "same" | "different";
  payer: CheckoutPartyDraft;
  participants: readonly CheckoutParticipantDraft[];
};

export type CheckoutIntent = {
  key: string;
  input: CreatePurchaseRequest;
};

export type CheckoutErrorState = {
  message: string;
  requestId: string | null;
  retryOriginal: boolean;
};

export class CheckoutValidationError extends Error {
  readonly field: string;

  constructor(field: string, message: string) {
    super(message);
    this.name = "CheckoutValidationError";
    this.field = field;
  }
}

export function buildCheckoutRequest(
  offeringId: string,
  participantCapacity: number,
  draft: CheckoutDraft,
): CreatePurchaseRequest {
  if (!isCatalogueID(offeringId)) {
    throw new CheckoutValidationError(
      "offering_id",
      "The selected Offering is unavailable.",
    );
  }
  if (!Number.isSafeInteger(participantCapacity) || participantCapacity < 1) {
    throw new CheckoutValidationError(
      "participants",
      "The Offering participant capacity is invalid.",
    );
  }
  const purchaser = partyDeclaration("purchaser", "purchaser", draft.purchaser);
  const payer =
    draft.payerMode === "same"
      ? { party_ref: "purchaser" }
      : partyDeclaration("payer", "payer", draft.payer);
  if (
    draft.participants.length < 1 ||
    draft.participants.length > participantCapacity
  ) {
    throw new CheckoutValidationError(
      "participants",
      `Add between 1 and ${String(participantCapacity)} intended participants.`,
    );
  }
  const participants = draft.participants.map((participant, index) => {
    if (participant.kind === "purchaser") {
      return { party_ref: "purchaser" };
    }
    if (participant.kind === "payer") {
      if (draft.payerMode !== "different") {
        throw new CheckoutValidationError(
          `participants.${String(index)}.kind`,
          "Choose purchaser or name-only while purchaser and payer are the same.",
        );
      }
      return { party_ref: "payer" };
    }
    return {
      display_name: requiredText(
        `participants.${String(index)}.display_name`,
        `Participant ${String(index + 1)} name`,
        participant.displayName,
        255,
      ),
    };
  });
  return {
    offering_id: offeringId,
    purchaser,
    payer,
    participants,
  };
}

export function createCheckoutIntent(
  input: CreatePurchaseRequest,
  createKey: () => string = () => crypto.randomUUID(),
): CheckoutIntent {
  return { key: createKey(), input };
}

export function checkoutErrorState(error: unknown): CheckoutErrorState {
  if (!(error instanceof ApiError)) {
    return {
      message: "The Purchase could not be created.",
      requestId: null,
      retryOriginal: false,
    };
  }
  const message = (() => {
    switch (error.code) {
      case "quota_unavailable":
        return "Participant quota is no longer available for this Offering.";
      case "state_conflict":
        return "The Event or Offering can no longer accept this Purchase.";
      case "idempotency_conflict":
        return "This checkout retry no longer matches the original request.";
      case "request_too_large":
        return "The checkout request is too large.";
      case "unsupported_media_type":
        return "The checkout request format was rejected.";
    }
    switch (error.kind) {
      case "validation":
        return "Review the checkout details and correct the highlighted values.";
      case "rate_limited":
        return error.retryAfterSeconds
          ? `Too many checkout requests. Try again in ${String(error.retryAfterSeconds)} ${error.retryAfterSeconds === 1 ? "second" : "seconds"}.`
          : "Too many checkout requests. Try again shortly.";
      case "service_unavailable":
      case "unavailable":
        return "The Purchase service is temporarily unavailable. Your entries are preserved.";
      case "cancelled":
        return "The checkout request was cancelled.";
      default:
        return "The Purchase could not be created.";
    }
  })();
  return {
    message,
    requestId: error.requestId ?? null,
    retryOriginal:
      error.kind === "service_unavailable" || error.kind === "unavailable",
  };
}

function partyDeclaration(
  reference: "purchaser" | "payer",
  field: "purchaser" | "payer",
  draft: CheckoutPartyDraft,
): PartyDeclaration {
  const declaration: PartyDeclaration = {
    party_ref: reference,
    display_name: requiredText(
      `${field}.display_name`,
      field === "purchaser" ? "Purchaser name" : "Payer name",
      draft.displayName,
      255,
    ),
  };
  const email = optionalText(`${field}.email`, draft.email, 320);
  const phone = optionalText(`${field}.phone`, draft.phone, 64);
  if (email) declaration.email = email;
  if (phone) declaration.phone = phone;
  return declaration;
}

function requiredText(
  field: string,
  label: string,
  value: string,
  maximum: number,
): string {
  const trimmed = value.trim();
  const length = [...trimmed].length;
  if (length < 1 || length > maximum) {
    throw new CheckoutValidationError(
      field,
      `${label} must contain 1 to ${String(maximum)} characters.`,
    );
  }
  return trimmed;
}

function optionalText(
  field: string,
  value: string,
  maximum: number,
): string | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  if ([...trimmed].length > maximum) {
    throw new CheckoutValidationError(
      field,
      `This value must contain at most ${String(maximum)} characters.`,
    );
  }
  return trimmed;
}
