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
          id: "/operations-web",
          name: "Qurban Operations",
          short_name: "Qurban Operations",
          description: "Internal Qurban event operations application",
          start_url: "/",
          scope: "/",
          display: "standalone",
          theme_color: "#0f172a",
          background_color: "#0f172a",
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
        workbox: {
          cleanupOutdatedCaches: true,
          globPatterns: ["**/*.{js,css,html}"],
          navigateFallback: "index.html",
          navigateFallbackDenylist: [/^\/api\//],
        },
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
