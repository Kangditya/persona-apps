import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { VitePWA } from "vite-plugin-pwa";

export function createApiProxy(target: string) {
  return {
    "/api": { target, changeOrigin: true },
    "/health": { target, changeOrigin: true },
    "/ready": { target, changeOrigin: true },
  };
}

export const storefrontWorkbox = {
  cleanupOutdatedCaches: true,
  globPatterns: ["**/*.{js,css,html}"],
  navigateFallback: "index.html",
  navigateFallbackDenylist: [/^\/api\//],
};

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");

  return {
    plugins: [
      react(),
      tailwindcss(),
      VitePWA({
        registerType: "prompt",
        includeAssets: ["icon.svg", "icon-192.svg", "icon-512.svg"],
        manifest: {
          id: "/storefront-web",
          name: "Qurban Storefront",
          short_name: "Qurban Storefront",
          description: "Public Qurban event and offering application",
          start_url: "/",
          scope: "/",
          display: "standalone",
          theme_color: "#0f766e",
          background_color: "#fafaf9",
          icons: [
            {
              src: "/icon-192.svg",
              sizes: "192x192",
              type: "image/svg+xml",
              purpose: "any",
            },
            {
              src: "/icon-512.svg",
              sizes: "512x512",
              type: "image/svg+xml",
              purpose: "any maskable",
            },
          ],
        },
        workbox: storefrontWorkbox,
        devOptions: { enabled: false },
      }),
    ],
    server: {
      proxy: createApiProxy(
        env.VITE_API_PROXY_TARGET || "http://127.0.0.1:8080",
      ),
    },
  };
});
