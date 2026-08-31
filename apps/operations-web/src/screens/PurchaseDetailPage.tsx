"use client";

import { ApiError } from "@persona-apps/api-client";
import {
    Alert,
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Skeleton,
} from "@persona-apps/ui";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";

import {
    isCanonicalUUID,
    type OperationsPartySummary,
    type OperationsPurchaseDetail,
} from "../api/purchases";
import { AccessBoundary } from "../features/events/AccessBoundary";
import {
    needsSessionRefresh,
    safeErrorMessage,
} from "../features/events/errors";
import { operationsQueryKeys } from "../features/events/queries";
import { usePurchase } from "../features/purchases/queries";
import { paths } from "../routes/paths";

export function PurchaseDetailPage() {
    const purchaseId = useParams<{ purchaseId: string }>()?.purchaseId ?? "";
    return (
        <AccessBoundary permission="purchase.read">
            {() =>
                isCanonicalUUID(purchaseId) ? (
                    <PurchaseDetailContent purchaseId={purchaseId} />
                ) : (
                    <InvalidPurchaseID />
                )
            }
        </AccessBoundary>
    );
}

function InvalidPurchaseID() {
    return (
        <section>
            <h1 className="text-3xl font-semibold" tabIndex={-1} autoFocus>
                Invalid Purchase ID
            </h1>
            <Alert variant="destructive" className="mt-5" role="alert">
                Purchase detail links require a canonical lowercase UUID. No
                Purchase request was sent.
            </Alert>
            <ReturnLink />
        </section>
    );
}

function PurchaseDetailContent({ purchaseId }: { purchaseId: string }) {
    const queryClient = useQueryClient();
    const purchase = usePurchase(purchaseId, true);

    useEffect(() => {
        if (needsSessionRefresh(purchase.error)) {
            void queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.session,
            });
        }
    }, [purchase.error, queryClient]);

    if (purchase.isPending) {
        return (
            <div className="grid gap-4" aria-label="Loading Purchase">
                <Skeleton className="h-10 w-72" />
                <Skeleton className="h-48 w-full" />
                <Skeleton className="h-48 w-full" />
            </div>
        );
    }

    if (purchase.isError) {
        const notFound =
            purchase.error instanceof ApiError &&
            purchase.error.kind === "not_found";
        return (
            <section>
                <h1 className="text-3xl font-semibold" tabIndex={-1} autoFocus>
                    {notFound ? "Purchase not found" : "Purchase unavailable"}
                </h1>
                <Alert variant="destructive" className="mt-5" role="alert">
                    {safeErrorMessage(purchase.error)}
                    {!notFound ? (
                        <div className="mt-4 flex flex-wrap gap-3">
                            <Button
                                variant="outline"
                                onClick={() => purchase.refetch()}
                            >
                                Try again
                            </Button>
                            {needsSessionRefresh(purchase.error) ? (
                                <Button
                                    variant="outline"
                                    onClick={() =>
                                        queryClient.invalidateQueries({
                                            queryKey:
                                                operationsQueryKeys.session,
                                        })
                                    }
                                >
                                    Refresh session
                                </Button>
                            ) : null}
                        </div>
                    ) : null}
                </Alert>
                <ReturnLink />
            </section>
        );
    }

    return <PurchaseSnapshot purchase={purchase.data} />;
}

function PurchaseSnapshot({
    purchase,
}: {
    purchase: OperationsPurchaseDetail;
}) {
    const headingRef = useRef<HTMLHeadingElement>(null);

    useEffect(() => {
        headingRef.current?.focus();
    }, []);

    return (
        <section className="grid gap-8">
            <header>
                <ReturnLink className="mt-0" />
                <div className="mt-4 flex flex-wrap items-center gap-3">
                    <h1
                        ref={headingRef}
                        className="min-w-0 break-all text-4xl font-semibold"
                        tabIndex={-1}
                        autoFocus
                    >
                        {purchase.purchaseRef}
                    </h1>
                    <Badge
                        variant={
                            purchase.status === "CANCELLED"
                                ? "destructive"
                                : "secondary"
                        }
                    >
                        {statusLabel(purchase.status)}
                    </Badge>
                </div>
                <p className="mt-3 max-w-3xl text-slate-300">
                    Stored Operations Purchase facts. Historical commercial
                    values are displayed without joining live Offering data.
                </p>
            </header>

            <Card>
                <CardHeader>
                    <CardTitle>Commercial snapshot</CardTitle>
                    <CardDescription>
                        Canonical identity, captured Offering, and exact amount.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <dl className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
                        <DetailItem label="Purchase ID" value={purchase.id} />
                        <DetailItem label="Channel" value={purchase.channel} />
                        <DetailItem label="Event ID" value={purchase.eventId} />
                        <DetailItem
                            label="Offering snapshot"
                            value={purchase.offeringNameSnapshot}
                        />
                        <DetailItem
                            label="Offering ID"
                            value={purchase.offeringId}
                        />
                        <DetailItem
                            label="Intended participants"
                            value={String(purchase.participantCount)}
                        />
                        <DetailItem
                            label="Captured total"
                            value={`${purchase.totalAmountMinor.toLocaleString("en-US")} ${purchase.currencyCode} minor units`}
                        />
                    </dl>
                </CardContent>
            </Card>

            <section aria-labelledby="party-relationships-heading">
                <h2
                    id="party-relationships-heading"
                    className="text-2xl font-semibold"
                >
                    Party relationships
                </h2>
                <p className="mt-2 text-slate-300">
                    Purchaser and payer remain explicit, even when they refer to
                    the same Party.
                </p>
                <div className="mt-5 grid gap-4 md:grid-cols-2">
                    <PartyCard label="Purchaser" party={purchase.purchaser} />
                    {purchase.payer ? (
                        <PartyCard label="Payer" party={purchase.payer} />
                    ) : (
                        <Card>
                            <CardHeader>
                                <CardTitle>Payer</CardTitle>
                                <CardDescription>
                                    No payer relationship is present in this
                                    Operations response. It is not inferred from
                                    the purchaser.
                                </CardDescription>
                            </CardHeader>
                        </Card>
                    )}
                </div>
            </section>

            <Card>
                <CardHeader>
                    <CardTitle>Intended participant snapshots</CardTitle>
                    <CardDescription>
                        Stored names in the exact server sequence; these records
                        do not imply payment or Sohibul activation.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <ol className="grid gap-3">
                        {purchase.participants.map((participant) => (
                            <li
                                key={participant.sequenceNo}
                                className="rounded-md border border-border bg-muted px-4 py-3 text-card-foreground"
                            >
                                <span className="mr-3 text-sm text-muted-foreground">
                                    {participant.sequenceNo}.
                                </span>
                                <span className="font-medium">
                                    {participant.displayName}
                                </span>
                            </li>
                        ))}
                    </ol>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Lifecycle timestamps</CardTitle>
                </CardHeader>
                <CardContent>
                    <dl className="grid gap-5 sm:grid-cols-2">
                        <DetailItem
                            label="Created"
                            value={formatInstant(purchase.createdAt)}
                        />
                        <DetailItem
                            label="Updated"
                            value={formatInstant(purchase.updatedAt)}
                        />
                        {purchase.eligibleAt ? (
                            <DetailItem
                                label="Eligible"
                                value={formatInstant(purchase.eligibleAt)}
                            />
                        ) : null}
                        {purchase.cancelledAt ? (
                            <DetailItem
                                label="Cancelled"
                                value={formatInstant(purchase.cancelledAt)}
                            />
                        ) : null}
                        {purchase.cancellationReason ? (
                            <DetailItem
                                label="Cancellation reason"
                                value={purchase.cancellationReason}
                            />
                        ) : null}
                    </dl>
                </CardContent>
            </Card>
        </section>
    );
}

function PartyCard({
    label,
    party,
}: {
    label: string;
    party: OperationsPartySummary;
}) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>{label}</CardTitle>
                <CardDescription>{party.displayName}</CardDescription>
            </CardHeader>
            <CardContent>
                <dl className="grid gap-3 text-sm">
                    <DetailItem label="Party ID" value={party.id} />
                    {party.email ? (
                        <DetailItem label="Email" value={party.email} />
                    ) : null}
                    {party.phone ? (
                        <DetailItem label="Phone" value={party.phone} />
                    ) : null}
                </dl>
            </CardContent>
        </Card>
    );
}

function DetailItem({ label, value }: { label: string; value: string }) {
    return (
        <div className="min-w-0">
            <dt className="text-sm text-muted-foreground">{label}</dt>
            <dd className="break-words font-medium text-card-foreground">
                {value}
            </dd>
        </div>
    );
}

function ReturnLink({ className = "mt-6" }: { className?: string }) {
    return (
        <Link
            className={`${className} inline-block text-cyan-300 underline underline-offset-4`}
            href={paths.purchasing}
        >
            Return to Purchases
        </Link>
    );
}

function statusLabel(status: string): string {
    const value = status.toLowerCase().replaceAll("_", " ");
    return value.charAt(0).toUpperCase() + value.slice(1);
}

function formatInstant(value: string): string {
    return new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
    }).format(new Date(value));
}
