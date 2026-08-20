"use client";

import { ApiError } from "@persona-apps/api-client";
import {
    Alert,
    Badge,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    Skeleton,
} from "@persona-apps/ui";
import Link from "next/link";
import { useParams } from "next/navigation";

import { isCatalogueID } from "../api/catalogue";
import { CatalogueError } from "../features/catalogue/CatalogueError";
import {
    availabilityText,
    exactMinorPrice,
} from "../features/catalogue/presentation";
import { usePublicOffering } from "../features/catalogue/queries";
import { paths } from "../routes/paths";

export function OfferingDetailPage() {
    const offeringId = useParams<{ offeringId: string }>()?.offeringId ?? "";
    const validID = isCatalogueID(offeringId);
    const offering = usePublicOffering(offeringId, validID);
    const unavailable =
        !validID ||
        (offering.error instanceof ApiError &&
            offering.error.kind === "not_found");

    return (
        <section className="grid gap-8">
            <div>
                <Link
                    className="text-amber-800 underline underline-offset-4"
                    href={paths.offerings}
                >
                    Back to Offerings
                </Link>
                <h1
                    className="mt-4 text-4xl font-semibold"
                    tabIndex={-1}
                    autoFocus
                >
                    Offering details
                </h1>
            </div>

            {unavailable ? (
                <Alert>The requested catalogue item is unavailable.</Alert>
            ) : offering.isPending ? (
                <div aria-label="Loading Offering details" role="status">
                    <Skeleton className="h-72 w-full max-w-2xl" />
                    <span className="sr-only">Loading Offering details…</span>
                </div>
            ) : offering.isError ? (
                <CatalogueError
                    error={offering.error}
                    retry={() => void offering.refetch()}
                />
            ) : (
                <Card className="max-w-2xl">
                    <CardHeader>
                        <div className="flex flex-wrap items-center gap-3">
                            <h2 className="text-lg font-semibold leading-none tracking-tight">
                                {offering.data.name}
                            </h2>
                            <Badge variant="secondary">
                                {offering.data.offeringKind}
                            </Badge>
                        </div>
                        <CardDescription>
                            Code {offering.data.code}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="grid gap-5">
                        {offering.data.description ? (
                            <p className="whitespace-pre-wrap text-stone-700">
                                {offering.data.description}
                            </p>
                        ) : (
                            <p className="text-stone-600">
                                No public description is available.
                            </p>
                        )}
                        <dl className="grid gap-4 sm:grid-cols-2">
                            <div>
                                <dt className="text-sm text-stone-500">
                                    Exact price
                                </dt>
                                <dd className="font-medium">
                                    {exactMinorPrice(
                                        offering.data.currencyCode,
                                        offering.data.priceMinor,
                                    )}
                                </dd>
                            </div>
                            <div>
                                <dt className="text-sm text-stone-500">
                                    Participant capacity
                                </dt>
                                <dd className="font-medium">
                                    {offering.data.participantCapacity} per
                                    Purchase
                                </dd>
                            </div>
                        </dl>
                        <Alert>
                            {availabilityText(
                                offering.data.availableParticipantUnits,
                            )}
                            <span className="mt-1 block text-sm">
                                Final availability will be confirmed by a future
                                server-side checkout.
                            </span>
                        </Alert>
                    </CardContent>
                </Card>
            )}
        </section>
    );
}
