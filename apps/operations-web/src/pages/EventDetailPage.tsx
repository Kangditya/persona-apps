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
    Textarea,
} from "@persona-apps/ui";
import { type FormEvent, useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router";

import {
    eventsApi,
    type CreateOfferingInput,
    type EventTransition,
    type OperationsEvent,
    type PatchEventInput,
} from "../api/events";
import type { OperationsSession } from "../api/session";
import { AccessBoundary } from "../features/events/AccessBoundary";
import {
    isStaleVersion,
    needsSessionRefresh,
    safeErrorMessage,
} from "../features/events/errors";
import {
    browserTimeZone,
    eventTransitions,
    exactInteger,
    FormValueError,
    localDateTimeToUTC,
    maximumCapacity,
    optionalExactInteger,
    utcPreview,
    utcToLocalDateTime,
} from "../features/events/forms";
import {
    operationsQueryKeys,
    useEvent,
    useOfferings,
} from "../features/events/queries";
import { offeringPath, paths } from "../routes/paths";

type LifecycleIntent = {
    action: EventTransition;
    expectedVersion: number;
    key: string;
};

type OfferingIntent = {
    input: CreateOfferingInput;
    key: string;
};

export function EventDetailPage() {
    const { eventId = "" } = useParams();
    return (
        <AccessBoundary permission="event.read">
            {(session) => (
                <EventDetailContent eventId={eventId} session={session} />
            )}
        </AccessBoundary>
    );
}

function EventDetailContent({
    eventId,
    session,
}: {
    eventId: string;
    session: OperationsSession;
}) {
    const queryClient = useQueryClient();
    const event = useEvent(eventId, true);
    useEffect(() => {
        if (needsSessionRefresh(event.error)) {
            void queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.session,
            });
        }
    }, [event.error, queryClient]);
    if (event.isPending) {
        return (
            <div className="grid gap-4" aria-label="Loading Event">
                <Skeleton className="h-10 w-72" />
                <Skeleton className="h-48 w-full" />
            </div>
        );
    }
    if (event.isError) {
        return (
            <section>
                <h1 className="text-3xl font-semibold" tabIndex={-1} autoFocus>
                    Event unavailable
                </h1>
                <Alert variant="destructive" className="mt-5">
                    {safeErrorMessage(event.error)}
                    <Button
                        variant="outline"
                        className="ml-3"
                        onClick={() => event.refetch()}
                    >
                        Try again
                    </Button>
                </Alert>
                <Link
                    className="mt-6 inline-block text-cyan-300 underline"
                    to={paths.events}
                >
                    Return to Events
                </Link>
            </section>
        );
    }

    return (
        <section className="grid gap-10">
            <header>
                <Link
                    className="text-sm text-cyan-300 underline"
                    to={paths.events}
                >
                    Events
                </Link>
                <div className="mt-3 flex flex-wrap items-center gap-3">
                    <h1
                        className="text-4xl font-semibold"
                        tabIndex={-1}
                        autoFocus
                    >
                        {event.data.name}
                    </h1>
                    <Badge variant="secondary">{event.data.status}</Badge>
                </div>
                <p className="mt-3 text-slate-300">
                    Event {event.data.eventYear}; version {event.data.version}.
                    Created {formatInstant(event.data.createdAt)}.
                </p>
            </header>
            <EventConfiguration
                value={event.data}
                session={session}
                reload={() => event.refetch()}
            />
            <OfferingCatalogue event={event.data} session={session} />
        </section>
    );
}

function EventConfiguration({
    value,
    session,
    reload,
}: {
    value: OperationsEvent;
    session: OperationsSession;
    reload: () => unknown;
}) {
    const queryClient = useQueryClient();
    const canManage = session.permissions.includes("event.manage");
    const canEdit = ["DRAFT", "PUBLISHED", "SUSPENDED"].includes(value.status);
    const [opensAt, setOpensAt] = useState(
        utcToLocalDateTime(value.registrationOpensAt),
    );
    const [closesAt, setClosesAt] = useState(
        utcToLocalDateTime(value.registrationClosesAt),
    );
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [notice, setNotice] = useState("");
    const [lifecycleIntent, setLifecycleIntent] =
        useState<LifecycleIntent | null>(null);
    const errorSummary = useRef<HTMLDivElement>(null);
    const patch = useMutation({
        mutationFn: (input: PatchEventInput) =>
            eventsApi.patchEvent(value.id, input, session.csrfToken),
        onSuccess: async (updated) => {
            setNotice(
                `Event configuration saved at version ${updated.version}.`,
            );
            await refreshEventQueries(queryClient, value.id);
        },
    });
    const transition = useMutation({
        mutationFn: (intent: LifecycleIntent) =>
            eventsApi.transitionEvent(
                value.id,
                intent.action,
                intent.expectedVersion,
                session.csrfToken,
                intent.key,
            ),
        onSuccess: async (updated) => {
            setLifecycleIntent(null);
            setNotice(
                `Event is now ${updated.status} at version ${updated.version}.`,
            );
            await refreshEventQueries(queryClient, value.id);
        },
    });

    useEffect(() => {
        if (
            patch.isError ||
            transition.isError ||
            Object.keys(errors).length > 0
        ) {
            errorSummary.current?.focus();
        }
    }, [errors, patch.isError, transition.isError]);

    function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const data = new FormData(event.currentTarget);
        try {
            const input: PatchEventInput = { expected_version: value.version };
            const name = String(data.get("name") ?? "").trim();
            if (name.length < 1 || name.length > 255) {
                throw new FormValueError(
                    "name",
                    "Name must contain 1 to 255 characters.",
                );
            }
            if (name !== value.name) input.name = name;
            if (opensAt !== utcToLocalDateTime(value.registrationOpensAt)) {
                input.registration_opens_at = localDateTimeToUTC(
                    "registration_opens_at",
                    "Registration opens",
                    opensAt,
                );
            }
            if (closesAt !== utcToLocalDateTime(value.registrationClosesAt)) {
                input.registration_closes_at = localDateTimeToUTC(
                    "registration_closes_at",
                    "Registration closes",
                    closesAt,
                );
            }
            const opens =
                "registration_opens_at" in input
                    ? input.registration_opens_at
                    : value.registrationOpensAt;
            const closes =
                "registration_closes_at" in input
                    ? input.registration_closes_at
                    : value.registrationClosesAt;
            if (opens && closes && opens >= closes) {
                throw new FormValueError(
                    "registration_closes_at",
                    "Registration closes must be after registration opens.",
                );
            }
            const quotaRaw = String(data.get("participant_quota") ?? "").trim();
            const originalQuota =
                value.participantQuota === null
                    ? ""
                    : String(value.participantQuota);
            if (quotaRaw !== originalQuota) {
                input.participant_quota = optionalExactInteger(
                    "participant_quota",
                    "Participant quota",
                    quotaRaw,
                );
            }
            if (Object.keys(input).length === 1) {
                throw new FormValueError(
                    "form",
                    "Change at least one configuration field.",
                );
            }
            setErrors({});
            setNotice("");
            patch.mutate(input);
        } catch (error) {
            setErrors(
                error instanceof FormValueError
                    ? { [error.field]: error.message }
                    : { form: "The form could not be read." },
            );
        }
    }

    function runTransition(action: EventTransition) {
        if (
            ["activate", "close", "archive"].includes(action) &&
            !window.confirm(`Confirm Event action: ${action}?`)
        ) {
            return;
        }
        const intent = {
            action,
            expectedVersion: value.version,
            key: crypto.randomUUID(),
        };
        setNotice("");
        setLifecycleIntent(intent);
        transition.mutate(intent);
    }

    const mutationError = patch.error ?? transition.error;
    return (
        <Card>
            <CardHeader>
                <CardTitle>Event configuration and lifecycle</CardTitle>
                <CardDescription>
                    Local datetime editor: {browserTimeZone()}. The displayed
                    server version is {value.version}.
                </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-6">
                {(mutationError || Object.keys(errors).length > 0) && (
                    <div ref={errorSummary} tabIndex={-1}>
                        <Alert variant="destructive" role="alert">
                            {mutationError
                                ? safeErrorMessage(mutationError)
                                : Object.values(errors)[0]}
                        </Alert>
                        {needsSessionRefresh(mutationError) ? (
                            <Button
                                variant="outline"
                                className="mt-3"
                                onClick={() =>
                                    queryClient.invalidateQueries({
                                        queryKey: operationsQueryKeys.session,
                                    })
                                }
                            >
                                Refresh session before retrying
                            </Button>
                        ) : null}
                        {isStaleVersion(mutationError) ? (
                            <Button
                                variant="outline"
                                className="mt-3"
                                onClick={() => reload()}
                            >
                                Reload latest server version
                            </Button>
                        ) : null}
                    </div>
                )}
                <dl className="grid gap-3 text-sm sm:grid-cols-2">
                    <div>
                        <dt className="text-slate-400">Participant quota</dt>
                        <dd>{value.participantQuota ?? "Unbounded"}</dd>
                    </div>
                    <div>
                        <dt className="text-slate-400">Updated</dt>
                        <dd>{formatInstant(value.updatedAt)}</dd>
                    </div>
                </dl>
                {canManage && canEdit ? (
                    <form className="grid gap-5" onSubmit={submit} noValidate>
                        <Field
                            id="edit-event-name"
                            label="Name"
                            required
                            error={errors.name}
                        >
                            <Input
                                name="name"
                                defaultValue={value.name}
                                maxLength={255}
                            />
                        </Field>
                        <div className="grid gap-5 md:grid-cols-2">
                            <Field
                                id="edit-registration-opens"
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
                                id="edit-registration-closes"
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
                            id="edit-event-quota"
                            label="Participant quota"
                            error={errors.participant_quota}
                            description="Leave empty to clear the total quota."
                        >
                            <Input
                                name="participant_quota"
                                inputMode="numeric"
                                defaultValue={value.participantQuota ?? ""}
                            />
                        </Field>
                        <Button type="submit" loading={patch.isPending}>
                            Save configuration
                        </Button>
                    </form>
                ) : (
                    <Alert>
                        {!canManage
                            ? "Editing requires the event.manage permission."
                            : "Configuration is immutable in the current Event state."}
                    </Alert>
                )}
                {canManage ? (
                    <div>
                        <h3 className="font-semibold">Lifecycle actions</h3>
                        <div className="mt-3 flex flex-wrap gap-3">
                            {eventTransitions(value.status).map((action) => (
                                <Button
                                    key={action}
                                    type="button"
                                    variant={
                                        action === "archive"
                                            ? "destructive"
                                            : "outline"
                                    }
                                    disabled={transition.isPending}
                                    onClick={() => runTransition(action)}
                                >
                                    {eventActionLabel(action)}
                                </Button>
                            ))}
                            {transition.isError && lifecycleIntent ? (
                                <Button
                                    type="button"
                                    variant="outline"
                                    disabled={transition.isPending}
                                    onClick={() =>
                                        transition.mutate(lifecycleIntent)
                                    }
                                >
                                    Retry the same action
                                </Button>
                            ) : null}
                        </div>
                    </div>
                ) : null}
                <p
                    role="status"
                    aria-live="polite"
                    className="text-sm text-cyan-200"
                >
                    {notice}
                </p>
            </CardContent>
        </Card>
    );
}

function OfferingCatalogue({
    event,
    session,
}: {
    event: OperationsEvent;
    session: OperationsSession;
}) {
    const queryClient = useQueryClient();
    const canRead = session.permissions.includes("offering.read");
    const canManage = session.permissions.includes("offering.manage");
    const canCreate =
        canManage && !["CLOSED", "ARCHIVED"].includes(event.status);
    const [cursorHistory, setCursorHistory] = useState([""]);
    const cursor = cursorHistory.at(-1) ?? "";
    const offerings = useOfferings(event.id, cursor, canRead);
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [intent, setIntent] = useState<OfferingIntent | null>(null);
    const [notice, setNotice] = useState("");
    const errorSummary = useRef<HTMLDivElement>(null);
    const create = useMutation({
        mutationFn: (value: OfferingIntent) =>
            eventsApi.createOffering(
                event.id,
                value.input,
                session.csrfToken,
                value.key,
            ),
        onSuccess: async (created) => {
            setIntent(null);
            setNotice(
                `Offering ${created.name} was created at version ${created.version}.`,
            );
            await queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.offeringLists(event.id),
            });
        },
    });

    useEffect(() => {
        if (needsSessionRefresh(offerings.error)) {
            void queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.session,
            });
        }
    }, [offerings.error, queryClient]);

    useEffect(() => {
        if (create.isError || Object.keys(errors).length > 0)
            errorSummary.current?.focus();
    }, [create.isError, errors]);

    function submit(eventTarget: FormEvent<HTMLFormElement>) {
        eventTarget.preventDefault();
        const form = eventTarget.currentTarget;
        const data = new FormData(form);
        try {
            const code = requiredText("code", "Code", data.get("code"), 64);
            const name = requiredText("name", "Name", data.get("name"), 255);
            const kind = requiredText(
                "offering_kind",
                "Offering kind",
                data.get("offering_kind"),
                64,
            );
            const descriptionText = String(
                data.get("description") ?? "",
            ).trim();
            if (descriptionText.length > 10_000) {
                throw new FormValueError(
                    "description",
                    "Description must not exceed 10,000 characters.",
                );
            }
            const currency = String(data.get("currency_code") ?? "")
                .trim()
                .toUpperCase();
            if (!/^[A-Z]{3}$/.test(currency)) {
                throw new FormValueError(
                    "currency_code",
                    "Currency must be a three-letter uppercase code.",
                );
            }
            const nextIntent: OfferingIntent = {
                key: crypto.randomUUID(),
                input: {
                    code,
                    name,
                    offering_kind: kind,
                    description: descriptionText || null,
                    price_minor: exactInteger(
                        "price_minor",
                        "Price minor units",
                        data.get("price_minor"),
                        0,
                        Number.MAX_SAFE_INTEGER,
                    ),
                    currency_code: currency,
                    participant_capacity: exactInteger(
                        "participant_capacity",
                        "Participant capacity",
                        data.get("participant_capacity"),
                        1,
                        maximumCapacity,
                    ),
                    participant_quota: optionalExactInteger(
                        "offering_participant_quota",
                        "Offering participant quota",
                        data.get("participant_quota"),
                    ),
                },
            };
            setErrors({});
            setNotice("");
            setIntent(nextIntent);
            create.mutate(nextIntent, { onSuccess: () => form.reset() });
        } catch (error) {
            setErrors(
                error instanceof FormValueError
                    ? { [error.field]: error.message }
                    : { form: "The form could not be read." },
            );
        }
    }

    return (
        <section aria-labelledby="offerings-heading" className="grid gap-6">
            <div>
                <h2 id="offerings-heading" className="text-2xl font-semibold">
                    Offerings
                </h2>
                <p className="mt-2 text-slate-300">
                    Availability is advisory. Mutation controls still rely on
                    the server.
                </p>
            </div>
            {!canRead ? (
                <Alert variant="destructive">
                    Listing Offerings requires the offering.read permission.
                </Alert>
            ) : offerings.isPending ? (
                <Skeleton className="h-28 w-full" />
            ) : offerings.isError ? (
                <Alert variant="destructive">
                    {safeErrorMessage(offerings.error)}
                    <Button
                        variant="outline"
                        className="ml-3"
                        onClick={() => offerings.refetch()}
                    >
                        Try again
                    </Button>
                </Alert>
            ) : offerings.data.data.length === 0 ? (
                <Alert>No Offerings are available on this page.</Alert>
            ) : (
                <div className="grid gap-4 md:grid-cols-2">
                    {offerings.data.data.map((offering) => (
                        <Card key={offering.id}>
                            <CardHeader>
                                <CardTitle>
                                    <Link
                                        className="text-cyan-300 underline"
                                        to={offeringPath(event.id, offering.id)}
                                    >
                                        {offering.code} — {offering.name}
                                    </Link>
                                </CardTitle>
                                <CardDescription>
                                    {offering.status}; version{" "}
                                    {offering.version}
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="text-sm text-slate-300">
                                {offering.priceMinor} {offering.currencyCode}{" "}
                                minor units;{" "}
                                {availabilityLabel(
                                    offering.availableParticipantUnits,
                                )}
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}
            {offerings.isSuccess ? (
                <div className="flex gap-3" aria-label="Offering pagination">
                    <Button
                        variant="outline"
                        disabled={cursorHistory.length === 1}
                        onClick={() =>
                            setCursorHistory((current) => current.slice(0, -1))
                        }
                    >
                        Previous page
                    </Button>
                    <Button
                        variant="outline"
                        disabled={!offerings.data.nextCursor}
                        onClick={() =>
                            offerings.data.nextCursor &&
                            setCursorHistory((current) => [
                                ...current,
                                offerings.data.nextCursor ?? "",
                            ])
                        }
                    >
                        Next page
                    </Button>
                </div>
            ) : null}

            {canCreate ? (
                <Card>
                    <CardHeader>
                        <CardTitle>Create a draft Offering</CardTitle>
                        <CardDescription>
                            Price is entered and displayed as exact minor units.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {(create.isError || Object.keys(errors).length > 0) && (
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
                            </div>
                        )}
                        <form
                            className="grid gap-5"
                            onSubmit={submit}
                            noValidate
                        >
                            <div className="grid gap-5 md:grid-cols-2">
                                <Field
                                    id="offering-code"
                                    label="Code"
                                    required
                                    error={errors.code}
                                >
                                    <Input
                                        name="code"
                                        maxLength={64}
                                        autoComplete="off"
                                    />
                                </Field>
                                <Field
                                    id="offering-name"
                                    label="Name"
                                    required
                                    error={errors.name}
                                >
                                    <Input name="name" maxLength={255} />
                                </Field>
                                <Field
                                    id="offering-kind"
                                    label="Offering kind"
                                    required
                                    error={errors.offering_kind}
                                    description="A bounded commercial label, for example SHARE."
                                >
                                    <Input
                                        name="offering_kind"
                                        maxLength={64}
                                    />
                                </Field>
                                <Field
                                    id="offering-currency"
                                    label="Currency"
                                    required
                                    error={errors.currency_code}
                                >
                                    <Input
                                        name="currency_code"
                                        defaultValue="IDR"
                                        maxLength={3}
                                    />
                                </Field>
                                <Field
                                    id="offering-price"
                                    label="Price minor units"
                                    required
                                    error={errors.price_minor}
                                >
                                    <Input
                                        name="price_minor"
                                        inputMode="numeric"
                                    />
                                </Field>
                                <Field
                                    id="offering-capacity"
                                    label="Participant capacity"
                                    required
                                    error={errors.participant_capacity}
                                >
                                    <Input
                                        name="participant_capacity"
                                        inputMode="numeric"
                                    />
                                </Field>
                            </div>
                            <Field
                                id="offering-description"
                                label="Description"
                                error={errors.description}
                            >
                                <Textarea
                                    name="description"
                                    maxLength={10_000}
                                />
                            </Field>
                            <Field
                                id="offering-quota"
                                label="Offering participant quota"
                                error={errors.offering_participant_quota}
                                description="Optional total quota. Leave empty for no Offering-specific cap."
                            >
                                <Input
                                    name="participant_quota"
                                    inputMode="numeric"
                                />
                            </Field>
                            <div className="flex flex-wrap items-center gap-3">
                                <Button
                                    type="submit"
                                    loading={create.isPending}
                                >
                                    Create Offering
                                </Button>
                                {create.isError && intent ? (
                                    <Button
                                        type="button"
                                        variant="outline"
                                        onClick={() => create.mutate(intent)}
                                    >
                                        Retry the same request
                                    </Button>
                                ) : null}
                                <span
                                    role="status"
                                    aria-live="polite"
                                    className="text-sm text-cyan-200"
                                >
                                    {notice}
                                </span>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            ) : (
                <Alert>
                    {!canManage
                        ? "Creating Offerings requires the offering.manage permission."
                        : "New Offerings cannot be added to a terminal Event."}
                </Alert>
            )}
        </section>
    );
}

async function refreshEventQueries(
    queryClient: ReturnType<typeof useQueryClient>,
    eventId: string,
) {
    await Promise.all([
        queryClient.invalidateQueries({
            queryKey: operationsQueryKeys.eventDetail(eventId),
        }),
        queryClient.invalidateQueries({
            queryKey: operationsQueryKeys.eventLists,
        }),
    ]);
}

function requiredText(
    field: string,
    label: string,
    raw: FormDataEntryValue | null,
    maximum: number,
): string {
    const value = typeof raw === "string" ? raw.trim() : "";
    if (value.length < 1 || value.length > maximum) {
        throw new FormValueError(
            field,
            `${label} must contain 1 to ${maximum} characters.`,
        );
    }
    return value;
}

function eventActionLabel(action: EventTransition): string {
    return {
        publish: "Publish",
        activate: "Activate",
        suspend: "Suspend",
        close: "Close",
        archive: "Archive",
    }[action];
}

function availabilityLabel(value: number | null): string {
    if (value === null) return "unbounded availability";
    if (value === 0) return "zero units available";
    return `${value} units available`;
}

function formatInstant(value: string): string {
    return new Date(value).toLocaleString();
}
