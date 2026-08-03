import { ApiError, createApiClient } from "@persona-apps/api-client";
import { describe, expect, it } from "vitest";

import { createStorefrontApi } from "./client";

describe("storefront API boundary", () => {
  it("uses the configured base URL and parses the API response", async () => {
    let url = "";
    let headers: Headers | undefined;
    const api = createStorefrontApi({
      baseUrl: "https://api.example.test/",
      fetch: async (input, init) => {
        url = String(input);
        headers = new Headers(init?.headers);
        return Response.json({ status: "ok" });
      },
    });

    await expect(api.diagnostics.health()).resolves.toEqual({ status: "ok" });
    expect(url).toBe("https://api.example.test/health");
    expect(headers?.get("Authorization")).toBeNull();
  });

  it("keeps the development provider deterministic and non-networked", async () => {
    const api = createStorefrontApi({
      baseUrl: "https://api.example.test",
      provider: "development",
      fetch: () => {
        throw new Error("development provider must not call fetch");
      },
    });

    await expect(api.diagnostics.health()).resolves.toEqual({
      status: "development",
    });
  });

  it("normalizes authorization and cancellation errors", async () => {
    const unauthorized = createApiClient({
      baseUrl: "https://api.example.test",
      fetch: async () =>
        Response.json(
          { error: { message: "Sign in required" } },
          { status: 401 },
        ),
    });
    await expect(unauthorized.request("/health")).rejects.toMatchObject({
      kind: "unauthorized",
      message: "Sign in required",
    } satisfies Partial<ApiError>);

    const controller = new AbortController();
    controller.abort();
    const cancelled = createApiClient({
      baseUrl: "https://api.example.test",
      fetch: async () => {
        throw new DOMException("Aborted", "AbortError");
      },
    });
    await expect(
      cancelled.request("/health", { signal: controller.signal }),
    ).rejects.toMatchObject({ kind: "cancelled" } satisfies Partial<ApiError>);
  });

  it("serializes requests and turns timeouts into unavailable errors", async () => {
    let request: RequestInit | undefined;
    const serialized = createApiClient({
      baseUrl: "https://api.example.test",
      getHeaders: () => ({ "X-Session": "test-session" }),
      fetch: async (_input, init) => {
        request = init;
        return Response.json({ accepted: true });
      },
    });

    await serialized.request("/diagnostic", {
      method: "POST",
      body: { check: true },
    });
    expect(request?.method).toBe("POST");
    expect(request?.body).toBe('{"check":true}');
    expect(new Headers(request?.headers).get("X-Session")).toBe("test-session");

    const timedOut = createApiClient({
      baseUrl: "https://api.example.test",
      timeoutMs: 1,
      fetch: (_input, init) =>
        new Promise((_resolve, reject) => {
          init?.signal?.addEventListener(
            "abort",
            () => reject(new DOMException("Aborted", "AbortError")),
            { once: true },
          );
        }),
    });
    await expect(timedOut.request("/health")).rejects.toMatchObject({
      kind: "unavailable",
      message: "Request timed out",
    } satisfies Partial<ApiError>);
  });
});
