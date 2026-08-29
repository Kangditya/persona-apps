import { ApiError, type ApiRequest } from "@persona-apps/api-client";

import { operationsApi, type OperationsApi } from "./client";

const sessionPath = "/api/operations/v1/auth/session";
const logoutPath = "/api/operations/v1/auth/logout";

export type OperationsSession = {
  operator: { id: string; displayName: string };
  permissions: readonly string[];
  expiresAt: string;
  csrfToken: string;
};

export function createSessionApi(client: OperationsApi = operationsApi) {
  return {
    async get(request?: ApiRequest): Promise<OperationsSession> {
      return parseSession(await client.request<unknown>(sessionPath, request));
    },
    logout(csrfToken: string, request?: ApiRequest): Promise<void> {
      return client.request<void>(logoutPath, {
        ...request,
        method: "POST",
        headers: mergeHeaders(request?.headers, { "X-CSRF-Token": csrfToken }),
      });
    },
  };
}

export const sessionApi = createSessionApi();

export function operationsLoginHref(returnTo: string): string {
  const safeReturnTo =
    returnTo.startsWith("/") && !returnTo.startsWith("//")
      ? returnTo
      : "/events";
  return `/api/operations/v1/auth/login?return_to=${encodeURIComponent(safeReturnTo)}`;
}

function parseSession(payload: unknown): OperationsSession {
  const data = envelopeData(payload);
  const operator = isRecord(data.operator) ? data.operator : undefined;
  const permissions = Array.isArray(data.permissions)
    ? data.permissions.filter(
        (value): value is string => typeof value === "string",
      )
    : undefined;
  if (
    !operator ||
    typeof operator.id !== "string" ||
    typeof operator.display_name !== "string" ||
    !permissions ||
    !Array.isArray(data.permissions) ||
    permissions.length !== data.permissions.length ||
    typeof data.expires_at !== "string" ||
    !validDate(data.expires_at) ||
    typeof data.csrf_token !== "string" ||
    data.csrf_token.length < 43
  ) {
    throw invalidResponse();
  }
  return {
    operator: { id: operator.id, displayName: operator.display_name },
    permissions,
    expiresAt: data.expires_at,
    csrfToken: data.csrf_token,
  };
}

function envelopeData(payload: unknown): Record<string, unknown> {
  if (!isRecord(payload) || !isRecord(payload.data)) throw invalidResponse();
  return payload.data;
}

function mergeHeaders(
  current: HeadersInit | undefined,
  values: Record<string, string>,
): Headers {
  const headers = new Headers(current);
  for (const [name, value] of Object.entries(values)) headers.set(name, value);
  return headers;
}

function validDate(value: string): boolean {
  return !Number.isNaN(Date.parse(value));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function invalidResponse(): ApiError {
  return new ApiError(
    "unknown",
    "The API returned an invalid session response",
  );
}
