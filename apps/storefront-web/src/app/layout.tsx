import type { Metadata, Viewport } from "next";
import type { ReactNode } from "react";

import { StorefrontLayout } from "../layouts/StorefrontLayout";
import "../styles/global.css";
import { Providers } from "./providers";

export const metadata: Metadata = {
    title: "Qurban Storefront",
    description: "Qurban public storefront",
    manifest: "/manifest.webmanifest",
};

export const viewport: Viewport = {
    themeColor: "#0f766e",
};

export default function RootLayout({ children }: { children: ReactNode }) {
    return (
        <html lang="en">
            <body>
                <Providers>
                    <StorefrontLayout>{children}</StorefrontLayout>
                </Providers>
            </body>
        </html>
    );
}
