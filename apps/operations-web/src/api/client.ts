import { createApiClient, type ApiRequest } from "@persona-apps/api-client";

export type ApiProvider = "api" | "development";
export type ApiHealth = { status: "ok" | "development" };

type OperationsApiOptions = {
  baseUrl: string;
  provider?: ApiProvider;
  fetch?: typeof fetch;
};

export type OperationsApi = ReturnType<typeof createOperationsApi>;

export function createOperationsApi({
  baseUrl,
  provider = "api",
  fetch,
}: OperationsApiOptions) {
  const client = createApiClient({ baseUrl, fetch });

  return {
    request: <T>(path: string, request: ApiRequest = {}): Promise<T> =>
      client.request<T>(path, { ...request, credentials: "include" }),
    diagnostics: {
      health: (request?: ApiRequest): Promise<ApiHealth> =>
        provider === "development"
          ? Promise.resolve({ status: "development" })
          : client.request<ApiHealth>("/health", request),
    },
  };
}

const provider: ApiProvider =
  process.env.NEXT_PUBLIC_API_PROVIDER === "development"
    ? "development"
    : "api";

export const operationsApi = createOperationsApi({
  baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || "",
  provider,
});
