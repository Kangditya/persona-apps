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
    Textarea,
} from "@persona-apps/ui";
import Link from "next/link";
import { useParams } from "next/navigation";
import { type FormEvent, useEffect, useRef, useState } from "react";

import {
    eventsApi,
    type OfferingTransition,
    type OperationsOffering,
    type PatchOfferingInput,
} from "../api/events";
import type { OperationsSession } from "../api/session";
import { AccessBoundary } from "../features/events/AccessBoundary";
import {
    isStaleVersion,
    needsSessionRefresh,
    safeErrorMessage,
} from "../features/events/errors";
import {
    exactInteger,
    FormValueError,
    offeringTransitions,
    optionalExactInteger,
} from "../features/events/forms";
import { operationsQueryKeys, useOffering } from "../features/events/queries";
import { eventPath } from "../routes/paths";

type LifecycleIntent = {
    action: OfferingTransition;
    expectedVersion: number;
    key: string;
};

export function OfferingDetailPage() {
    const offeringId = useParams<{ offeringId: string }>()?.offeringId ?? "";
    return (
        <AccessBoundary permission="offering.read">
            {(session) => (
                <OfferingContent offeringId={offeringId} session={session} />
            )}
        </AccessBoundary>
    );
}

function OfferingContent({
    offeringId,
    session,
}: {
    offeringId: string;
    session: OperationsSession;
}) {
    const queryClient = useQueryClient();
    const offering = useOffering(offeringId, true);
    useEffect(() => {
        if (needsSessionRefresh(offering.error)) {
            void queryClient.invalidateQueries({
                queryKey: operationsQueryKeys.session,
            });
        }
    }, [offering.error, queryClient]);
    if (offering.isPending) {
        return (
            <div className="grid gap-4" aria-label="Loading Offering">
                <Skeleton className="h-10 w-72" />
                <Skeleton className="h-48 w-full" />
            </div>
        );
    }
    if (offering.isError) {
        return (
            <section>
                <h1 className="text-3xl font-semibold" tabIndex={-1} autoFocus>
                    Offering unavailable
                </h1>
                <Alert variant="destructive" className="mt-5">
                    {safeErrorMessage(offering.error)}
                    <Button
                        variant="outline"
                        className="ml-3"
                        onClick={() => offering.refetch()}
                    >
                        Try again
                    </Button>
                </Alert>
            </section>
        );
    }
    return (
        <section className="grid gap-8">
            <header>
                <Link
                    className="text-sm text-cyan-300 underline"
                    href={eventPath(offering.data.eventId)}
                >
                    Return to Event
                </Link>
                <div className="mt-3 flex flex-wrap items-center gap-3">
                    <h1
                        className="text-4xl font-semibold"
                        tabIndex={-1}
                        autoFocus
                    >
                        {offering.data.name}
                    </h1>
                    <Badge variant="secondary">{offering.data.status}</Badge>
                </div>
                <p className="mt-3 text-slate-300">
                    {offering.data.code}; version {offering.data.version}.
                </p>
            </header>
            <OfferingEditor
                value={offering.data}
                session={session}
                reload={() => offering.refetch()}
            />
        </section>
    );
}

function OfferingEditor({
    value,
    session,
    reload,
}: {
    value: OperationsOffering;
    session: OperationsSession;
    reload: () => unknown;
}) {
    const queryClient = useQueryClient();
    const canManage = session.permissions.includes("offering.manage");
    const canEdit = value.status === "DRAFT" || value.status === "UNAVAILABLE";
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [notice, setNotice] = useState("");
    const [lifecycleIntent, setLifecycleIntent] =
        useState<LifecycleIntent | null>(null);
    const errorSummary = useRef<HTMLDivElement>(null);
    const patch = useMutation({
        mutationFn: (input: PatchOfferingInput) =>
            eventsApi.patchOffering(value.id, input, session.csrfToken),
        onSuccess: async (updated) => {
            setNotice(
                `Offering configuration saved at version ${updated.version}.`,
            );
            await refreshOfferingQueries(queryClient, value);
        },
    });
    const transition = useMutation({
        mutationFn: (intent: LifecycleIntent) =>
            eventsApi.transitionOffering(
                value.id,
                intent.action,
                intent.expectedVersion,
                session.csrfToken,
                intent.key,
            ),
        onSuccess: async (updated) => {
            setLifecycleIntent(null);
            setNotice(
                `Offering is now ${updated.status} at version ${updated.version}.`,
            );
            await refreshOfferingQueries(queryClient, value);
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
            const input: PatchOfferingInput = {
                expected_version: value.version,
            };
            const name = String(data.get("name") ?? "").trim();
            if (name.length < 1 || name.length > 255) {
                throw new FormValueError(
                    "name",
                    "Name must contain 1 to 255 characters.",
                );
            }
            if (name !== value.name) input.name = name;
            const description = String(data.get("description") ?? "").trim();
            if (description.length > 10_000) {
                throw new FormValueError(
                    "description",
                    "Description must not exceed 10,000 characters.",
                );
            }
            if (description !== (value.description ?? "")) {
                input.description = description || null;
            }
            const priceRaw = String(data.get("price_minor") ?? "").trim();
            if (priceRaw !== String(value.priceMinor)) {
                input.price_minor = exactInteger(
                    "price_minor",
                    "Price minor units",
                    priceRaw,
                    0,
                    Number.MAX_SAFE_INTEGER,
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

    function runTransition(action: OfferingTransition) {
        if (
            action === "archive" &&
            !window.confirm("Confirm Offering archive?")
        )
            return;
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
                <CardTitle>Offering configuration and lifecycle</CardTitle>
                <CardDescription>
                    Availability is advisory and prices remain exact minor-unit
                    integers.
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
                <dl className="grid gap-4 text-sm sm:grid-cols-2">
                    <Description label="Code" value={value.code} />
                    <Description
                        label="Offering kind"
                        value={value.offeringKind}
                    />
                    <Description
                        label="Price"
                        value={`${value.priceMinor} ${value.currencyCode} minor units`}
                    />
                    <Description
                        label="Participant capacity"
                        value={String(value.participantCapacity)}
                    />
                    <Description
                        label="Participant quota"
                        value={
                            value.participantQuota === null
                                ? "Unbounded"
                                : String(value.participantQuota)
                        }
                    />
                    <Description
                        label="Advisory availability"
                        value={availabilityLabel(
                            value.availableParticipantUnits,
                        )}
                    />
                    <Description
                        label="Published"
                        value={
                            value.publishedAt
                                ? new Date(value.publishedAt).toLocaleString()
                                : "Not yet"
                        }
                    />
                    <Description
                        label="Updated"
                        value={new Date(value.updatedAt).toLocaleString()}
                    />
                </dl>
                {canManage && canEdit ? (
                    <form className="grid gap-5" onSubmit={submit} noValidate>
                        <Field
                            id="edit-offering-name"
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
                        <Field
                            id="edit-offering-description"
                            label="Description"
                            error={errors.description}
                        >
                            <Textarea
                                name="description"
                                defaultValue={value.description ?? ""}
                                maxLength={10_000}
                            />
                        </Field>
                        <div className="grid gap-5 md:grid-cols-2">
                            <Field
                                id="edit-offering-price"
                                label="Price minor units"
                                required
                                error={errors.price_minor}
                            >
                                <Input
                                    name="price_minor"
                                    inputMode="numeric"
                                    defaultValue={value.priceMinor}
                                />
                            </Field>
                            <Field
                                id="edit-offering-quota"
                                label="Participant quota"
                                error={errors.participant_quota}
                                description="Leave empty to clear the Offering-specific quota."
                            >
                                <Input
                                    name="participant_quota"
                                    inputMode="numeric"
                                    defaultValue={value.participantQuota ?? ""}
                                />
                            </Field>
                        </div>
                        <Button type="submit" loading={patch.isPending}>
                            Save configuration
                        </Button>
                    </form>
                ) : (
                    <Alert>
                        {!canManage
                            ? "Editing requires the offering.manage permission."
                            : "Configuration is immutable in the current Offering state."}
                    </Alert>
                )}
                {canManage ? (
                    <div>
                        <h2 className="font-semibold">Lifecycle actions</h2>
                        <div className="mt-3 flex flex-wrap gap-3">
                            {offeringTransitions(value.status).map((action) => (
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
                                    {offeringActionLabel(action)}
                                </Button>
                            ))}
                            {transition.isError && lifecycleIntent ? (
                                <Button
                                    type="button"
                                    variant="outline"
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

function Description({ label, value }: { label: string; value: string }) {
    return (
        <div>
            <dt className="text-slate-400">{label}</dt>
            <dd>{value}</dd>
        </div>
    );
}

async function refreshOfferingQueries(
    queryClient: ReturnType<typeof useQueryClient>,
    value: OperationsOffering,
) {
    await Promise.all([
        queryClient.invalidateQueries({
            queryKey: operationsQueryKeys.offeringDetail(value.id),
        }),
        queryClient.invalidateQueries({
            queryKey: operationsQueryKeys.offeringLists(value.eventId),
        }),
    ]);
}

function availabilityLabel(value: number | null): string {
    if (value === null) return "Unbounded by Event and Offering quotas";
    if (value === 0) return "Zero participant units available";
    return `${value} participant units available`;
}

function offeringActionLabel(action: OfferingTransition): string {
    return {
        publish: "Publish",
        unavailable: "Mark unavailable",
        archive: "Archive",
    }[action];
}
