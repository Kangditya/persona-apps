import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
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
        purpose: "maskable",
      },
    ],
  };
}
