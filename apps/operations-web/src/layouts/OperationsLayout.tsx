"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@persona-apps/ui";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { ReactNode } from "react";

import { sessionApi } from "../api/session";
import {
    needsSessionRefresh,
    safeErrorMessage,
} from "../features/events/errors";
import {
    operationsQueryKeys,
    useOperationsSession,
} from "../features/events/queries";
import { OperationsPwaStatus } from "../pwa/OperationsPwaStatus";
import { isActivePath, paths } from "../routes/paths";

const links = [
    [paths.operatorLogin, "Operator Login"],
    [paths.events, "Events"],
    [paths.eventDashboard, "Event Dashboard"],
    [paths.purchasing, "Purchasing"],
    [paths.paymentVerification, "Payment Verification"],
] as const;

export function OperationsLayout({ children }: { children: ReactNode }) {
    const session = useOperationsSession();
    const queryClient = useQueryClient();
    const pathname = usePathname() ?? "";
    const router = useRouter();
    const logout = useMutation({
        mutationFn: () => sessionApi.logout(session.data?.csrfToken ?? ""),
        onSuccess: () => {
            queryClient.removeQueries({
                queryKey: operationsQueryKeys.privateRoot,
            });
            router.push(paths.operatorLogin);
        },
    });

    return (
        <div className="min-h-screen bg-slate-950 text-slate-100">
            <header className="border-b border-slate-800 bg-slate-900">
                <nav
                    aria-label="Operations navigation"
                    className="mx-auto flex max-w-5xl flex-wrap items-center gap-4 px-6 py-4"
                >
                    <strong className="w-full sm:mr-auto sm:w-auto">
                        Qurban Operations
                    </strong>
                    {links.map(([to, label]) => (
                        <Link
                            key={to}
                            href={to}
                            className={
                                isActivePath(pathname, to)
                                    ? "text-cyan-300"
                                    : "text-slate-300 hover:text-white"
                            }
                        >
                            {label}
                        </Link>
                    ))}
                    {session.isSuccess ? (
                        <div className="flex w-full flex-wrap items-center justify-between gap-3 sm:ml-auto sm:w-auto">
                            <span className="text-sm text-slate-300">
                                {session.data.operator.displayName}
                            </span>
                            <Button
                                size="sm"
                                variant="outline"
                                loading={logout.isPending}
                                onClick={() => logout.mutate()}
                            >
                                Log out
                            </Button>
                        </div>
                    ) : null}
                </nav>
                {logout.isError ? (
                    <p
                        className="mx-auto max-w-5xl px-6 pb-3 text-sm text-red-300"
                        role="alert"
                    >
                        {safeErrorMessage(logout.error)}
                        {needsSessionRefresh(logout.error) ? (
                            <Button
                                size="sm"
                                variant="outline"
                                className="ml-3"
                                onClick={() => session.refetch()}
                            >
                                Refresh session
                            </Button>
                        ) : null}
                    </p>
                ) : null}
            </header>
            <OperationsPwaStatus />
            <main className="mx-auto max-w-5xl px-6 py-16">{children}</main>
        </div>
    );
}
