import { readFileSync } from "node:fs";
import vm from "node:vm";

import { describe, expect, it, vi } from "vitest";

const source = readFileSync(
  new URL("./service-worker.js", import.meta.url),
  "utf8",
)
  .replace("__CACHE_NAME__", '"test-shell"')
  .replace("__PRECACHE_URLS__", '["/_next/static/chunk.js", "/offline.html"]');

describe("storefront service-worker policy", () => {
  it("handles only immutable shell assets and data-free navigation fallback", () => {
    const listeners = new Map<string, (event: unknown) => void>();
    const respondWith = vi.fn();
    const context = {
      URL,
      caches: {
        open: vi.fn(),
        keys: vi.fn(),
        delete: vi.fn(),
        match: vi.fn(() => Promise.resolve(undefined)),
      },
      fetch: vi.fn(() => Promise.resolve(new Response())),
      self: {
        location: { origin: "https://storefront.example.test" },
        clients: { claim: vi.fn() },
        skipWaiting: vi.fn(),
        addEventListener: (type: string, listener: (event: unknown) => void) =>
          listeners.set(type, listener),
      },
    };
    vm.runInNewContext(source, context);
    const dispatchFetch = (method: string, pathname: string, mode = "cors") => {
      respondWith.mockClear();
      listeners.get("fetch")?.({
        request: {
          method,
          mode,
          url: `https://storefront.example.test${pathname}`,
        },
        respondWith,
      });
      return respondWith;
    };

    expect(
      dispatchFetch("GET", "/api/public/v1/events/active"),
    ).not.toHaveBeenCalled();
    expect(
      dispatchFetch("GET", "/api/operations/v1/auth/callback"),
    ).not.toHaveBeenCalled();
    expect(dispatchFetch("GET", "/health")).not.toHaveBeenCalled();
    expect(dispatchFetch("GET", "/ready")).not.toHaveBeenCalled();
    expect(dispatchFetch("POST", "/purchase")).not.toHaveBeenCalled();
    expect(
      dispatchFetch("GET", "/_next/static/chunk.js"),
    ).toHaveBeenCalledOnce();
    expect(
      dispatchFetch("GET", "/offerings", "navigate"),
    ).toHaveBeenCalledOnce();
    expect(source).not.toContain("cache.put");
  });
});
