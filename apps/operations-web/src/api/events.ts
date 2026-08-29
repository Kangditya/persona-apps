import { ApiError, type ApiRequest } from "@persona-apps/api-client";

import { operationsApi, type OperationsApi } from "./client";

const apiRoot = "/api/operations/v1";
const eventStatuses = [
  "DRAFT",
  "PUBLISHED",
  "ACTIVE",
  "SUSPENDED",
  "CLOSED",
  "ARCHIVED",
] as const;
const offeringStatuses = [
  "DRAFT",
  "PUBLISHED",
  "UNAVAILABLE",
  "ARCHIVED",
] as const;

export type EventStatus = (typeof eventStatuses)[number];
export type OfferingStatus = (typeof offeringStatuses)[number];

export type OperationsEvent = {
  id: string;
  eventYear: number;
  name: string;
  status: EventStatus;
  registrationOpensAt: string | null;
  registrationClosesAt: string | null;
  participantQuota: number | null;
  version: number;
  createdAt: string;
  updatedAt: string;
};

export type OperationsOffering = {
  id: string;
  eventId: string;
  code: string;
  name: string;
  offeringKind: string;
  description: string | null;
  priceMinor: number;
  currencyCode: string;
  participantCapacity: number;
  participantQuota: number | null;
  availableParticipantUnits: number | null;
  status: OfferingStatus;
  publishedAt: string | null;
  version: number;
  createdAt: string;
  updatedAt: string;
};

export type Page<T> = {
  data: readonly T[];
  limit: number;
  nextCursor: string | null;
};

export type CreateEventInput = {
  event_year: number;
  name: string;
  registration_opens_at: string | null;
  registration_closes_at: string | null;
  participant_quota: number | null;
};

export type PatchEventInput = Partial<
  Pick<
    CreateEventInput,
    | "name"
    | "registration_opens_at"
    | "registration_closes_at"
    | "participant_quota"
  >
> & { expected_version: number };

export type CreateOfferingInput = {
  code: string;
  name: string;
  offering_kind: string;
  description: string | null;
  price_minor: number;
  currency_code: string;
  participant_capacity: number;
  participant_quota: number | null;
};

export type PatchOfferingInput = Partial<
  Pick<
    CreateOfferingInput,
    "name" | "description" | "price_minor" | "participant_quota"
  >
> & { expected_version: number };

export type EventTransition =
  "publish" | "activate" | "suspend" | "close" | "archive";
export type OfferingTransition = "publish" | "unavailable" | "archive";

type ListInput = { cursor?: string; limit?: number; signal?: AbortSignal };

export function createEventsApi(client: OperationsApi = operationsApi) {
  return {
    async listEvents(input: ListInput = {}): Promise<Page<OperationsEvent>> {
      const payload = await client.request<unknown>(
        `${apiRoot}/events${listQuery(input)}`,
        { signal: input.signal },
      );
      return parsePage(payload, parseEvent);
    },
    async getEvent(id: string, request?: ApiRequest): Promise<OperationsEvent> {
      return parseResource(
        await client.request<unknown>(
          `${apiRoot}/events/${encodeURIComponent(id)}`,
          request,
        ),
        parseEvent,
      );
    },
    async createEvent(
      input: CreateEventInput,
      csrfToken: string,
      idempotencyKey: string,
      request?: ApiRequest,
    ): Promise<OperationsEvent> {
      return parseResource(
        await command(
          client,
          `${apiRoot}/events`,
          input,
          csrfToken,
          idempotencyKey,
          request,
        ),
        parseEvent,
      );
    },
    async patchEvent(
      id: string,
      input: PatchEventInput,
      csrfToken: string,
      request?: ApiRequest,
    ): Promise<OperationsEvent> {
      return parseResource(
        await command(
          client,
          `${apiRoot}/events/${encodeURIComponent(id)}`,
          input,
          csrfToken,
          undefined,
          { ...request, method: "PATCH" },
        ),
        parseEvent,
      );
    },
    async transitionEvent(
      id: string,
      action: EventTransition,
      expectedVersion: number,
      csrfToken: string,
      idempotencyKey: string,
      request?: ApiRequest,
    ): Promise<OperationsEvent> {
      return parseResource(
        await command(
          client,
          `${apiRoot}/events/${encodeURIComponent(id)}/${action}`,
          { expected_version: expectedVersion },
          csrfToken,
          idempotencyKey,
          request,
        ),
        parseEvent,
      );
    },
    async listOfferings(
      eventId: string,
      input: ListInput = {},
    ): Promise<Page<OperationsOffering>> {
      const payload = await client.request<unknown>(
        `${apiRoot}/events/${encodeURIComponent(eventId)}/offerings${listQuery(input)}`,
        { signal: input.signal },
      );
      return parsePage(payload, parseOffering);
    },
    async getOffering(
      id: string,
      request?: ApiRequest,
    ): Promise<OperationsOffering> {
      return parseResource(
        await client.request<unknown>(
          `${apiRoot}/offerings/${encodeURIComponent(id)}`,
          request,
        ),
        parseOffering,
      );
    },
    async createOffering(
      eventId: string,
      input: CreateOfferingInput,
      csrfToken: string,
      idempotencyKey: string,
      request?: ApiRequest,
    ): Promise<OperationsOffering> {
      return parseResource(
        await command(
          client,
          `${apiRoot}/events/${encodeURIComponent(eventId)}/offerings`,
          input,
          csrfToken,
          idempotencyKey,
          request,
        ),
        parseOffering,
      );
    },
    async patchOffering(
      id: string,
      input: PatchOfferingInput,
      csrfToken: string,
      request?: ApiRequest,
    ): Promise<OperationsOffering> {
      return parseResource(
        await command(
          client,
          `${apiRoot}/offerings/${encodeURIComponent(id)}`,
          input,
          csrfToken,
          undefined,
          { ...request, method: "PATCH" },
        ),
        parseOffering,
      );
    },
    async transitionOffering(
      id: string,
      action: OfferingTransition,
      expectedVersion: number,
      csrfToken: string,
      idempotencyKey: string,
      request?: ApiRequest,
    ): Promise<OperationsOffering> {
      return parseResource(
        await command(
          client,
          `${apiRoot}/offerings/${encodeURIComponent(id)}/${action}`,
          { expected_version: expectedVersion },
          csrfToken,
          idempotencyKey,
          request,
        ),
        parseOffering,
      );
    },
  };
}

export const eventsApi = createEventsApi();

async function command(
  client: OperationsApi,
  path: string,
  body: unknown,
  csrfToken: string,
  idempotencyKey?: string,
  request: ApiRequest = {},
): Promise<unknown> {
  const headers = new Headers(request.headers);
  headers.set("X-CSRF-Token", csrfToken);
  if (idempotencyKey) headers.set("Idempotency-Key", idempotencyKey);
  return client.request<unknown>(path, {
    ...request,
    method: request.method ?? "POST",
    headers,
    body,
  });
}

function listQuery(input: ListInput): string {
  const query = new URLSearchParams();
  if (input.cursor) query.set("cursor", input.cursor);
  if (input.limit !== undefined) query.set("limit", String(input.limit));
  const value = query.toString();
  return value ? `?${value}` : "";
}

function parsePage<T>(payload: unknown, parse: (value: unknown) => T): Page<T> {
  if (
    !isRecord(payload) ||
    !Array.isArray(payload.data) ||
    !isRecord(payload.page)
  ) {
    throw invalidResponse();
  }
  const limit = safeInteger(payload.page.limit, 1, 100);
  const nextCursor = nullableString(payload.page.next_cursor);
  if (limit === undefined || nextCursor === undefined) throw invalidResponse();
  return { data: payload.data.map(parse), limit, nextCursor };
}

function parseResource<T>(payload: unknown, parse: (value: unknown) => T): T {
  if (!isRecord(payload) || !("data" in payload)) throw invalidResponse();
  return parse(payload.data);
}

function parseEvent(value: unknown): OperationsEvent {
  if (!isRecord(value)) throw invalidResponse();
  const eventYear = safeInteger(value.event_year, 1900, 9999);
  const quota = nullableSafeInteger(value.participant_quota);
  const version = safeInteger(value.version, 1, Number.MAX_SAFE_INTEGER);
  const status = enumValue(value.status, eventStatuses);
  const opens = nullableDate(value.registration_opens_at);
  const closes = nullableDate(value.registration_closes_at);
  if (
    !requiredString(value.id) ||
    eventYear === undefined ||
    !requiredString(value.name) ||
    status === undefined ||
    opens === undefined ||
    closes === undefined ||
    quota === undefined ||
    version === undefined ||
    !validDate(value.created_at) ||
    !validDate(value.updated_at)
  ) {
    throw invalidResponse();
  }
  return {
    id: value.id,
    eventYear,
    name: value.name,
    status,
    registrationOpensAt: opens,
    registrationClosesAt: closes,
    participantQuota: quota,
    version,
    createdAt: value.created_at,
    updatedAt: value.updated_at,
  };
}

function parseOffering(value: unknown): OperationsOffering {
  if (!isRecord(value)) throw invalidResponse();
  const price = safeInteger(value.price_minor, 0, Number.MAX_SAFE_INTEGER);
  const capacity = safeInteger(value.participant_capacity, 1, 2_147_483_647);
  const quota = nullableSafeInteger(value.participant_quota);
  const available = nullableSafeInteger(value.available_participant_units);
  const version = safeInteger(value.version, 1, Number.MAX_SAFE_INTEGER);
  const status = enumValue(value.status, offeringStatuses);
  const publishedAt = nullableDate(value.published_at);
  const description = nullableString(value.description);
  if (
    !requiredString(value.id) ||
    !requiredString(value.event_id) ||
    !requiredString(value.code) ||
    !requiredString(value.name) ||
    !requiredString(value.offering_kind) ||
    description === undefined ||
    price === undefined ||
    !currency(value.currency_code) ||
    capacity === undefined ||
    quota === undefined ||
    available === undefined ||
    status === undefined ||
    publishedAt === undefined ||
    version === undefined ||
    !validDate(value.created_at) ||
    !validDate(value.updated_at)
  ) {
    throw invalidResponse();
  }
  return {
    id: value.id,
    eventId: value.event_id,
    code: value.code,
    name: value.name,
    offeringKind: value.offering_kind,
    description,
    priceMinor: price,
    currencyCode: value.currency_code,
    participantCapacity: capacity,
    participantQuota: quota,
    availableParticipantUnits: available,
    status,
    publishedAt,
    version,
    createdAt: value.created_at,
    updatedAt: value.updated_at,
  };
}

function nullableString(value: unknown): string | null | undefined {
  return value === undefined || value === null
    ? null
    : typeof value === "string"
      ? value
      : undefined;
}

function nullableDate(value: unknown): string | null | undefined {
  const normalized = nullableString(value);
  return normalized === null ||
    (typeof normalized === "string" && validDate(normalized))
    ? normalized
    : undefined;
}

function nullableSafeInteger(value: unknown): number | null | undefined {
  if (value === undefined || value === null) return null;
  return safeInteger(value, 0, Number.MAX_SAFE_INTEGER);
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

function requiredString(value: unknown): value is string {
  return typeof value === "string" && value.length > 0;
}

function currency(value: unknown): value is string {
  return typeof value === "string" && /^[A-Z]{3}$/.test(value);
}

function validDate(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function invalidResponse(): ApiError {
  return new ApiError(
    "unknown",
    "The API returned an invalid Event or Offering response",
  );
}
