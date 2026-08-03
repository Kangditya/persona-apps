export type ApiErrorKind =
  | "cancelled"
  | "unauthorized"
  | "forbidden"
  | "validation"
  | "conflict"
  | "unavailable"
  | "unknown";

export class ApiError extends Error {
  readonly kind: ApiErrorKind;
  readonly status?: number;

  constructor(kind: ApiErrorKind, message: string, status?: number) {
    super(message);
    this.name = "ApiError";
    this.kind = kind;
    this.status = status;
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
          throw new ApiError(
            errorKind(response.status),
            errorMessage(payload, response.statusText),
            response.status,
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
  if (status === 400 || status === 422) return "validation";
  if (status === 409) return "conflict";
  if (status === 429 || status >= 500) return "unavailable";
  return "unknown";
}

function errorMessage(payload: unknown, fallback: string): string {
  if (
    typeof payload === "object" &&
    payload !== null &&
    "error" in payload &&
    typeof payload.error === "object" &&
    payload.error !== null &&
    "message" in payload.error &&
    typeof payload.error.message === "string"
  ) {
    return payload.error.message;
  }
  return fallback || "API request failed";
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}
