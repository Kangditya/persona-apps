"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
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
import { type FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";

import { eventsApi, type CreateEventInput } from "../api/events";
import type { OperationsSession } from "../api/session";
import { AccessBoundary } from "../features/events/AccessBoundary";
import {
    needsSessionRefresh,
    safeErrorMessage,
} from "../features/events/errors";
import {
    browserTimeZone,
    exactInteger,
    FormValueError,
    localDateTimeToUTC,
    optionalExactInteger,
    utcPreview,
} from "../features/events/forms";
import { operationsQueryKeys, useEvents } from "../features/events/queries";
import { eventPath } from "../routes/paths";

type CreateIntent = {
    key: string;
    input: CreateEventInput;
};

export function EventsPage() {
    return (
        <AccessBoundary permission="event.read">
            {(session) => <EventsContent session={session} />}
        </AccessBoundary>
    );
}

function EventsContent({ session }: { session: OperationsSession }) {
    const queryClient = useQueryClient();
    const [cursorHistory, setCursorHistory] = useState([""]);
    const cursor = cursorHistory.at(-1) ?? "";
    const events = useEvents(cursor, true);
    const canManage = session.permissions.includes("event.manage");
    const [opensAt, setOpensAt] = useState("");
    const [closesAt, setClosesAt] = useState("");
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [intent, setIntent] = useState<CreateIntent | null>(null);
    const [notice, setNotice] = useState("");
    const errorSummary = useRef<HTMLDivElement>(null);
    const create = useMutation({
        mutationFn: (value: CreateIntent) =>
            eventsApi.createEvent(value.input, session.csrfToken, value.key),
        onSuccess: async (created) => {
            setIntent(null);
            setNotice(
                `Event ${created.name} was created at version ${created.version}.`,
            );
            setOpensAt("");
            setClosesAt("");
            await queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.eventLists,
            });
        },
    });

    useEffect(() => {
        if (needsSessionRefresh(events.error)) {
            void queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.session,
            });
        }
    }, [events.error, queryClient]);

    useEffect(() => {
        if (create.isError || Object.keys(errors).length > 0)
            errorSummary.current?.focus();
    }, [create.isError, errors]);

    function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const form = event.currentTarget;
        const data = new FormData(form);
        try {
            const name = String(data.get("name") ?? "").trim();
            if (name.length < 1 || name.length > 255) {
                throw new FormValueError(
                    "name",
                    "Name must contain 1 to 255 characters.",
                );
            }
            const registrationOpensAt = localDateTimeToUTC(
                "registration_opens_at",
                "Registration opens",
                opensAt,
            );
            const registrationClosesAt = localDateTimeToUTC(
                "registration_closes_at",
                "Registration closes",
                closesAt,
            );
            if (
                registrationOpensAt &&
                registrationClosesAt &&
                registrationOpensAt >= registrationClosesAt
            ) {
                throw new FormValueError(
                    "registration_closes_at",
                    "Registration closes must be after registration opens.",
                );
            }
            const nextIntent: CreateIntent = {
                key: crypto.randomUUID(),
                input: {
                    event_year: exactInteger(
                        "event_year",
                        "Event year",
                        data.get("event_year"),
                        1900,
                        9999,
                    ),
                    name,
                    registration_opens_at: registrationOpensAt,
                    registration_closes_at: registrationClosesAt,
                    participant_quota: optionalExactInteger(
                        "participant_quota",
                        "Participant quota",
                        data.get("participant_quota"),
                    ),
                },
            };
            setErrors({});
            setNotice("");
            setIntent(nextIntent);
            create.mutate(nextIntent, {
                onSuccess: () => form.reset(),
            });
        } catch (error) {
            if (error instanceof FormValueError) {
                setErrors({ [error.field]: error.message });
            } else {
                setErrors({ form: "The form could not be read." });
            }
        }
    }

    return (
        <section className="grid gap-10">
            <div>
                <p className="font-medium text-cyan-300">
                    Commerce configuration
                </p>
                <h1
                    className="mt-2 text-4xl font-semibold"
                    tabIndex={-1}
                    autoFocus
                >
                    Events
                </h1>
                <p className="mt-3 text-slate-300">
                    Bounded Operations records. The API remains authoritative
                    for lifecycle rules.
                </p>
            </div>

            <section aria-labelledby="event-list-heading">
                <h2 id="event-list-heading" className="text-2xl font-semibold">
                    Event records
                </h2>
                {events.isPending ? (
                    <div
                        className="mt-5 grid gap-3"
                        aria-label="Loading Events"
                    >
                        <Skeleton className="h-24 w-full" />
                        <Skeleton className="h-24 w-full" />
                    </div>
                ) : events.isError ? (
                    <Alert variant="destructive" className="mt-5">
                        {safeErrorMessage(events.error)}
                        <Button
                            variant="outline"
                            className="ml-3"
                            onClick={() => events.refetch()}
                        >
                            Try again
                        </Button>
                    </Alert>
                ) : events.data.data.length === 0 ? (
                    <Alert className="mt-5">
                        No Events are available on this page.
                    </Alert>
                ) : (
                    <div className="mt-5 grid gap-4">
                        {events.data.data.map((event) => (
                            <Card key={event.id}>
                                <CardHeader>
                                    <div className="flex flex-wrap items-center gap-3">
                                        <CardTitle>
                                            <Link
                                                className="text-cyan-300 underline-offset-4 hover:underline"
                                                href={eventPath(event.id)}
                                            >
                                                {event.eventYear} — {event.name}
                                            </Link>
                                        </CardTitle>
                                        <Badge variant="secondary">
                                            {event.status}
                                        </Badge>
                                    </div>
                                    <CardDescription>
                                        Version {event.version}; quota{" "}
                                        {event.participantQuota ?? "unbounded"}.
                                    </CardDescription>
                                </CardHeader>
                                <CardContent className="text-sm text-slate-300">
                                    Registration:{" "}
                                    {formatInstant(event.registrationOpensAt)}{" "}
                                    to{" "}
                                    {formatInstant(event.registrationClosesAt)}
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                )}
                {events.isSuccess ? (
                    <div
                        className="mt-5 flex gap-3"
                        aria-label="Event pagination"
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
                            disabled={!events.data.nextCursor}
                            onClick={() =>
                                events.data.nextCursor &&
                                setCursorHistory((current) => [
                                    ...current,
                                    events.data.nextCursor ?? "",
                                ])
                            }
                        >
                            Next page
                        </Button>
                    </div>
                ) : null}
            </section>

            {canManage ? (
                <Card>
                    <CardHeader>
                        <CardTitle>Create a draft Event</CardTitle>
                        <CardDescription>
                            Datetimes use {browserTimeZone()} locally and are
                            previewed as UTC.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {(Object.keys(errors).length > 0 || create.isError) && (
                            <div ref={errorSummary} tabIndex={-1}>
                                <Alert
                                    variant="destructive"
                                    role="alert"
                                    className="mb-5"
                                >
                                    {create.isError
                                        ? safeErrorMessage(create.error)
                                        : Object.values(errors)[0]}
                                </Alert>
                                {needsSessionRefresh(create.error) ? (
                                    <Button
                                        type="button"
                                        variant="outline"
                                        className="mb-5"
                                        onClick={() =>
                                            queryClient.invalidateQueries({
                                                queryKey:
                                                    operationsQueryKeys.session,
                                            })
                                        }
                                    >
                                        Refresh session before retrying
                                    </Button>
                                ) : null}
                            </div>
                        )}
                        <form
                            className="grid gap-5"
                            onSubmit={submit}
                            noValidate
                        >
                            <div className="grid gap-5 md:grid-cols-2">
                                <Field
                                    id="event-year"
                                    label="Event year"
                                    required
                                    error={errors.event_year}
                                >
                                    <Input
                                        name="event_year"
                                        inputMode="numeric"
                                        autoComplete="off"
                                    />
                                </Field>
                                <Field
                                    id="event-name"
                                    label="Name"
                                    required
                                    error={errors.name}
                                >
                                    <Input name="name" maxLength={255} />
                                </Field>
                                <Field
                                    id="registration-opens"
                                    label="Registration opens"
                                    error={errors.registration_opens_at}
                                    description={`UTC preview: ${utcPreview(opensAt)}`}
                                >
                                    <Input
                                        name="registration_opens_at"
                                        type="datetime-local"
                                        value={opensAt}
                                        onChange={(event) =>
                                            setOpensAt(event.target.value)
                                        }
                                    />
                                </Field>
                                <Field
                                    id="registration-closes"
                                    label="Registration closes"
                                    error={errors.registration_closes_at}
                                    description={`UTC preview: ${utcPreview(closesAt)}`}
                                >
                                    <Input
                                        name="registration_closes_at"
                                        type="datetime-local"
                                        value={closesAt}
                                        onChange={(event) =>
                                            setClosesAt(event.target.value)
                                        }
                                    />
                                </Field>
                            </div>
                            <Field
                                id="participant-quota"
                                label="Participant quota"
                                error={errors.participant_quota}
                                description="Optional total participant-unit quota. Leave empty for unbounded."
                            >
                                <Input
                                    name="participant_quota"
                                    inputMode="numeric"
                                    autoComplete="off"
                                />
                            </Field>
                            <div className="flex flex-wrap items-center gap-3">
                                <Button
                                    type="submit"
                                    loading={create.isPending}
                                >
                                    Create Event
                                </Button>
                                {create.isError && intent ? (
                                    <Button
                                        type="button"
                                        variant="outline"
                                        disabled={create.isPending}
                                        onClick={() => create.mutate(intent)}
                                    >
                                        Retry the same request
                                    </Button>
                                ) : null}
                                <span
                                    role="status"
                                    aria-live="polite"
                                    className="text-sm text-slate-300"
                                >
                                    {notice}
                                </span>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            ) : (
                <Alert>
                    Creating Events requires the event.manage permission.
                </Alert>
            )}
        </section>
    );
}

function formatInstant(value: string | null): string {
    return value ? new Date(value).toLocaleString() : "not set";
}
