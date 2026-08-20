"use client";

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

import type { PublicEvent } from "../api/catalogue";
import { CatalogueError } from "../features/catalogue/CatalogueError";
import { registrationPresentation } from "../features/catalogue/presentation";
import { useActiveEvent } from "../features/catalogue/queries";
import { paths } from "../routes/paths";

export function HomePage() {
    const event = useActiveEvent();

    return (
        <section className="grid gap-8">
            <div>
                <p className="font-medium text-amber-700">Qurban Event</p>
                <h1
                    className="mt-2 text-4xl font-semibold"
                    tabIndex={-1}
                    autoFocus
                >
                    Event landing
                </h1>
                <p className="mt-4 max-w-2xl text-stone-600">
                    Discover the currently active Qurban Event and its public
                    Offerings.
                </p>
            </div>

            {event.isPending ? (
                <div aria-label="Loading active Event" role="status">
                    <Skeleton className="h-48 w-full max-w-2xl" />
                    <span className="sr-only">Loading active Event…</span>
                </div>
            ) : event.isError ? (
                <CatalogueError
                    error={event.error}
                    retry={() => void event.refetch()}
                />
            ) : event.data === null ? (
                <Alert>
                    No active Qurban Event is available right now. Please check
                    again later.
                </Alert>
            ) : (
                <EventCard event={event.data} />
            )}
        </section>
    );
}

function EventCard({ event }: { event: PublicEvent }) {
    const registration = registrationPresentation(event);
    return (
        <Card className="max-w-2xl">
            <CardHeader>
                <div className="flex flex-wrap items-center gap-3">
                    <h2 className="text-lg font-semibold leading-none tracking-tight">
                        {event.eventYear} — {event.name}
                    </h2>
                    <Badge>Active Event</Badge>
                </div>
                <CardDescription>
                    Public Event discovery remains available outside the
                    registration window.
                </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4">
                <div>
                    <p className="font-medium">{registration.label}</p>
                    {registration.details.map((detail) => (
                        <p className="text-sm text-stone-600" key={detail}>
                            {detail}
                        </p>
                    ))}
                    {registration.timeZone ? (
                        <p className="mt-1 text-sm text-stone-500">
                            Times shown in {registration.timeZone}.
                        </p>
                    ) : null}
                </div>
                <p className="text-sm text-stone-600">
                    Registration and availability are informational snapshots; a
                    future checkout will revalidate them on the server.
                </p>
                <Link
                    className="inline-flex min-h-11 w-fit items-center rounded-md bg-amber-700 px-4 py-2 font-medium text-white hover:bg-amber-800"
                    href={paths.offerings}
                >
                    Explore Offerings
                </Link>
            </CardContent>
        </Card>
    );
}
