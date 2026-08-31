import { ApiError } from "@persona-apps/api-client";

import { operationsApi, type OperationsApi } from "./client";

const apiRoot = "/api/operations/v1";
const channels = ["COMMON", "SAVING", "GIVEAWAY"] as const;
const statuses = [
  "DRAFT",
  "PENDING_PAYMENT",
  "PAID",
  "ELIGIBLE",
  "ALLOCATED",
  "COMPLETED",
  "CANCELLED",
] as const;

export const purchaseFilterStatuses = [
  "PENDING_PAYMENT",
  "PAID",
  "ELIGIBLE",
  "CANCELLED",
] as const;

export type PurchaseChannel = (typeof channels)[number];
export type PurchaseStatus = (typeof statuses)[number];
export type PurchaseFilterStatus = (typeof purchaseFilterStatuses)[number];

export type OperationsPartySummary = {
  id: string;
  displayName: string;
  email?: string;
  phone?: string;
};

export type OperationsPurchaseSummary = {
  id: string;
  eventId: string;
  purchaseRef: string;
  channel: PurchaseChannel;
  purchaser: OperationsPartySummary;
  payer: OperationsPartySummary | null;
  offeringId: string;
  offeringNameSnapshot: string;
  participantCount: number;
  totalAmountMinor: number;
  currencyCode: string;
  status: PurchaseStatus;
  eligibleAt: string | null;
  cancelledAt: string | null;
  cancellationReason: string | null;
  createdAt: string;
  updatedAt: string;
};

export type OperationsPurchaseParticipant = {
  sequenceNo: number;
  displayName: string;
};

export type OperationsPurchaseDetail = OperationsPurchaseSummary & {
  participants: readonly OperationsPurchaseParticipant[];
};

export type PurchasePage = {
  data: readonly OperationsPurchaseSummary[];
  limit: number;
  nextCursor: string | null;
};

export type PurchaseFilters = {
  eventId: string;
  status: PurchaseFilterStatus | "";
};

type ListPurchasesInput = PurchaseFilters & {
  cursor?: string;
  limit?: number;
  signal?: AbortSignal;
};

export function createPurchasesApi(client: OperationsApi = operationsApi) {
  return {
    async listPurchases(input: ListPurchasesInput): Promise<PurchasePage> {
      const query = new URLSearchParams();
      if (input.cursor) query.set("cursor", input.cursor);
      if (input.limit !== undefined) query.set("limit", String(input.limit));
      if (input.eventId) query.set("event_id", input.eventId);
      if (input.status) query.set("status", input.status);
      const suffix = query.size > 0 ? `?${query.toString()}` : "";
      return parsePage(
        await client.request<unknown>(`${apiRoot}/purchases${suffix}`, {
          signal: input.signal,
        }),
      );
    },
    async getPurchase(
      purchaseId: string,
      input: { signal?: AbortSignal } = {},
    ): Promise<OperationsPurchaseDetail> {
      return parseResource(
        await client.request<unknown>(
          `${apiRoot}/purchases/${encodeURIComponent(purchaseId)}`,
          { signal: input.signal },
        ),
      );
    },
  };
}

export const purchasesApi = createPurchasesApi();

export function isCanonicalUUID(value: string): boolean {
  return (
    value !== "00000000-0000-0000-0000-000000000000" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(value)
  );
}

function parsePage(payload: unknown): PurchasePage {
  if (
    !isRecord(payload) ||
    !onlyKeys(payload, ["data", "page"]) ||
    !Array.isArray(payload.data) ||
    !isRecord(payload.page) ||
    !onlyKeys(payload.page, ["limit", "next_cursor"])
  ) {
    throw invalidResponse();
  }
  const limit = safeInteger(payload.page.limit, 1, 100);
  const nextCursor = optionalBoundedString(payload.page.next_cursor, 512);
  if (limit === undefined || nextCursor === undefined) throw invalidResponse();
  return {
    data: payload.data.map((value) => parsePurchase(value, false)),
    limit,
    nextCursor,
  };
}

function parseResource(payload: unknown): OperationsPurchaseDetail {
  if (
    !isRecord(payload) ||
    !onlyKeys(payload, ["data"]) ||
    !("data" in payload)
  ) {
    throw invalidResponse();
  }
  return parsePurchase(payload.data, true);
}

function parsePurchase(
  value: unknown,
  includeParticipants: false,
): OperationsPurchaseSummary;
function parsePurchase(
  value: unknown,
  includeParticipants: true,
): OperationsPurchaseDetail;
function parsePurchase(
  value: unknown,
  includeParticipants: boolean,
): OperationsPurchaseSummary | OperationsPurchaseDetail {
  const allowedKeys = [
    "id",
    "event_id",
    "purchase_ref",
    "channel",
    "purchaser",
    "payer",
    "offering_id",
    "offering_name_snapshot",
    "participant_count",
    "total_amount_minor",
    "currency_code",
    "status",
    "eligible_at",
    "cancelled_at",
    "cancellation_reason",
    "created_at",
    "updated_at",
    ...(includeParticipants ? ["participants"] : []),
  ];
  if (!isRecord(value) || !onlyKeys(value, allowedKeys)) {
    throw invalidResponse();
  }

  const channel = enumValue(value.channel, channels);
  const status = enumValue(value.status, statuses);
  const participantCount = safeInteger(
    value.participant_count,
    1,
    2_147_483_647,
  );
  const totalAmountMinor = safeInteger(
    value.total_amount_minor,
    0,
    Number.MAX_SAFE_INTEGER,
  );
  const payer = optionalParty(value.payer);
  const eligibleAt = optionalDate(value.eligible_at);
  const cancelledAt = optionalDate(value.cancelled_at);
  const cancellationReason = optionalBoundedString(
    value.cancellation_reason,
    1_000,
  );

  if (
    !isCanonicalUUIDValue(value.id) ||
    !isCanonicalUUIDValue(value.event_id) ||
    !boundedString(value.purchase_ref, 255) ||
    channel === undefined ||
    !isRecord(value.purchaser) ||
    payer === undefined ||
    !isCanonicalUUIDValue(value.offering_id) ||
    !boundedString(value.offering_name_snapshot, 255) ||
    participantCount === undefined ||
    totalAmountMinor === undefined ||
    !currency(value.currency_code) ||
    status === undefined ||
    eligibleAt === undefined ||
    cancelledAt === undefined ||
    cancellationReason === undefined ||
    !validDate(value.created_at) ||
    !validDate(value.updated_at)
  ) {
    throw invalidResponse();
  }

  const summary: OperationsPurchaseSummary = {
    id: value.id,
    eventId: value.event_id,
    purchaseRef: value.purchase_ref,
    channel,
    purchaser: parseParty(value.purchaser),
    payer,
    offeringId: value.offering_id,
    offeringNameSnapshot: value.offering_name_snapshot,
    participantCount,
    totalAmountMinor,
    currencyCode: value.currency_code,
    status,
    eligibleAt,
    cancelledAt,
    cancellationReason,
    createdAt: value.created_at,
    updatedAt: value.updated_at,
  };

  if (!includeParticipants) return summary;
  if (!Array.isArray(value.participants)) throw invalidResponse();
  const participants = value.participants.map(parseParticipant);
  if (
    participants.length !== participantCount ||
    participants.some(
      (participant, index) => participant.sequenceNo !== index + 1,
    )
  ) {
    throw invalidResponse();
  }
  return { ...summary, participants };
}

function parseParty(value: Record<string, unknown>): OperationsPartySummary {
  if (!onlyKeys(value, ["id", "display_name", "email", "phone"])) {
    throw invalidResponse();
  }
  const email = optionalBoundedString(value.email, 320);
  const phone = optionalBoundedString(value.phone, 64);
  if (
    !isCanonicalUUIDValue(value.id) ||
    !boundedString(value.display_name, 255) ||
    email === undefined ||
    phone === undefined ||
    (email !== null && !/^[^\s@]+@[^\s@]+$/.test(email))
  ) {
    throw invalidResponse();
  }
  return {
    id: value.id,
    displayName: value.display_name,
    ...(email === null ? {} : { email }),
    ...(phone === null ? {} : { phone }),
  };
}

function optionalParty(
  value: unknown,
): OperationsPartySummary | null | undefined {
  if (value === undefined || value === null) return null;
  return isRecord(value) ? parseParty(value) : undefined;
}

function parseParticipant(value: unknown): OperationsPurchaseParticipant {
  if (!isRecord(value) || !onlyKeys(value, ["sequence_no", "display_name"])) {
    throw invalidResponse();
  }
  const sequenceNo = safeInteger(value.sequence_no, 1, 2_147_483_647);
  if (sequenceNo === undefined || !boundedString(value.display_name, 255)) {
    throw invalidResponse();
  }
  return { sequenceNo, displayName: value.display_name };
}

function optionalBoundedString(
  value: unknown,
  maximum: number,
): string | null | undefined {
  if (value === undefined || value === null) return null;
  return boundedString(value, maximum) ? value : undefined;
}

function boundedString(value: unknown, maximum: number): value is string {
  return (
    typeof value === "string" && value.length > 0 && value.length <= maximum
  );
}

function optionalDate(value: unknown): string | null | undefined {
  if (value === undefined || value === null) return null;
  return validDate(value) ? value : undefined;
}

function validDate(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
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

function enumValue<const T extends readonly string[]>(
  value: unknown,
  values: T,
): T[number] | undefined {
  return typeof value === "string" && values.includes(value)
    ? value
    : undefined;
}

function currency(value: unknown): value is string {
  return typeof value === "string" && /^[A-Z]{3}$/.test(value);
}

function isCanonicalUUIDValue(value: unknown): value is string {
  return typeof value === "string" && isCanonicalUUID(value);
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
    "The API returned an invalid Operations Purchase response",
  );
}
