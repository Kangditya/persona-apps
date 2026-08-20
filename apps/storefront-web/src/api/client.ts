import { createApiClient, type ApiRequest } from "@persona-apps/api-client";

export type ApiProvider = "api" | "development";
export type ApiHealth = { status: "ok" | "development" };

type StorefrontApiOptions = {
  baseUrl: string;
  provider?: ApiProvider;
  fetch?: typeof fetch;
};

export function createStorefrontApi({
  baseUrl,
  provider = "api",
  fetch,
}: StorefrontApiOptions) {
  const client = createApiClient({ baseUrl, fetch });

  return {
    request: <T>(path: string, request: ApiRequest = {}): Promise<T> =>
      client.request<T>(path, { ...request, credentials: "omit" }),
    diagnostics: {
      health: (request?: ApiRequest): Promise<ApiHealth> =>
        provider === "development"
          ? Promise.resolve({ status: "development" })
          : client.request<ApiHealth>("/health", {
              ...request,
              credentials: "omit",
            }),
    },
  };
}

export type StorefrontApi = ReturnType<typeof createStorefrontApi>;

const provider: ApiProvider =
  process.env.NEXT_PUBLIC_API_PROVIDER === "development"
    ? "development"
    : "api";

export const storefrontApi = createStorefrontApi({
  baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || "",
  provider,
});
