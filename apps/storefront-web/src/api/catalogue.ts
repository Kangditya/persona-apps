import { ApiError, type ApiRequest } from "@persona-apps/api-client";

import { storefrontApi, type StorefrontApi } from "./client";

const publicRoot = "/api/public/v1";
const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export type PublicEvent = {
  id: string;
  eventYear: number;
  name: string;
  status: "ACTIVE";
  registrationOpensAt: string | null;
  registrationClosesAt: string | null;
};

export type PublicOffering = {
  id: string;
  eventId: string;
  code: string;
  name: string;
  offeringKind: string;
  description: string | null;
  priceMinor: number;
  currencyCode: string;
  participantCapacity: number;
  availableParticipantUnits: number | null;
  status: "PUBLISHED";
};

export type PublicOfferingPage = {
  data: readonly PublicOffering[];
  limit: number;
  nextCursor: string | null;
};

type CatalogueClient = Pick<StorefrontApi, "request">;

export function createCatalogueApi(client: CatalogueClient = storefrontApi) {
  return {
    async getActiveEvent(request?: ApiRequest): Promise<PublicEvent | null> {
      const payload = await client.request<unknown>(
        `${publicRoot}/events/active`,
        request,
      );
      if (!isRecord(payload) || !onlyKeys(payload, ["data"])) {
        throw invalidResponse();
      }
      return payload.data === null ? null : parseEvent(payload.data);
    },

    async listOfferings(
      eventId: string,
      request?: ApiRequest,
    ): Promise<PublicOfferingPage> {
      if (!isCatalogueID(eventId)) throw unavailableResource();
      const payload = await client.request<unknown>(
        `${publicRoot}/events/${encodeURIComponent(eventId)}/offerings?limit=100`,
        request,
      );
      return parseOfferingPage(payload);
    },

    async getOffering(
      offeringId: string,
      request?: ApiRequest,
    ): Promise<PublicOffering> {
      if (!isCatalogueID(offeringId)) throw unavailableResource();
      const payload = await client.request<unknown>(
        `${publicRoot}/offerings/${encodeURIComponent(offeringId)}`,
        request,
      );
      if (!isRecord(payload) || !onlyKeys(payload, ["data"])) {
        throw invalidResponse();
      }
      return parseOffering(payload.data);
    },
  };
}

export const catalogueApi = createCatalogueApi();

export function isCatalogueID(value: string): boolean {
  return uuidPattern.test(value);
}

export function isRFC3339(value: unknown): value is string {
  if (typeof value !== "string") return false;
  const match =
    /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):([0-5]\d):([0-5]\d)(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/.exec(
      value,
    );
  if (!match) return false;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  return (
    day <= new Date(Date.UTC(year, month, 0)).getUTCDate() &&
    Number.isFinite(Date.parse(value))
  );
}

function parseOfferingPage(payload: unknown): PublicOfferingPage {
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
  const nextCursor = optionalString(payload.page, "next_cursor", 1, 512);
  if (limit === undefined || nextCursor === undefined) throw invalidResponse();
  return {
    data: payload.data.map(parseOffering),
    limit,
    nextCursor,
  };
}

function parseEvent(value: unknown): PublicEvent {
  if (
    !isRecord(value) ||
    !onlyKeys(value, [
      "id",
      "event_year",
      "name",
      "status",
      "registration_opens_at",
      "registration_closes_at",
    ])
  ) {
    throw invalidResponse();
  }
  const eventYear = safeInteger(value.event_year, 1900, 9999);
  const opensAt = optionalInstant(value, "registration_opens_at");
  const closesAt = optionalInstant(value, "registration_closes_at");
  if (
    !isCatalogueIDValue(value.id) ||
    eventYear === undefined ||
    !boundedString(value.name, 1, 255) ||
    value.status !== "ACTIVE" ||
    opensAt === undefined ||
    closesAt === undefined ||
    (opensAt !== null &&
      closesAt !== null &&
      Date.parse(opensAt) >= Date.parse(closesAt))
  ) {
    throw invalidResponse();
  }
  return {
    id: value.id,
    eventYear,
    name: value.name,
    status: "ACTIVE",
    registrationOpensAt: opensAt,
    registrationClosesAt: closesAt,
  };
}

function parseOffering(value: unknown): PublicOffering {
  if (
    !isRecord(value) ||
    !onlyKeys(value, [
      "id",
      "event_id",
      "code",
      "name",
      "offering_kind",
      "description",
      "price_minor",
      "currency_code",
      "participant_capacity",
      "available_participant_units",
      "status",
    ])
  ) {
    throw invalidResponse();
  }
  const description = optionalString(value, "description", 0, 10_000);
  const priceMinor = safeInteger(value.price_minor, 0, Number.MAX_SAFE_INTEGER);
  const participantCapacity = safeInteger(
    value.participant_capacity,
    1,
    2_147_483_647,
  );
  const availability = nullableSafeInteger(
    value,
    "available_participant_units",
  );
  if (
    !isCatalogueIDValue(value.id) ||
    !isCatalogueIDValue(value.event_id) ||
    !boundedString(value.code, 1, 64) ||
    !boundedString(value.name, 1, 255) ||
    !boundedString(value.offering_kind, 1, 64) ||
    description === undefined ||
    priceMinor === undefined ||
    !currency(value.currency_code) ||
    participantCapacity === undefined ||
    availability === undefined ||
    value.status !== "PUBLISHED"
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
    priceMinor,
    currencyCode: value.currency_code,
    participantCapacity,
    availableParticipantUnits: availability,
    status: "PUBLISHED",
  };
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

function optionalString(
  value: Record<string, unknown>,
  key: string,
  minimum: number,
  maximum: number,
): string | null | undefined {
  if (!(key in value)) return null;
  return boundedString(value[key], minimum, maximum) ? value[key] : undefined;
}

function optionalInstant(
  value: Record<string, unknown>,
  key: string,
): string | null | undefined {
  if (!(key in value)) return null;
  return isRFC3339(value[key]) ? value[key] : undefined;
}

function nullableSafeInteger(
  value: Record<string, unknown>,
  key: string,
): number | null | undefined {
  if (!(key in value)) return undefined;
  return value[key] === null
    ? null
    : safeInteger(value[key], 0, Number.MAX_SAFE_INTEGER);
}

function boundedString(
  value: unknown,
  minimum: number,
  maximum: number,
): value is string {
  return (
    typeof value === "string" &&
    value.length >= minimum &&
    value.length <= maximum
  );
}

function currency(value: unknown): value is string {
  return typeof value === "string" && /^[A-Z]{3}$/.test(value);
}

function isCatalogueIDValue(value: unknown): value is string {
  return typeof value === "string" && isCatalogueID(value);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function onlyKeys(
  value: Record<string, unknown>,
  allowed: readonly string[],
): boolean {
  return Object.keys(value).every((key) => allowed.includes(key));
}

function invalidResponse(): ApiError {
  return new ApiError(
    "unknown",
    "The API returned an invalid public catalogue response",
  );
}

function unavailableResource(): ApiError {
  return new ApiError("not_found", "The public resource is unavailable", 404, {
    code: "not_found",
  });
}
