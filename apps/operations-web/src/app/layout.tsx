import type { Metadata, Viewport } from "next";
import type { ReactNode } from "react";

import { OperationsLayout } from "../layouts/OperationsLayout";
import "../styles/global.css";
import { Providers } from "./providers";

export const metadata: Metadata = {
    title: "Qurban Operations",
    description: "Qurban operations application",
    manifest: "/manifest.webmanifest",
};

export const viewport: Viewport = {
    themeColor: "#0f172a",
};

export default function RootLayout({ children }: { children: ReactNode }) {
    return (
        <html lang="en">
            <body>
                <Providers>
                    <OperationsLayout>{children}</OperationsLayout>
                </Providers>
            </body>
        </html>
    );
}
