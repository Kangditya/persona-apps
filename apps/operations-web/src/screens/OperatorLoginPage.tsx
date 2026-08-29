"use client";

import { ApiError } from "@persona-apps/api-client";
import {
    Alert,
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Skeleton,
} from "@persona-apps/ui";
import Link from "next/link";

import { operationsLoginHref } from "../api/session";
import { safeErrorMessage } from "../features/events/errors";
import { useOperationsSession } from "../features/events/queries";
import { paths } from "../routes/paths";

export function OperatorLoginPage() {
    const session = useOperationsSession();

    return (
        <section>
            <p className="font-medium text-cyan-300">Identity & Access</p>
            <h1 className="mt-2 text-4xl font-semibold" tabIndex={-1} autoFocus>
                Operator session
            </h1>
            {session.isPending ? (
                <div
                    className="mt-6 grid gap-3"
                    aria-label="Loading operator session"
                >
                    <Skeleton className="h-5 w-52" />
                    <Skeleton className="h-28 w-full" />
                </div>
            ) : session.error instanceof ApiError &&
              session.error.kind === "unauthorized" ? (
                <Card className="mt-8 max-w-2xl">
                    <CardHeader>
                        <CardTitle>Sign in required</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-slate-300">
                            Authentication uses the configured OpenID Connect
                            provider and a secure server session.
                        </p>
                        <a
                            className="mt-5 inline-flex rounded-md bg-cyan-400 px-4 py-2 font-medium text-slate-950"
                            href={operationsLoginHref(paths.events)}
                        >
                            Sign in with the identity provider
                        </a>
                    </CardContent>
                </Card>
            ) : session.isError ? (
                <Alert variant="destructive" className="mt-6">
                    {safeErrorMessage(session.error)}
                    <Button
                        variant="outline"
                        className="ml-3"
                        onClick={() => session.refetch()}
                    >
                        Try again
                    </Button>
                </Alert>
            ) : (
                <Card className="mt-8 max-w-2xl">
                    <CardHeader>
                        <CardTitle>
                            Signed in as {session.data.operator.displayName}
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-slate-300">
                            Session expires{" "}
                            {new Date(session.data.expiresAt).toLocaleString()}.
                        </p>
                        <Link
                            className="mt-5 inline-block text-cyan-300 underline"
                            href={paths.events}
                        >
                            Open Event administration
                        </Link>
                    </CardContent>
                </Card>
            )}
        </section>
    );
}
