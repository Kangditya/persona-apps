"use client";

import { ApiError } from "@persona-apps/api-client";
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
import Link from "next/link";
import { useParams } from "next/navigation";
import { type FormEvent, useEffect, useRef, useState } from "react";

import { isCatalogueID, type PublicOffering } from "../api/catalogue";
import { purchasesApi, type CreatedPurchase } from "../api/purchases";
import { CatalogueError } from "../features/catalogue/CatalogueError";
import {
    availabilityText,
    exactMinorPrice,
} from "../features/catalogue/presentation";
import {
    catalogueQueryKeys,
    usePublicOffering,
} from "../features/catalogue/queries";
import {
    buildCheckoutRequest,
    checkoutErrorState,
    CheckoutValidationError,
    createCheckoutIntent,
    type CheckoutIntent,
    type CheckoutParticipantDraft,
    type CheckoutPartyDraft,
} from "../features/purchases/checkout";
import { paths } from "../routes/paths";

const initialParticipants: CheckoutParticipantDraft[] = [
    {
        rowId: "participant-1",
        kind: "purchaser",
        displayName: "",
    },
];

export function CheckoutPage() {
    const offeringId = useParams<{ offeringId: string }>()?.offeringId ?? "";
    const validID = isCatalogueID(offeringId);
    const offering = usePublicOffering(offeringId, validID);
    const [created, setCreated] = useState(false);
    const unavailable =
        !validID ||
        (offering.error instanceof ApiError &&
            offering.error.kind === "not_found");

    return (
        <section className="grid gap-8">
            {!created ? (
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
                        Complete qurban details
                    </h1>
                    <p className="mt-3 max-w-2xl text-stone-600">
                        Purchaser, payer, and intended qurban participants may
                        be different people.
                    </p>
                </div>
            ) : null}

            {unavailable ? (
                <Alert>The selected Offering is unavailable.</Alert>
            ) : offering.isPending ? (
                <div
                    aria-label="Loading checkout Offering"
                    className="grid gap-3"
                    role="status"
                >
                    <Skeleton className="h-28 w-full max-w-3xl" />
                    <Skeleton className="h-72 w-full max-w-3xl" />
                    <span className="sr-only">Loading checkout Offering…</span>
                </div>
            ) : offering.isError ? (
                <CatalogueError
                    error={offering.error}
                    retry={() => void offering.refetch()}
                />
            ) : (
                <CheckoutContent
                    offering={offering.data}
                    onCreated={() => setCreated(true)}
                />
            )}
        </section>
    );
}

function CheckoutContent({
    offering,
    onCreated,
}: {
    offering: PublicOffering;
    onCreated: () => void;
}) {
    const queryClient = useQueryClient();
    const [payerMode, setPayerMode] = useState<"same" | "different">("same");
    const [participants, setParticipants] =
        useState<CheckoutParticipantDraft[]>(initialParticipants);
    const nextParticipantNumber = useRef(2);
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [intent, setIntent] = useState<CheckoutIntent | null>(null);
    const errorSummary = useRef<HTMLDivElement>(null);
    const create = useMutation({
        mutationFn: (value: CheckoutIntent) =>
            purchasesApi.create(value.input, value.key),
        retry: false,
        onSuccess: async () => {
            onCreated();
            setIntent(null);
            setErrors({});
            await Promise.all([
                queryClient.invalidateQueries({
                    queryKey: catalogueQueryKeys.offering(offering.id),
                }),
                queryClient.invalidateQueries({
                    queryKey: catalogueQueryKeys.offerings(offering.eventId),
                }),
            ]);
        },
    });

    useEffect(() => {
        if (create.isError || Object.keys(errors).length > 0) {
            errorSummary.current?.focus();
        }
    }, [create.isError, errors]);

    if (create.isSuccess) {
        return <CheckoutConfirmation purchase={create.data} />;
    }

    const serverError = create.isError
        ? checkoutErrorState(create.error)
        : null;

    function clearFrozenIntent() {
        setIntent(null);
        setErrors({});
        if (create.isError) create.reset();
    }

    function changePayerMode(mode: "same" | "different") {
        clearFrozenIntent();
        setPayerMode(mode);
        if (mode === "same") {
            setParticipants((current) =>
                current.map((participant) =>
                    participant.kind === "payer"
                        ? { ...participant, kind: "purchaser" }
                        : participant,
                ),
            );
        }
    }

    function updateParticipant(
        rowId: string,
        update: Partial<CheckoutParticipantDraft>,
    ) {
        clearFrozenIntent();
        setParticipants((current) =>
            current.map((participant) =>
                participant.rowId === rowId
                    ? { ...participant, ...update }
                    : participant,
            ),
        );
    }

    function addParticipant() {
        if (participants.length >= offering.participantCapacity) return;
        clearFrozenIntent();
        const number = nextParticipantNumber.current++;
        setParticipants((current) => [
            ...current,
            {
                rowId: `participant-${String(number)}`,
                kind: "name",
                displayName: "",
            },
        ]);
    }

    function removeParticipant(rowId: string) {
        if (participants.length === 1) return;
        clearFrozenIntent();
        setParticipants((current) =>
            current.filter((participant) => participant.rowId !== rowId),
        );
    }

    function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const form = event.currentTarget;
        const data = new FormData(form);
        try {
            const input = buildCheckoutRequest(
                offering.id,
                offering.participantCapacity,
                {
                    purchaser: partyDraft(data, "purchaser"),
                    payerMode,
                    payer: partyDraft(data, "payer"),
                    participants,
                },
            );
            const nextIntent = createCheckoutIntent(input);
            setErrors({});
            setIntent(nextIntent);
            create.reset();
            create.mutate(nextIntent);
        } catch (error) {
            if (error instanceof CheckoutValidationError) {
                setErrors({ [error.field]: error.message });
            } else {
                setErrors({
                    form: "The checkout form could not be read.",
                });
            }
        }
    }

    return (
        <>
            <CheckoutProgress />

            {(Object.keys(errors).length > 0 || serverError) && (
                <div ref={errorSummary} tabIndex={-1}>
                    <Alert variant="destructive" role="alert">
                        <p>
                            {Object.values(errors)[0] ?? serverError?.message}
                        </p>
                        {serverError?.requestId ? (
                            <p className="mt-2 text-sm">
                                Request ID: <code>{serverError.requestId}</code>
                            </p>
                        ) : null}
                    </Alert>
                    {serverError?.retryOriginal && intent ? (
                        <Button
                            type="button"
                            variant="outline"
                            className="mt-3"
                            disabled={create.isPending}
                            onClick={() => create.mutate(intent)}
                        >
                            Retry the same submitted request
                        </Button>
                    ) : null}
                </div>
            )}

            <form
                className="grid max-w-3xl gap-6"
                onChange={clearFrozenIntent}
                onSubmit={submit}
            >
                <BuyerCard
                    errors={errors}
                    payerMode={payerMode}
                    setPayerMode={changePayerMode}
                />
                <ParticipantsCard
                    capacity={offering.participantCapacity}
                    errors={errors}
                    participants={participants}
                    payerMode={payerMode}
                    addParticipant={addParticipant}
                    removeParticipant={removeParticipant}
                    updateParticipant={updateParticipant}
                />
                <PackageSummary
                    offering={offering}
                    participantCount={participants.length}
                />

                <Alert>
                    <p className="font-medium">
                        Your data can be reviewed before submission.
                    </p>
                    <p className="mt-1 text-sm">
                        Only information required to create and operate this
                        qurban Purchase is requested.
                    </p>
                </Alert>

                <Button type="submit" size="lg" loading={create.isPending}>
                    Create pending Purchase
                </Button>
                <span
                    aria-live="polite"
                    className="text-sm text-stone-600"
                    role="status"
                >
                    {create.isPending
                        ? "Creating the Purchase and reserving quota…"
                        : ""}
                </span>
            </form>

            <HelpCard />
        </>
    );
}

function CheckoutProgress() {
    const steps = [
        ["1", "Data", "Current step"],
        ["2", "Payment", "Upcoming"],
        ["3", "Finish", "Upcoming"],
    ] as const;
    return (
        <ol
            aria-label="Checkout progress"
            className="grid max-w-3xl grid-cols-3 gap-3"
        >
            {steps.map(([number, label, state], index) => (
                <li
                    key={number}
                    aria-current={index === 0 ? "step" : undefined}
                    className="grid justify-items-center gap-2 text-center"
                >
                    <span
                        className={
                            index === 0
                                ? "grid size-9 place-items-center rounded-full bg-primary font-semibold text-primary-foreground"
                                : "grid size-9 place-items-center rounded-full bg-muted font-semibold text-muted-foreground"
                        }
                    >
                        {number}
                    </span>
                    <span className="font-medium">{label}</span>
                    <span className="text-xs text-stone-500">{state}</span>
                </li>
            ))}
        </ol>
    );
}

function BuyerCard({
    errors,
    payerMode,
    setPayerMode,
}: {
    errors: Record<string, string>;
    payerMode: "same" | "different";
    setPayerMode: (mode: "same" | "different") => void;
}) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Buyer and payer data</CardTitle>
                <CardDescription>
                    Purchaser and payer may be the same or different people.
                </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-5">
                <div className="grid gap-5 sm:grid-cols-2">
                    <Field
                        id="purchaser-display-name"
                        label="Purchaser name"
                        required
                        error={errors["purchaser.display_name"]}
                    >
                        <Input
                            name="purchaser_display_name"
                            autoComplete="name"
                            maxLength={255}
                        />
                    </Field>
                    <Field
                        id="purchaser-email"
                        label="Purchaser email"
                        error={errors["purchaser.email"]}
                    >
                        <Input
                            name="purchaser_email"
                            type="email"
                            autoComplete="email"
                            maxLength={320}
                        />
                    </Field>
                    <Field
                        id="purchaser-phone"
                        label="Purchaser phone"
                        error={errors["purchaser.phone"]}
                    >
                        <Input
                            name="purchaser_phone"
                            type="tel"
                            autoComplete="tel"
                            maxLength={64}
                        />
                    </Field>
                </div>

                <fieldset className="grid gap-3">
                    <legend className="font-medium">Who is paying?</legend>
                    <label className="flex min-h-11 items-center gap-3">
                        <input
                            type="radio"
                            name="payer_mode"
                            value="same"
                            checked={payerMode === "same"}
                            onChange={() => setPayerMode("same")}
                        />
                        Purchaser is also the payer
                    </label>
                    <label className="flex min-h-11 items-center gap-3">
                        <input
                            type="radio"
                            name="payer_mode"
                            value="different"
                            checked={payerMode === "different"}
                            onChange={() => setPayerMode("different")}
                        />
                        A different person is paying
                    </label>
                </fieldset>

                {payerMode === "different" ? (
                    <div className="grid gap-5 rounded-lg border p-4 sm:grid-cols-2">
                        <Field
                            id="payer-display-name"
                            label="Payer name"
                            required
                            error={errors["payer.display_name"]}
                        >
                            <Input
                                name="payer_display_name"
                                autoComplete="name"
                                maxLength={255}
                            />
                        </Field>
                        <Field
                            id="payer-email"
                            label="Payer email"
                            error={errors["payer.email"]}
                        >
                            <Input
                                name="payer_email"
                                type="email"
                                autoComplete="email"
                                maxLength={320}
                            />
                        </Field>
                        <Field
                            id="payer-phone"
                            label="Payer phone"
                            error={errors["payer.phone"]}
                        >
                            <Input
                                name="payer_phone"
                                type="tel"
                                autoComplete="tel"
                                maxLength={64}
                            />
                        </Field>
                    </div>
                ) : null}
            </CardContent>
        </Card>
    );
}

function ParticipantsCard({
    capacity,
    errors,
    participants,
    payerMode,
    addParticipant,
    removeParticipant,
    updateParticipant,
}: {
    capacity: number;
    errors: Record<string, string>;
    participants: CheckoutParticipantDraft[];
    payerMode: "same" | "different";
    addParticipant: () => void;
    removeParticipant: (rowId: string) => void;
    updateParticipant: (
        rowId: string,
        update: Partial<CheckoutParticipantDraft>,
    ) => void;
}) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Intended qurban participants</CardTitle>
                <CardDescription>
                    These people become Sohibul Qurban only after later
                    eligibility and activation.
                </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-5">
                {errors.participants ? (
                    <Alert variant="destructive">{errors.participants}</Alert>
                ) : null}
                {participants.map((participant, index) => {
                    const fieldPrefix = `participants.${String(index)}`;
                    return (
                        <fieldset
                            key={participant.rowId}
                            className="grid gap-4 rounded-lg border p-4"
                        >
                            <legend className="px-1 font-medium">
                                Participant {index + 1}
                            </legend>
                            <Field
                                id={`${participant.rowId}-kind`}
                                label="Participant identity"
                                error={errors[`${fieldPrefix}.kind`]}
                            >
                                <select
                                    className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                                    value={participant.kind}
                                    onChange={(event) =>
                                        updateParticipant(participant.rowId, {
                                            kind: event.target
                                                .value as CheckoutParticipantDraft["kind"],
                                            displayName: "",
                                        })
                                    }
                                >
                                    <option value="purchaser">Purchaser</option>
                                    {payerMode === "different" ? (
                                        <option value="payer">Payer</option>
                                    ) : null}
                                    <option value="name">
                                        Name-only participant
                                    </option>
                                </select>
                            </Field>
                            {participant.kind === "name" ? (
                                <Field
                                    id={`${participant.rowId}-display-name`}
                                    label="Participant name"
                                    required
                                    error={
                                        errors[`${fieldPrefix}.display_name`]
                                    }
                                >
                                    <Input
                                        maxLength={255}
                                        value={participant.displayName}
                                        onChange={(event) =>
                                            updateParticipant(
                                                participant.rowId,
                                                {
                                                    displayName:
                                                        event.target.value,
                                                },
                                            )
                                        }
                                    />
                                </Field>
                            ) : (
                                <p className="text-sm text-stone-600">
                                    The stored participant name is captured from
                                    the selected declared Party.
                                </p>
                            )}
                            {participants.length > 1 ? (
                                <Button
                                    type="button"
                                    variant="outline"
                                    className="w-fit"
                                    onClick={() =>
                                        removeParticipant(participant.rowId)
                                    }
                                >
                                    Remove participant
                                </Button>
                            ) : null}
                        </fieldset>
                    );
                })}
                <Button
                    type="button"
                    variant="outline"
                    className="w-fit"
                    disabled={participants.length >= capacity}
                    onClick={addParticipant}
                >
                    Add participant
                </Button>
                <p className="text-sm text-stone-600">
                    {participants.length} of {capacity} participant positions
                    used for this Purchase.
                </p>
            </CardContent>
        </Card>
    );
}

function PackageSummary({
    offering,
    participantCount,
}: {
    offering: PublicOffering;
    participantCount: number;
}) {
    return (
        <Card>
            <CardHeader>
                <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                        <CardTitle>Package summary</CardTitle>
                        <CardDescription>
                            {offering.name} · {offering.code}
                        </CardDescription>
                    </div>
                    <Badge variant="secondary">{offering.offeringKind}</Badge>
                </div>
            </CardHeader>
            <CardContent className="grid gap-3 sm:grid-cols-2">
                <div>
                    <p className="text-sm text-stone-500">Exact unit price</p>
                    <p className="font-medium">
                        {exactMinorPrice(
                            offering.currencyCode,
                            offering.priceMinor,
                        )}
                    </p>
                </div>
                <div>
                    <p className="text-sm text-stone-500">
                        Intended participants
                    </p>
                    <p className="font-medium">{participantCount}</p>
                </div>
                <p className="text-sm text-stone-600 sm:col-span-2">
                    {availabilityText(offering.availableParticipantUnits)} Final
                    total and quota are confirmed by the API when the Purchase
                    is created.
                </p>
            </CardContent>
        </Card>
    );
}

function CheckoutConfirmation({ purchase }: { purchase: CreatedPurchase }) {
    return (
        <section className="grid max-w-3xl gap-6">
            <div className="text-center">
                <div
                    aria-hidden="true"
                    className="mx-auto grid size-12 place-items-center rounded-full bg-secondary text-2xl text-secondary-foreground"
                >
                    ✓
                </div>
                <h1
                    className="mt-4 text-3xl font-semibold"
                    tabIndex={-1}
                    autoFocus
                >
                    Purchase created
                </h1>
                <p className="mt-2 text-stone-600">
                    The pending Purchase and quota reservation were created
                    successfully.
                </p>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div>
                            <Badge variant="secondary">Pending payment</Badge>
                            <CardTitle className="mt-3">
                                {purchase.purchaseRef}
                            </CardTitle>
                            <CardDescription>
                                Created {formatInstant(purchase.createdAt)}
                            </CardDescription>
                        </div>
                        <div className="text-right">
                            <p className="text-sm text-stone-500">
                                Exact total
                            </p>
                            <p className="text-xl font-semibold text-primary">
                                {exactMinorPrice(
                                    purchase.currencyCode,
                                    purchase.totalAmountMinor,
                                )}
                            </p>
                        </div>
                    </div>
                </CardHeader>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Purchase summary</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4">
                    <dl className="grid gap-4 sm:grid-cols-2">
                        <div>
                            <dt className="text-sm text-stone-500">Offering</dt>
                            <dd className="font-medium">
                                {purchase.offering.name}
                            </dd>
                        </div>
                        <div>
                            <dt className="text-sm text-stone-500">
                                Intended participants
                            </dt>
                            <dd className="font-medium">
                                {purchase.participantCount}
                            </dd>
                        </div>
                        <div className="sm:col-span-2">
                            <dt className="text-sm text-stone-500">
                                Reservation expires
                            </dt>
                            <dd className="font-medium">
                                {formatInstant(purchase.reservationExpiresAt)}
                            </dd>
                        </div>
                    </dl>
                    <ol
                        aria-label="Intended participant names"
                        className="grid gap-2"
                    >
                        {purchase.participants.map((participant) => (
                            <li
                                key={participant.sequenceNo}
                                className="rounded-md bg-muted px-3 py-2 text-sm"
                            >
                                {participant.sequenceNo}.{" "}
                                {participant.displayName}
                            </li>
                        ))}
                    </ol>
                </CardContent>
            </Card>

            <Alert>
                Payment instructions and Purchase tracking are not available in
                this Week 3 slice. Keep the non-secret Purchase reference for
                support.
            </Alert>

            <Link
                className="w-fit text-amber-800 underline underline-offset-4"
                href={paths.offerings}
            >
                Return to Offerings
            </Link>

            <HelpCard />
        </section>
    );
}

function HelpCard() {
    return (
        <Card className="max-w-3xl bg-secondary/40">
            <CardHeader>
                <CardTitle>Need help?</CardTitle>
                <CardDescription>
                    If checkout fails, keep the displayed request ID and share
                    it through your programme&apos;s existing support channel.
                </CardDescription>
            </CardHeader>
        </Card>
    );
}

function partyDraft(
    data: FormData,
    prefix: "purchaser" | "payer",
): CheckoutPartyDraft {
    return {
        displayName: formString(data, `${prefix}_display_name`),
        email: formString(data, `${prefix}_email`),
        phone: formString(data, `${prefix}_phone`),
    };
}

function formString(data: FormData, key: string): string {
    const value = data.get(key);
    return typeof value === "string" ? value : "";
}

function formatInstant(value: string): string {
    return new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
    }).format(new Date(value));
}
