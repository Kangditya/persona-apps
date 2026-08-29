"use client";

import { ApiError } from "@persona-apps/api-client";
import {
    Alert,
    Badge,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Skeleton,
} from "@persona-apps/ui";
import Link from "next/link";

import type { PublicOffering } from "../api/catalogue";
import { CatalogueError } from "../features/catalogue/CatalogueError";
import {
    availabilityText,
    exactMinorPrice,
} from "../features/catalogue/presentation";
import {
    useActiveEvent,
    usePublicOfferings,
} from "../features/catalogue/queries";
import { offeringPath } from "../routes/paths";

export function OfferingsPage() {
    const event = useActiveEvent();
    const offerings = usePublicOfferings(event.data?.id ?? "");

    return (
        <section className="grid gap-8">
            <div>
                <p className="font-medium text-amber-700">Storefront</p>
                <h1
                    className="mt-2 text-4xl font-semibold"
                    tabIndex={-1}
                    autoFocus
                >
                    Qurban Offerings
                </h1>
                <p className="mt-4 max-w-2xl text-stone-600">
                    Published Offerings for the active Event. Availability is
                    advisory and does not reserve participant capacity.
                </p>
            </div>

            {event.isPending ? (
                <LoadingCatalogue label="Loading active Event" />
            ) : event.isError ? (
                <CatalogueError
                    error={event.error}
                    retry={() => void event.refetch()}
                />
            ) : event.data === null ? (
                <Alert>
                    No active Qurban Event is available, so there is no public
                    Offering catalogue right now.
                </Alert>
            ) : (
                <section aria-labelledby="published-offerings-heading">
                    <h2
                        id="published-offerings-heading"
                        className="text-2xl font-semibold"
                    >
                        {event.data.eventYear} — {event.data.name}
                    </h2>
                    {offerings.isPending ? (
                        <div className="mt-5">
                            <LoadingCatalogue label="Loading published Offerings" />
                        </div>
                    ) : offerings.isError ? (
                        <div className="mt-5">
                            {offerings.error instanceof ApiError &&
                            offerings.error.kind === "not_found" ? (
                                <Alert>
                                    This active Event is no longer available.
                                </Alert>
                            ) : (
                                <CatalogueError
                                    error={offerings.error}
                                    retry={() => void offerings.refetch()}
                                />
                            )}
                        </div>
                    ) : offerings.data.data.length === 0 ? (
                        <Alert className="mt-5">
                            This Event has no published Offerings yet.
                        </Alert>
                    ) : (
                        <>
                            <div className="mt-5 grid gap-5 md:grid-cols-2">
                                {offerings.data.data.map((offering) => (
                                    <OfferingCard
                                        key={offering.id}
                                        offering={offering}
                                    />
                                ))}
                            </div>
                            {offerings.data.nextCursor ? (
                                <Alert className="mt-5">
                                    This Week 2 view shows the first 100
                                    published Offerings.
                                </Alert>
                            ) : null}
                        </>
                    )}
                </section>
            )}
        </section>
    );
}

function OfferingCard({ offering }: { offering: PublicOffering }) {
    return (
        <Card>
            <CardHeader>
                <div className="flex flex-wrap items-center gap-3">
                    <CardTitle>{offering.name}</CardTitle>
                    <Badge variant="secondary">{offering.offeringKind}</Badge>
                </div>
                <CardDescription>Code {offering.code}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-3">
                {offering.description ? (
                    <p className="line-clamp-3 text-sm text-stone-600">
                        {offering.description}
                    </p>
                ) : null}
                <p className="font-medium">
                    {exactMinorPrice(
                        offering.currencyCode,
                        offering.priceMinor,
                    )}
                </p>
                <p className="text-sm text-stone-600">
                    Up to {offering.participantCapacity} participants per
                    Purchase.
                </p>
                <p className="text-sm text-stone-600">
                    {availabilityText(offering.availableParticipantUnits)}
                </p>
                <Link
                    className="inline-flex min-h-11 w-fit items-center text-amber-800 underline underline-offset-4"
                    href={offeringPath(offering.id)}
                >
                    View {offering.name} details
                </Link>
            </CardContent>
        </Card>
    );
}

function LoadingCatalogue({ label }: { label: string }) {
    return (
        <div className="grid gap-3" aria-label={label} role="status">
            <Skeleton className="h-36 w-full" />
            <Skeleton className="h-36 w-full" />
            <span className="sr-only">{label}…</span>
        </div>
    );
}
