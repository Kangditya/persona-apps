import { describe, expect, it } from "vitest";

import {
  createApiRewrites,
  serviceWorkerHeaders,
  serviceWorkerRewrite,
} from "./next.config";

describe("Operations Next.js runtime configuration", () => {
  it("preserves canonical API and probe paths when a local proxy is enabled", () => {
    expect(createApiRewrites("http://127.0.0.1:8081")).toEqual([
      {
        source: "/api/:path*",
        destination: "http://127.0.0.1:8081/api/:path*",
      },
      {
        source: "/health",
        destination: "http://127.0.0.1:8081/health",
      },
      {
        source: "/ready",
        destination: "http://127.0.0.1:8081/ready",
      },
    ]);
    expect(createApiRewrites("")).toEqual([]);
    expect(() => createApiRewrites("https://user:secret@example.test")).toThrow(
      /without credentials/,
    );
  });

  it("serves the generated worker at root scope without HTTP caching", () => {
    expect(serviceWorkerRewrite).toEqual({
      source: "/sw.js",
      destination: "/_next/static/sw.js",
    });
    expect(serviceWorkerHeaders.headers).toContainEqual({
      key: "Service-Worker-Allowed",
      value: "/",
    });
    expect(serviceWorkerHeaders.headers).toContainEqual({
      key: "Cache-Control",
      value: "no-cache, no-store, must-revalidate",
    });
  });
});
