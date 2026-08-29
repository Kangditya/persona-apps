"use client";

import { ApiError } from "@persona-apps/api-client";
import { Alert, Button, Skeleton } from "@persona-apps/ui";
import { usePathname } from "next/navigation";
import { type ReactNode, useEffect, useState } from "react";

import type { OperationsSession } from "../../api/session";
import { operationsLoginHref } from "../../api/session";
import { useOperationsSession } from "./queries";

type AccessBoundaryProps = {
    permission: string;
    children: (session: OperationsSession) => ReactNode;
};

export function AccessBoundary({ permission, children }: AccessBoundaryProps) {
    const pathname = usePathname() ?? "";
    const [returnTo, setReturnTo] = useState(pathname);
    const session = useOperationsSession();

    useEffect(() => {
        setReturnTo(`${pathname}${window.location.search}`);
    }, [pathname]);

    if (session.isPending) {
        return (
            <section
                aria-label="Loading Operations session"
                className="grid gap-4"
            >
                <Skeleton className="h-8 w-64" />
                <Skeleton className="h-24 w-full" />
            </section>
        );
    }
    if (
        session.error instanceof ApiError &&
        session.error.kind === "unauthorized"
    ) {
        return (
            <section>
                <h1 className="text-3xl font-semibold" tabIndex={-1} autoFocus>
                    Sign in required
                </h1>
                <p className="mt-3 text-slate-300">
                    Use your authorized operator identity to continue.
                </p>
                <a
                    className="mt-6 inline-flex rounded-md bg-cyan-400 px-4 py-2 font-medium text-slate-950"
                    href={operationsLoginHref(returnTo)}
                >
                    Sign in with the identity provider
                </a>
            </section>
        );
    }
    if (session.isError) {
        return (
            <Alert variant="destructive" role="alert">
                The Operations session could not be loaded.
                <Button
                    variant="outline"
                    className="ml-3"
                    onClick={() => session.refetch()}
                >
                    Try again
                </Button>
            </Alert>
        );
    }
    if (!session.data.permissions.includes(permission)) {
        return (
            <section>
                <h1 className="text-3xl font-semibold" tabIndex={-1} autoFocus>
                    Access forbidden
                </h1>
                <Alert variant="destructive" className="mt-5">
                    Your current operator session does not include{" "}
                    <code>{permission}</code>.
                </Alert>
            </section>
        );
    }
    return children(session.data);
}
