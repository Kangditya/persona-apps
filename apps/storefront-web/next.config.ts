import type { NextConfig } from "next";

type Rewrite = { source: string; destination: string };

export const serviceWorkerRewrite: Rewrite = {
  source: "/sw.js",
  destination: "/_next/static/sw.js",
};

export const serviceWorkerHeaders = {
  source: "/sw.js",
  headers: [
    {
      key: "Content-Type",
      value: "application/javascript; charset=utf-8",
    },
    { key: "Cache-Control", value: "no-cache, no-store, must-revalidate" },
    {
      key: "Content-Security-Policy",
      value: "default-src 'self'; script-src 'self'",
    },
    { key: "Service-Worker-Allowed", value: "/" },
    { key: "X-Content-Type-Options", value: "nosniff" },
  ],
};

export function createApiRewrites(target = process.env.API_PROXY_TARGET) {
  if (!target) return [];

  const url = new URL(target);
  if (
    !["http:", "https:"].includes(url.protocol) ||
    url.username ||
    url.password ||
    url.pathname !== "/" ||
    url.search ||
    url.hash
  ) {
    throw new Error(
      "API_PROXY_TARGET must be an HTTP(S) origin without credentials or a path",
    );
  }

  return [
    { source: "/api/:path*", destination: `${url.origin}/api/:path*` },
    { source: "/health", destination: `${url.origin}/health` },
    { source: "/ready", destination: `${url.origin}/ready` },
  ] satisfies Rewrite[];
}

const nextConfig: NextConfig = {
  transpilePackages: ["@persona-apps/api-client", "@persona-apps/ui"],
  async headers() {
    return [serviceWorkerHeaders];
  },
  async rewrites() {
    return [serviceWorkerRewrite, ...createApiRewrites()];
  },
};

export default nextConfig;
