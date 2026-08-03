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
    diagnostics: {
      health: (request?: ApiRequest): Promise<ApiHealth> =>
        provider === "development"
          ? Promise.resolve({ status: "development" })
          : client.request<ApiHealth>("/health", request),
    },
  };
}

const provider: ApiProvider =
  import.meta.env.VITE_API_PROVIDER === "development" ? "development" : "api";

export const storefrontApi = createStorefrontApi({
  baseUrl: import.meta.env.VITE_API_BASE_URL || "/api",
  provider,
});
