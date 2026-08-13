import { describe, expect, it } from "vitest";

import { createApiProxy } from "./vite.config";

describe("storefront Vite API proxy", () => {
  it("preserves canonical API paths and proxies infrastructure probes", () => {
    const proxy = createApiProxy("http://127.0.0.1:8080");

    expect(proxy["/api"]).toEqual({
      target: "http://127.0.0.1:8080",
      changeOrigin: true,
    });
    expect(proxy["/health"]).toEqual({
      target: "http://127.0.0.1:8080",
      changeOrigin: true,
    });
    expect(proxy["/ready"]).toEqual({
      target: "http://127.0.0.1:8080",
      changeOrigin: true,
    });
    expect(proxy["/api"]).not.toHaveProperty("rewrite");
  });
});
