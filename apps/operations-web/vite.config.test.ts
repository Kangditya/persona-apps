import { describe, expect, it } from "vitest";

import { createApiProxy } from "./vite.config";

describe("Operations development proxy", () => {
  it("forwards canonical API and diagnostic paths without rewriting them", () => {
    const proxy = createApiProxy("http://127.0.0.1:8080");

    expect(proxy).toEqual({
      "/api": { target: "http://127.0.0.1:8080", changeOrigin: true },
      "/health": { target: "http://127.0.0.1:8080", changeOrigin: true },
      "/ready": { target: "http://127.0.0.1:8080", changeOrigin: true },
    });
    expect("rewrite" in proxy["/api"]).toBe(false);
  });
});
