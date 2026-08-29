package domain

import (
    "errors"
    "strings"
    "testing"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
)

func TestCreateValidatesOfferingAndParent(t *testing.T) {
    input := validCreateInput()
    for _, status := range []eventdomain.Status{eventdomain.StatusDraft, eventdomain.StatusPublished, eventdomain.StatusActive, eventdomain.StatusSuspended} {
        t.Run(string(status), func(t *testing.T) {
            mutation, err := Create(eventdomain.Event{Status: status}, input)
            if err != nil {
                t.Fatalf("Create() error = %v", err)
            }
            if mutation.Offering.Status != StatusDraft || mutation.Offering.Version != 1 || mutation.Action != ActionCreate || !mutation.RetrySensitive || mutation.Before != nil || mutation.OutboxEventType != "" {
                t.Fatalf("Create() mutation = %#v", mutation)
            }
        })
    }
    for _, status := range []eventdomain.Status{eventdomain.StatusClosed, eventdomain.StatusArchived, "UNKNOWN"} {
        t.Run("rejected_"+string(status), func(t *testing.T) {
            if _, err := Create(eventdomain.Event{Status: status}, input); !errors.Is(err, ErrInvalidParentState) {
                t.Fatalf("Create() error = %v, want invalid parent", err)
            }
        })
    }

    tests := []struct {
        name  string
        input CreateInput
    }{
        {name: "missing event", input: CreateInput{Code: "COW", Name: "Cow", Kind: "SHARE", CurrencyCode: "IDR", ParticipantCapacity: 1}},
        {name: "blank code", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.Code = " \t " })},
        {name: "long code", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.Code = strings.Repeat("a", MaxCodeLength+1) })},
        {name: "blank name", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.Name = " " })},
        {name: "blank kind", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.Kind = " " })},
        {name: "long description", input: replaceCreate(validCreateInput(), func(input *CreateInput) {
            input.Description = stringPointer(strings.Repeat("a", MaxDescriptionLength+1))
        })},
        {name: "negative price", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.PriceMinor = -1 })},
        {name: "unsafe price", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.PriceMinor = MaxSafeInteger + 1 })},
        {name: "lowercase currency", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.CurrencyCode = "idr" })},
        {name: "zero capacity", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.ParticipantCapacity = 0 })},
        {name: "large capacity", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.ParticipantCapacity = MaxParticipantCapacity + 1 })},
        {name: "negative quota", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.ParticipantQuota = int64Pointer(-1) })},
        {name: "unsafe quota", input: replaceCreate(validCreateInput(), func(input *CreateInput) { input.ParticipantQuota = int64Pointer(MaxSafeInteger + 1) })},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if _, err := Create(eventdomain.Event{Status: eventdomain.StatusDraft}, test.input); !errors.Is(err, ErrInvalidInput) {
                t.Fatalf("Create() error = %v, want invalid input", err)
            }
        })
    }
}

func TestUpdateHonorsCommercialMutationPolicy(t *testing.T) {
    description := "Updated description"
    price := int64(2_000_000)
    name := "Updated Cow Share"
    input := UpdateInput{
        ExpectedVersion:  1,
        Name:             &name,
        Description:      OptionalString{Set: true, Value: &description},
        PriceMinor:       &price,
        ParticipantQuota: OptionalInt64{Set: true},
    }

    tests := []struct {
        name    string
        parent  eventdomain.Status
        status  Status
        wantErr error
    }{
        {name: "draft", parent: eventdomain.StatusActive, status: StatusDraft},
        {name: "unavailable", parent: eventdomain.StatusSuspended, status: StatusUnavailable},
        {name: "published frozen", parent: eventdomain.StatusActive, status: StatusPublished, wantErr: ErrConfigurationFrozen},
        {name: "archived immutable", parent: eventdomain.StatusActive, status: StatusArchived, wantErr: ErrImmutable},
        {name: "closed event", parent: eventdomain.StatusClosed, status: StatusDraft, wantErr: ErrInvalidParentState},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            current := testOffering(test.status)
            current.Description = stringPointer("Old description")
            current.ParticipantQuota = int64Pointer(4)
            mutation, err := Update(eventdomain.Event{Status: test.parent}, current, input)
            if test.wantErr != nil {
                if !errors.Is(err, test.wantErr) {
                    t.Fatalf("Update() error = %v, want %v", err, test.wantErr)
                }
                return
            }
            if err != nil {
                t.Fatalf("Update() error = %v", err)
            }
            if mutation.Offering.Name != name || mutation.Offering.Description == nil || *mutation.Offering.Description != description || mutation.Offering.PriceMinor != price || mutation.Offering.ParticipantQuota != nil {
                t.Fatalf("Update() offering = %#v", mutation.Offering)
            }
            if mutation.Offering.Code != current.Code || mutation.Offering.CurrencyCode != current.CurrencyCode || mutation.Offering.ParticipantCapacity != current.ParticipantCapacity || mutation.Offering.Version != 2 || mutation.Action != ActionUpdate || mutation.RetrySensitive || mutation.Before == nil || mutation.OutboxEventType != "" {
                t.Fatalf("Update() mutation = %#v", mutation)
            }
        })
    }

    if _, err := Update(eventdomain.Event{Status: eventdomain.StatusActive}, testOffering(StatusDraft), UpdateInput{ExpectedVersion: 2, Name: stringPointer("Changed")}); !errors.Is(err, ErrStaleVersion) {
        t.Fatalf("Update() stale error = %v", err)
    }
}

func TestUpdateReturnsCurrentOfferingForNormalizedNoOp(t *testing.T) {
    current := testOffering(StatusDraft)
    mutation, err := Update(eventdomain.Event{Status: eventdomain.StatusActive}, current, UpdateInput{ExpectedVersion: 1, Name: stringPointer("  Cow Share  ")})
    if err != nil {
        t.Fatalf("Update() error = %v", err)
    }
    if mutation.Offering.Version != 1 || mutation.Action != "" || mutation.Before != nil || mutation.After.Version != 1 {
        t.Fatalf("Update() no-op mutation = %#v", mutation)
    }
}

func TestLifecyclePreservesFirstPublicationAndCleanup(t *testing.T) {
    publication := time.Date(2026, time.January, 1, 8, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
    mutation, err := Publish(eventdomain.Event{Status: eventdomain.StatusPublished}, testOffering(StatusDraft), PublishInput{ExpectedVersion: 1, OccurredAt: publication})
    if err != nil {
        t.Fatalf("Publish() error = %v", err)
    }
    if mutation.Offering.Status != StatusPublished || mutation.Offering.PublishedAt == nil || !mutation.Offering.PublishedAt.Equal(publication.UTC()) || mutation.Action != ActionPublish || mutation.OutboxEventType != OutboxOfferingPublished || !mutation.RetrySensitive {
        t.Fatalf("Publish() mutation = %#v", mutation)
    }

    firstPublication := publication.UTC()
    current := testOffering(StatusUnavailable)
    current.PublishedAt = &firstPublication
    republished, err := Publish(eventdomain.Event{Status: eventdomain.StatusSuspended}, current, PublishInput{ExpectedVersion: 1, OccurredAt: publication.Add(24 * time.Hour)})
    if err != nil {
        t.Fatalf("republish error = %v", err)
    }
    if !republished.Offering.PublishedAt.Equal(firstPublication) {
        t.Fatalf("republish time = %s, want %s", republished.Offering.PublishedAt, firstPublication)
    }

    if _, err := Publish(eventdomain.Event{Status: eventdomain.StatusDraft}, testOffering(StatusDraft), PublishInput{ExpectedVersion: 1, OccurredAt: publication}); !errors.Is(err, ErrInvalidParentState) {
        t.Fatalf("Publish() error = %v, want invalid parent", err)
    }
    if _, err := Publish(eventdomain.Event{Status: eventdomain.StatusPublished}, testOffering(StatusPublished), PublishInput{ExpectedVersion: 1, OccurredAt: publication}); !errors.Is(err, ErrInvalidTransition) {
        t.Fatalf("Publish() error = %v, want invalid transition", err)
    }

    unavailable, err := MarkUnavailable(testOffering(StatusPublished), TransitionInput{ExpectedVersion: 1})
    if err != nil || unavailable.Offering.Status != StatusUnavailable || unavailable.OutboxEventType != OutboxOfferingUnavailable {
        t.Fatalf("MarkUnavailable() = %#v, %v", unavailable, err)
    }
    for _, status := range []Status{StatusDraft, StatusPublished, StatusUnavailable} {
        t.Run("archive_"+string(status), func(t *testing.T) {
            archived, err := Archive(testOffering(status), TransitionInput{ExpectedVersion: 1})
            if err != nil || archived.Offering.Status != StatusArchived || archived.OutboxEventType != OutboxOfferingArchived {
                t.Fatalf("Archive() = %#v, %v", archived, err)
            }
        })
    }
}

func TestLifecycleMatrixRejectsEveryOtherStatus(t *testing.T) {
    statuses := []Status{StatusDraft, StatusPublished, StatusUnavailable, StatusArchived}
    occurredAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
    commands := []struct {
        name    string
        command func(Offering) (Mutation, error)
        allowed map[Status]Status
        action  string
        outbox  string
    }{
        {
            name: "publish",
            command: func(current Offering) (Mutation, error) {
                return Publish(eventdomain.Event{Status: eventdomain.StatusActive}, current, PublishInput{ExpectedVersion: 1, OccurredAt: occurredAt})
            },
            allowed: map[Status]Status{StatusDraft: StatusPublished, StatusUnavailable: StatusPublished},
            action:  ActionPublish,
            outbox:  OutboxOfferingPublished,
        },
        {
            name: "unavailable",
            command: func(current Offering) (Mutation, error) {
                return MarkUnavailable(current, TransitionInput{ExpectedVersion: 1})
            },
            allowed: map[Status]Status{StatusPublished: StatusUnavailable},
            action:  ActionUnavailable,
            outbox:  OutboxOfferingUnavailable,
        },
        {
            name: "archive",
            command: func(current Offering) (Mutation, error) {
                return Archive(current, TransitionInput{ExpectedVersion: 1})
            },
            allowed: map[Status]Status{StatusDraft: StatusArchived, StatusPublished: StatusArchived, StatusUnavailable: StatusArchived},
            action:  ActionArchive,
            outbox:  OutboxOfferingArchived,
        },
    }
    for _, command := range commands {
        for _, status := range statuses {
            t.Run(command.name+"_from_"+string(status), func(t *testing.T) {
                mutation, err := command.command(testOffering(status))
                target, allowed := command.allowed[status]
                if !allowed {
                    if !errors.Is(err, ErrInvalidTransition) {
                        t.Fatalf("command error = %v, want invalid transition", err)
                    }
                    return
                }
                if err != nil {
                    t.Fatalf("command error = %v", err)
                }
                if mutation.Offering.Status != target || mutation.Offering.Version != 2 || mutation.Action != command.action || mutation.OutboxEventType != command.outbox || !mutation.RetrySensitive || mutation.Before == nil {
                    t.Fatalf("command result = %#v", mutation)
                }
            })
        }
    }

    if _, err := MarkUnavailable(testOffering(StatusPublished), TransitionInput{ExpectedVersion: 2}); !errors.Is(err, ErrStaleVersion) {
        t.Fatalf("stale lifecycle error = %v", err)
    }
    maximum := testOffering(StatusPublished)
    maximum.Version = MaxSafeInteger
    if _, err := MarkUnavailable(maximum, TransitionInput{ExpectedVersion: MaxSafeInteger}); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("maximum version error = %v", err)
    }
}

func TestPublicVisibilityAndAvailability(t *testing.T) {
    if !IsPublic(eventdomain.Event{Status: eventdomain.StatusActive}, testOffering(StatusPublished)) {
        t.Fatal("active published offering should be public")
    }
    if IsPublic(eventdomain.Event{Status: eventdomain.StatusSuspended}, testOffering(StatusPublished)) || IsPublic(eventdomain.Event{Status: eventdomain.StatusActive}, testOffering(StatusUnavailable)) {
        t.Fatal("hidden offering was public")
    }

    tests := []struct {
        name  string
        input AvailabilityInput
        want  *int64
    }{
        {name: "fully unbounded", input: AvailabilityInput{}, want: nil},
        {name: "event quota", input: AvailabilityInput{EventQuota: int64Pointer(10), EventUsage: ReservationUsage{Reserved: 3, Consumed: 2}}, want: int64Pointer(5)},
        {name: "offering quota", input: AvailabilityInput{OfferingQuota: int64Pointer(8), OfferingUsage: ReservationUsage{Reserved: 1, Consumed: 3}}, want: int64Pointer(4)},
        {name: "minimum finite quota", input: AvailabilityInput{EventQuota: int64Pointer(10), OfferingQuota: int64Pointer(8), EventUsage: ReservationUsage{Reserved: 2}, OfferingUsage: ReservationUsage{Consumed: 3}}, want: int64Pointer(5)},
        {name: "released and expired ignored", input: AvailabilityInput{EventQuota: int64Pointer(5), EventUsage: ReservationUsage{Reserved: 1, Released: 100, Expired: 100}}, want: int64Pointer(4)},
        {name: "exhausted", input: AvailabilityInput{OfferingQuota: int64Pointer(3), OfferingUsage: ReservationUsage{Reserved: 1, Consumed: 2}}, want: int64Pointer(0)},
        {name: "overcommitted clamps", input: AvailabilityInput{OfferingQuota: int64Pointer(3), OfferingUsage: ReservationUsage{Reserved: 5}}, want: int64Pointer(0)},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            got, err := AvailableParticipantUnits(test.input)
            if err != nil {
                t.Fatalf("AvailableParticipantUnits() error = %v", err)
            }
            if !sameInt64(got, test.want) {
                t.Fatalf("AvailableParticipantUnits() = %v, want %v", got, test.want)
            }
        })
    }
    if _, err := AvailableParticipantUnits(AvailabilityInput{EventQuota: int64Pointer(1), EventUsage: ReservationUsage{Reserved: -1}}); !errors.Is(err, ErrInvalidAvailability) {
        t.Fatalf("AvailableParticipantUnits() error = %v, want invalid availability", err)
    }
}

func validCreateInput() CreateInput {
    return CreateInput{
        EventID:             "event-1",
        Code:                "COW-SHARE",
        Name:                "Cow Share",
        Kind:                "SHARE",
        Description:         stringPointer("Seven participant share"),
        PriceMinor:          1_500_000,
        CurrencyCode:        "IDR",
        ParticipantCapacity: 7,
        ParticipantQuota:    int64Pointer(21),
    }
}

func testOffering(status Status) Offering {
    return Offering{
        EventID:             "event-1",
        Code:                "COW-SHARE",
        Name:                "Cow Share",
        Kind:                "SHARE",
        PriceMinor:          1_500_000,
        CurrencyCode:        "IDR",
        ParticipantCapacity: 7,
        Status:              status,
        Version:             1,
    }
}

func replaceCreate(input CreateInput, change func(*CreateInput)) CreateInput {
    change(&input)
    return input
}

func sameInt64(got, want *int64) bool {
    return got == nil && want == nil || got != nil && want != nil && *got == *want
}

func stringPointer(value string) *string {
    return &value
}
