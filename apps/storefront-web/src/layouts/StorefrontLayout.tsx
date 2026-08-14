import { NavLink, Outlet } from "react-router";

import { paths } from "../routes/paths";
import { StorefrontPwaStatus } from "../pwa/StorefrontPwaStatus";

const links = [
    [paths.home, "Event Home"],
    [paths.offerings, "Offerings"],
    [paths.purchaseTracking, "Purchase Tracking"],
] as const;

export function StorefrontLayout() {
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
                        <NavLink
                            key={to}
                            to={to}
                            className={({ isActive }) =>
                                isActive
                                    ? "text-amber-700"
                                    : "text-stone-600 hover:text-stone-950"
                            }
                        >
                            {label}
                        </NavLink>
                    ))}
                </nav>
            </header>
            <StorefrontPwaStatus />
            <main className="mx-auto max-w-5xl px-6 py-16">
                <Outlet />
            </main>
        </div>
    );
}
