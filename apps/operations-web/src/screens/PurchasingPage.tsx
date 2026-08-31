"use client";

import { useQueryClient } from "@tanstack/react-query";
import {
    Alert,
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Field,
    Input,
    Skeleton,
} from "@persona-apps/ui";
import Link from "next/link";
import { type FormEvent, useEffect, useRef, useState } from "react";

import {
    isCanonicalUUID,
    purchaseFilterStatuses,
    type OperationsPurchaseSummary,
    type PurchaseFilterStatus,
    type PurchaseFilters,
    type PurchaseStatus,
} from "../api/purchases";
import { AccessBoundary } from "../features/events/AccessBoundary";
import {
    needsSessionRefresh,
    safeErrorMessage,
} from "../features/events/errors";
import { operationsQueryKeys } from "../features/events/queries";
import { usePurchases } from "../features/purchases/queries";
import { purchasePath } from "../routes/paths";

const emptyFilters: PurchaseFilters = { eventId: "", status: "" };

export function PurchasingPage() {
    return (
        <AccessBoundary permission="purchase.read">
            {() => <PurchasingContent />}
        </AccessBoundary>
    );
}

function PurchasingContent() {
    const queryClient = useQueryClient();
    const [eventId, setEventId] = useState("");
    const [status, setStatus] = useState<PurchaseFilterStatus | "">("");
    const [filters, setFilters] = useState<PurchaseFilters>(emptyFilters);
    const [filterError, setFilterError] = useState("");
    const [cursorHistory, setCursorHistory] = useState([""]);
    const cursor = cursorHistory.at(-1) ?? "";
    const purchases = usePurchases(filters, cursor, true);
    const headingRef = useRef<HTMLHeadingElement>(null);
    const filterErrorRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        headingRef.current?.focus();
    }, []);

    useEffect(() => {
        if (needsSessionRefresh(purchases.error)) {
            void queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.session,
            });
        }
    }, [purchases.error, queryClient]);

    useEffect(() => {
        if (filterError) filterErrorRef.current?.focus();
    }, [filterError]);

    function applyFilters(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const normalizedEventId = eventId.trim();
        if (normalizedEventId && !isCanonicalUUID(normalizedEventId)) {
            setFilterError(
                "Event ID must be a canonical lowercase UUID before a request can be sent.",
            );
            return;
        }
        setFilterError("");
        setFilters({ eventId: normalizedEventId, status });
        setCursorHistory([""]);
    }

    function clearFilters() {
        setEventId("");
        setStatus("");
        setFilterError("");
        setFilters(emptyFilters);
        setCursorHistory([""]);
    }

    return (
        <section className="grid gap-8">
            <header>
                <p className="font-medium text-cyan-300">Purchasing records</p>
                <h1
                    ref={headingRef}
                    className="mt-2 text-4xl font-semibold"
                    tabIndex={-1}
                    autoFocus
                >
                    Purchases
                </h1>
                <p className="mt-3 max-w-3xl text-slate-300">
                    Read authoritative Purchase snapshots. Filters and page
                    navigation match the current Operations API exactly.
                </p>
            </header>

            <Card>
                <CardHeader>
                    <CardTitle>Filter purchases</CardTitle>
                    <CardDescription>
                        The API currently supports only an exact Event UUID and
                        Purchase status.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {filterError ? (
                        <div ref={filterErrorRef} tabIndex={-1}>
                            <Alert
                                id="purchase-filter-error"
                                variant="destructive"
                                role="alert"
                                className="mb-5"
                            >
                                {filterError}
                            </Alert>
                        </div>
                    ) : null}
                    <form
                        className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(12rem,0.45fr)_auto] lg:items-end"
                        onSubmit={applyFilters}
                        noValidate
                    >
                        <Field
                            id="purchase-event-id"
                            label="Event ID"
                            description="Optional canonical lowercase UUID."
                        >
                            <Input
                                name="event_id"
                                value={eventId}
                                onChange={(event) =>
                                    setEventId(event.currentTarget.value)
                                }
                                autoComplete="off"
                                spellCheck={false}
                                aria-invalid={filterError ? true : undefined}
                                aria-describedby={
                                    filterError
                                        ? "purchase-filter-error"
                                        : undefined
                                }
                            />
                        </Field>
                        <Field id="purchase-status" label="Status">
                            <select
                                name="status"
                                value={status}
                                onChange={(event) =>
                                    setStatus(
                                        event.currentTarget.value as
                                            PurchaseFilterStatus | "",
                                    )
                                }
                                className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                            >
                                <option value="">All supported statuses</option>
                                {purchaseFilterStatuses.map((value) => (
                                    <option key={value} value={value}>
                                        {statusLabel(value)}
                                    </option>
                                ))}
                            </select>
                        </Field>
                        <div className="flex flex-wrap gap-3">
                            <Button type="submit">Apply filters</Button>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={clearFilters}
                            >
                                Clear
                            </Button>
                        </div>
                    </form>
                </CardContent>
            </Card>

            <section aria-labelledby="purchase-results-heading">
                <div className="flex flex-wrap items-end justify-between gap-3">
                    <div>
                        <h2
                            id="purchase-results-heading"
                            className="text-2xl font-semibold"
                        >
                            Purchase records
                        </h2>
                        <p className="mt-2 text-sm text-slate-300">
                            Page {cursorHistory.length}; no background page or
                            per-row detail requests.
                        </p>
                    </div>
                    {filters.eventId || filters.status ? (
                        <p className="text-sm text-cyan-200" role="status">
                            Filters applied
                            {filters.status
                                ? ` · ${statusLabel(filters.status)}`
                                : ""}
                            {filters.eventId ? " · exact Event" : ""}
                        </p>
                    ) : null}
                </div>

                {purchases.isPending ? (
                    <div
                        className="mt-5 grid gap-3"
                        aria-label="Loading Purchases"
                    >
                        <Skeleton className="h-14 w-full" />
                        <Skeleton className="h-14 w-full" />
                        <Skeleton className="h-14 w-full" />
                    </div>
                ) : purchases.isError ? (
                    <Alert variant="destructive" className="mt-5" role="alert">
                        {safeErrorMessage(purchases.error)}
                        <div className="mt-4 flex flex-wrap gap-3">
                            <Button
                                variant="outline"
                                onClick={() => purchases.refetch()}
                            >
                                Try again
                            </Button>
                            {needsSessionRefresh(purchases.error) ? (
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
                    </Alert>
                ) : purchases.data.data.length === 0 ? (
                    <Alert className="mt-5">
                        No Purchases match this page and filter combination.
                    </Alert>
                ) : (
                    <PurchaseResults purchases={purchases.data.data} />
                )}

                {purchases.isSuccess ? (
                    <nav
                        className="mt-5 flex flex-wrap gap-3"
                        aria-label="Purchase pagination"
                    >
                        <Button
                            variant="outline"
                            disabled={cursorHistory.length === 1}
                            onClick={() =>
                                setCursorHistory((current) =>
                                    current.slice(0, -1),
                                )
                            }
                        >
                            Previous page
                        </Button>
                        <Button
                            variant="outline"
                            disabled={!purchases.data.nextCursor}
                            onClick={() => {
                                if (purchases.data.nextCursor) {
                                    setCursorHistory((current) => [
                                        ...current,
                                        purchases.data.nextCursor ?? "",
                                    ]);
                                }
                            }}
                        >
                            Next page
                        </Button>
                    </nav>
                ) : null}
            </section>
        </section>
    );
}

function PurchaseResults({
    purchases,
}: {
    purchases: readonly OperationsPurchaseSummary[];
}) {
    return (
        <div className="mt-5">
            <div className="hidden overflow-x-auto rounded-lg border border-slate-800 bg-slate-900 lg:block">
                <table className="w-full text-left text-sm">
                    <caption className="sr-only">
                        Bounded Operations Purchase records
                    </caption>
                    <thead className="bg-slate-800/70 text-slate-200">
                        <tr>
                            <th className="px-4 py-3 font-medium" scope="col">
                                Reference
                            </th>
                            <th className="px-4 py-3 font-medium" scope="col">
                                Status
                            </th>
                            <th className="px-4 py-3 font-medium" scope="col">
                                Offering
                            </th>
                            <th className="px-4 py-3 font-medium" scope="col">
                                Purchaser
                            </th>
                            <th className="px-4 py-3 font-medium" scope="col">
                                Amount
                            </th>
                            <th className="px-4 py-3 font-medium" scope="col">
                                Created
                            </th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800">
                        {purchases.map((purchase) => (
                            <tr key={purchase.id}>
                                <th
                                    className="px-4 py-4 font-medium"
                                    scope="row"
                                >
                                    <Link
                                        className="break-all text-cyan-300 underline-offset-4 hover:underline"
                                        href={purchasePath(purchase.id)}
                                    >
                                        {purchase.purchaseRef}
                                    </Link>
                                    <span className="mt-1 block font-normal text-slate-400">
                                        Event {purchase.eventId}
                                    </span>
                                </th>
                                <td className="px-4 py-4">
                                    <PurchaseStatusBadge
                                        status={purchase.status}
                                    />
                                </td>
                                <td className="px-4 py-4">
                                    {purchase.offeringNameSnapshot}
                                </td>
                                <td className="px-4 py-4">
                                    {purchase.purchaser.displayName}
                                </td>
                                <td className="whitespace-nowrap px-4 py-4">
                                    {formatAmount(purchase)}
                                </td>
                                <td className="whitespace-nowrap px-4 py-4 text-slate-300">
                                    {formatInstant(purchase.createdAt)}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="grid gap-4 lg:hidden">
                {purchases.map((purchase) => (
                    <Card key={purchase.id}>
                        <CardHeader>
                            <div className="flex flex-wrap items-center justify-between gap-3">
                                <CardTitle>
                                    <Link
                                        className="break-all text-cyan-300 underline-offset-4 hover:underline"
                                        href={purchasePath(purchase.id)}
                                    >
                                        {purchase.purchaseRef}
                                    </Link>
                                </CardTitle>
                                <PurchaseStatusBadge status={purchase.status} />
                            </div>
                            <CardDescription>
                                Created {formatInstant(purchase.createdAt)}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <dl className="grid gap-3 text-sm sm:grid-cols-2">
                                <SummaryItem
                                    label="Offering snapshot"
                                    value={purchase.offeringNameSnapshot}
                                />
                                <SummaryItem
                                    label="Purchaser"
                                    value={purchase.purchaser.displayName}
                                />
                                <SummaryItem
                                    label="Amount"
                                    value={formatAmount(purchase)}
                                />
                                <SummaryItem
                                    label="Event ID"
                                    value={purchase.eventId}
                                />
                            </dl>
                        </CardContent>
                    </Card>
                ))}
            </div>
        </div>
    );
}

function SummaryItem({ label, value }: { label: string; value: string }) {
    return (
        <div className="min-w-0">
            <dt className="text-muted-foreground">{label}</dt>
            <dd className="break-words font-medium text-card-foreground">
                {value}
            </dd>
        </div>
    );
}

function PurchaseStatusBadge({ status }: { status: PurchaseStatus }) {
    return (
        <Badge variant={status === "CANCELLED" ? "destructive" : "secondary"}>
            {statusLabel(status)}
        </Badge>
    );
}

function statusLabel(status: PurchaseStatus): string {
    const value = status.toLowerCase().replaceAll("_", " ");
    return value.charAt(0).toUpperCase() + value.slice(1);
}

function formatAmount(purchase: OperationsPurchaseSummary): string {
    return `${purchase.totalAmountMinor.toLocaleString("en-US")} ${purchase.currencyCode} minor units`;
}

function formatInstant(value: string): string {
    return new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
    }).format(new Date(value));
}
