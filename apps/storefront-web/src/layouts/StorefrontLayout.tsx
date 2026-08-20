"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { StorefrontPwaStatus } from "../pwa/StorefrontPwaStatus";
import { isActivePath, paths } from "../routes/paths";

const links = [
    [paths.home, "Event Home"],
    [paths.offerings, "Offerings"],
    [paths.purchaseTracking, "Purchase Tracking"],
] as const;

export function StorefrontLayout({ children }: { children: ReactNode }) {
    const pathname = usePathname() ?? "";

    return (
        <div className="min-h-screen bg-amber-50 text-stone-900">
            <header className="border-b border-amber-200 bg-white">
                <nav
                    aria-label="Storefront navigation"
                    className="mx-auto flex max-w-5xl flex-wrap items-center gap-4 px-6 py-4"
                >
                    <strong className="w-full sm:mr-auto sm:w-auto">
                        Qurban Storefront
                    </strong>
                    {links.map(([to, label]) => (
                        <Link
                            key={to}
                            href={to}
                            className={
                                isActivePath(pathname, to)
                                    ? "text-amber-700"
                                    : "text-stone-600 hover:text-stone-950"
                            }
                        >
                            {label}
                        </Link>
                    ))}
                </nav>
            </header>
            <StorefrontPwaStatus />
            <main className="mx-auto max-w-5xl px-6 py-16">{children}</main>
        </div>
    );
}
