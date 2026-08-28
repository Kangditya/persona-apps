import { ApiError, type ApiRequest } from "@persona-apps/api-client";

import { isCatalogueID, isRFC3339 } from "./catalogue";
import { storefrontApi, type StorefrontApi } from "./client";

const purchasePath = "/api/public/v1/purchases";
const idempotencyKeyPattern = /^[A-Za-z0-9][A-Za-z0-9._:-]{7,254}$/;
const purchaseReferencePattern = /^QRB-[0-9]{4}-[A-Z2-7]{16}$/;
const accessTokenPattern = /^[A-Za-z0-9_-]{43}$/;

export type PartyDeclaration = {
  party_ref: string;
  display_name: string;
  email?: string;
  phone?: string;
};

export type PartyReference = {
  party_ref: string;
};

export type ParticipantNameInput = {
  display_name: string;
};

export type CreatePurchaseRequest = {
  offering_id: string;
  purchaser: PartyDeclaration;
  payer: PartyDeclaration | PartyReference;
  participants: readonly (PartyReference | ParticipantNameInput)[];
};

export type CreatedPurchase = {
  id: string;
  purchaseRef: string;
  channel: "COMMON";
  offering: {
    id: string;
    name: string;
    kind: string;
    unitPriceMinor: number;
    participantCapacity: number;
  };
  participantCount: number;
  participants: readonly {
    sequenceNo: number;
    displayName: string;
  }[];
  totalAmountMinor: number;
  currencyCode: string;
  status: "PENDING_PAYMENT";
  reservationExpiresAt: string;
  createdAt: string;
};

export type CreatePurchaseResult = {
  purchase: CreatedPurchase;
  accessToken: string;
};

type PurchasesClient = Pick<StorefrontApi, "request">;

export function createPurchasesApi(client: PurchasesClient = storefrontApi) {
  return {
    async create(
      input: CreatePurchaseRequest,
      idempotencyKey: string,
      request?: ApiRequest,
    ): Promise<CreatePurchaseResult> {
      if (!idempotencyKeyPattern.test(idempotencyKey)) {
        throw new ApiError("validation", "The checkout retry key is invalid");
      }
      const headers = new Headers(request?.headers);
      headers.set("Idempotency-Key", idempotencyKey);
      return parseCreatePurchaseResponse(
        await client.request<unknown>(purchasePath, {
          ...request,
          method: "POST",
          headers,
          body: input,
        }),
      );
    },
  };
}

export const purchasesApi = createPurchasesApi();

function parseCreatePurchaseResponse(payload: unknown): CreatePurchaseResult {
  if (!isRecord(payload) || !onlyKeys(payload, ["data"])) {
    throw invalidResponse();
  }
  const data = payload.data;
  if (
    !isRecord(data) ||
    !onlyKeys(data, ["purchase", "access_token"]) ||
    typeof data.access_token !== "string" ||
    !accessTokenPattern.test(data.access_token)
  ) {
    throw invalidResponse();
  }
  return {
    purchase: parsePurchase(data.purchase),
    accessToken: data.access_token,
  };
}

function parsePurchase(value: unknown): CreatedPurchase {
  if (
    !isRecord(value) ||
    !onlyKeys(value, [
      "id",
      "purchase_ref",
      "channel",
      "offering",
      "participant_count",
      "participants",
      "total_amount_minor",
      "currency_code",
      "status",
      "reservation_expires_at",
      "created_at",
    ])
  ) {
    throw invalidResponse();
  }
  const offering = parseOffering(value.offering);
  const participantCount = safeInteger(
    value.participant_count,
    1,
    offering.participantCapacity,
  );
  const totalAmountMinor = safeInteger(
    value.total_amount_minor,
    0,
    Number.MAX_SAFE_INTEGER,
  );
  if (
    !isCatalogueIDValue(value.id) ||
    typeof value.purchase_ref !== "string" ||
    !purchaseReferencePattern.test(value.purchase_ref) ||
    value.channel !== "COMMON" ||
    participantCount === undefined ||
    !Array.isArray(value.participants) ||
    totalAmountMinor === undefined ||
    !currency(value.currency_code) ||
    value.status !== "PENDING_PAYMENT" ||
    !isRFC3339(value.reservation_expires_at) ||
    !isRFC3339(value.created_at)
  ) {
    throw invalidResponse();
  }
  const participants = value.participants.map(parseParticipant);
  if (
    participants.length !== participantCount ||
    participants.some(
      (participant, index) => participant.sequenceNo !== index + 1,
    ) ||
    !validTotal(offering.unitPriceMinor, participantCount, totalAmountMinor)
  ) {
    throw invalidResponse();
  }
  return {
    id: value.id,
    purchaseRef: value.purchase_ref,
    channel: "COMMON",
    offering,
    participantCount,
    participants,
    totalAmountMinor,
    currencyCode: value.currency_code,
    status: "PENDING_PAYMENT",
    reservationExpiresAt: value.reservation_expires_at,
    createdAt: value.created_at,
  };
}

function parseOffering(value: unknown): CreatedPurchase["offering"] {
  if (
    !isRecord(value) ||
    !onlyKeys(value, [
      "id",
      "name",
      "kind",
      "unit_price_minor",
      "participant_capacity",
    ])
  ) {
    throw invalidResponse();
  }
  const unitPriceMinor = safeInteger(
    value.unit_price_minor,
    0,
    Number.MAX_SAFE_INTEGER,
  );
  const participantCapacity = safeInteger(
    value.participant_capacity,
    1,
    2_147_483_647,
  );
  if (
    !isCatalogueIDValue(value.id) ||
    !boundedString(value.name, 1, 255) ||
    !boundedString(value.kind, 1, 64) ||
    unitPriceMinor === undefined ||
    participantCapacity === undefined
  ) {
    throw invalidResponse();
  }
  return {
    id: value.id,
    name: value.name,
    kind: value.kind,
    unitPriceMinor,
    participantCapacity,
  };
}

function parseParticipant(
  value: unknown,
): CreatedPurchase["participants"][number] {
  if (!isRecord(value) || !onlyKeys(value, ["sequence_no", "display_name"])) {
    throw invalidResponse();
  }
  const sequenceNo = safeInteger(value.sequence_no, 1, Number.MAX_SAFE_INTEGER);
  if (sequenceNo === undefined || !boundedString(value.display_name, 1, 255)) {
    throw invalidResponse();
  }
  return { sequenceNo, displayName: value.display_name };
}

function validTotal(
  unitPriceMinor: number,
  participantCount: number,
  totalAmountMinor: number,
): boolean {
  return (
    (unitPriceMinor === 0 ||
      participantCount <= Number.MAX_SAFE_INTEGER / unitPriceMinor) &&
    unitPriceMinor * participantCount === totalAmountMinor
  );
}

function safeInteger(
  value: unknown,
  minimum: number,
  maximum: number,
): number | undefined {
  return typeof value === "number" &&
    Number.isSafeInteger(value) &&
    value >= minimum &&
    value <= maximum
    ? value
    : undefined;
}

function boundedString(
  value: unknown,
  minimum: number,
  maximum: number,
): value is string {
  return (
    typeof value === "string" &&
    [...value].length >= minimum &&
    [...value].length <= maximum
  );
}

function currency(value: unknown): value is string {
  return typeof value === "string" && /^[A-Z]{3}$/.test(value);
}

function isCatalogueIDValue(value: unknown): value is string {
  return typeof value === "string" && isCatalogueID(value);
}

function onlyKeys(
  value: Record<string, unknown>,
  allowed: readonly string[],
): boolean {
  return Object.keys(value).every((key) => allowed.includes(key));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function invalidResponse(): ApiError {
  return new ApiError(
    "unknown",
    "The API returned an invalid Purchase response",
  );
}
