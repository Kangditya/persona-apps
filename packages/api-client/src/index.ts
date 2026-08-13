export type ApiErrorKind =
  | "cancelled"
  | "unauthorized"
  | "forbidden"
  | "validation"
  | "conflict"
  | "not_found"
  | "method_not_allowed"
  | "rate_limited"
  | "service_unavailable"
  | "unavailable"
  | "unknown";

export type ApiErrorDetails = Readonly<Record<string, unknown>>;

type ApiErrorMetadata = {
  code?: string;
  requestId?: string;
  details?: ApiErrorDetails;
  retryAfterSeconds?: number;
};

export class ApiError extends Error {
  readonly kind: ApiErrorKind;
  readonly status?: number;
  readonly code?: string;
  readonly requestId?: string;
  /** Untrusted response data: validate a recognized shape before using it. */
  readonly details?: ApiErrorDetails;
  readonly retryAfterSeconds?: number;

  constructor(
    kind: ApiErrorKind,
    message: string,
    status?: number,
    metadata: ApiErrorMetadata = {},
  ) {
    super(message);
    this.name = "ApiError";
    this.kind = kind;
    this.status = status;
    this.code = metadata.code;
    this.requestId = metadata.requestId;
    this.details = metadata.details;
    this.retryAfterSeconds = metadata.retryAfterSeconds;
  }
}

export type ApiRequest = Omit<RequestInit, "body" | "headers" | "signal"> & {
  body?: unknown;
  headers?: HeadersInit;
  signal?: AbortSignal;
};

export type ApiClientOptions = {
  baseUrl: string;
  fetch?: typeof fetch;
  getHeaders?: () => HeadersInit | undefined;
  timeoutMs?: number;
};

const defaultTimeoutMs = 10_000;

export function createApiClient({
  baseUrl,
  fetch: fetchImplementation = fetch,
  getHeaders,
  timeoutMs = defaultTimeoutMs,
}: ApiClientOptions) {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, "");

  return {
    async request<T>(path: string, request: ApiRequest = {}): Promise<T> {
      const controller = new AbortController();
      let timedOut = false;
      const timeout = setTimeout(() => {
        timedOut = true;
        controller.abort();
      }, timeoutMs);
      const abort = () => controller.abort();
      request.signal?.addEventListener("abort", abort, { once: true });

      try {
        const headers = new Headers(getHeaders?.());
        new Headers(request.headers).forEach((value, key) =>
          headers.set(key, value),
        );
        const body =
          request.body === undefined ? undefined : JSON.stringify(request.body);
        if (body !== undefined && !headers.has("Content-Type")) {
          headers.set("Content-Type", "application/json");
        }

        const response = await fetchImplementation(
          `${normalizedBaseUrl}${path}`,
          {
            ...request,
            body,
            headers,
            signal: controller.signal,
          },
        );
        const payload = await parseBody(response);

        if (!response.ok) {
          const error = responseError(payload, response.statusText);
          throw new ApiError(
            errorKind(response.status),
            error.message,
            response.status,
            {
              code: error.code,
              requestId: responseRequestID(response, error.requestId),
              details: error.details,
              retryAfterSeconds:
                response.status === 429
                  ? retryAfterSeconds(response.headers.get("Retry-After"))
                  : undefined,
            },
          );
        }

        return payload as T;
      } catch (error) {
        if (error instanceof ApiError) throw error;
        if (timedOut) throw new ApiError("unavailable", "Request timed out");
        if (request.signal?.aborted || isAbortError(error)) {
          throw new ApiError("cancelled", "Request cancelled");
        }
        throw new ApiError("unavailable", "Unable to reach the API");
      } finally {
        clearTimeout(timeout);
        request.signal?.removeEventListener("abort", abort);
      }
    },
  };
}

async function parseBody(response: Response): Promise<unknown> {
  if (response.status === 204) return undefined;
  const text = await response.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function errorKind(status: number): ApiErrorKind {
  if (status === 401) return "unauthorized";
  if (status === 403) return "forbidden";
  if (status === 404) return "not_found";
  if (status === 405) return "method_not_allowed";
  if (status === 400 || status === 422) return "validation";
  if (status === 409) return "conflict";
  if (status === 429) return "rate_limited";
  if (status === 503) return "service_unavailable";
  if (status >= 500) return "unavailable";
  return "unknown";
}

function responseError(
  payload: unknown,
  fallback: string,
): ApiErrorMetadata & {
  message: string;
} {
  const body = errorBody(payload);
  return {
    message:
      typeof body?.message === "string"
        ? body.message
        : fallback || "API request failed",
    code: safeCode(body?.code),
    requestId: safeRequestID(body?.request_id),
    details: isRecord(body?.details) ? body.details : undefined,
  };
}

function errorBody(payload: unknown): Record<string, unknown> | undefined {
  if (isRecord(payload) && "error" in payload && isRecord(payload.error)) {
    return payload.error;
  }
  return undefined;
}

function responseRequestID(
  response: Response,
  bodyRequestID?: string,
): string | undefined {
  return safeRequestID(response.headers.get("X-Request-ID")) ?? bodyRequestID;
}

function retryAfterSeconds(value: string | null): number | undefined {
  if (value === null || !/^[1-9]\d{0,3}$/.test(value)) return undefined;
  const seconds = Number(value);
  return seconds <= 3_600 ? seconds : undefined;
}

function safeCode(value: unknown): string | undefined {
  return typeof value === "string" && /^[a-z][a-z0-9_]{0,63}$/.test(value)
    ? value
    : undefined;
}

function safeRequestID(value: unknown): string | undefined {
  return typeof value === "string" &&
    /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/.test(value)
    ? value
    : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}
