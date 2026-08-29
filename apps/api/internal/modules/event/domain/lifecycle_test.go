package domain

import (
    "errors"
    "strings"
    "testing"
    "time"
)

func TestCreateNormalizesAndValidatesConfiguration(t *testing.T) {
    quota := int64(4)
    opens := time.Date(2026, time.January, 1, 8, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
    closes := opens.Add(24 * time.Hour)

    mutation, err := Create(CreateInput{
        EventYear:            2026,
        Name:                 "  Qurban 2026  ",
        RegistrationOpensAt:  &opens,
        RegistrationClosesAt: &closes,
        ParticipantQuota:     &quota,
    })
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    if mutation.Event.Name != "Qurban 2026" || mutation.Event.Status != StatusDraft || mutation.Event.Version != 1 {
        t.Fatalf("Create() event = %#v", mutation.Event)
    }
    if mutation.Event.RegistrationOpensAt.Location() != time.UTC || mutation.Event.RegistrationClosesAt.Location() != time.UTC {
        t.Fatal("Create() did not normalize registration instants to UTC")
    }
    if mutation.Action != ActionCreate || !mutation.RetrySensitive || mutation.Before != nil || mutation.OutboxEventType != "" {
        t.Fatalf("Create() effects = %#v", mutation)
    }

    tests := []struct {
        name  string
        input CreateInput
    }{
        {name: "year before supported range", input: CreateInput{EventYear: MinEventYear - 1, Name: "Qurban"}},
        {name: "year after supported range", input: CreateInput{EventYear: MaxEventYear + 1, Name: "Qurban"}},
        {name: "blank name", input: CreateInput{EventYear: 2026, Name: " \t "}},
        {name: "name too long", input: CreateInput{EventYear: 2026, Name: strings.Repeat("a", MaxNameLength+1)}},
        {name: "same window boundary", input: CreateInput{EventYear: 2026, Name: "Qurban", RegistrationOpensAt: &opens, RegistrationClosesAt: &opens}},
        {name: "negative quota", input: CreateInput{EventYear: 2026, Name: "Qurban", ParticipantQuota: int64Pointer(-1)}},
        {name: "unsafe quota", input: CreateInput{EventYear: 2026, Name: "Qurban", ParticipantQuota: int64Pointer(MaxSafeInteger + 1)}},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if _, err := Create(test.input); !errors.Is(err, ErrInvalidInput) {
                t.Fatalf("Create() error = %v, want invalid input", err)
            }
        })
    }
}

func TestUpdateHonorsConfigurationFreezeAndNullableClearing(t *testing.T) {
    opened := time.Date(2026, time.January, 1, 8, 0, 0, 0, time.UTC)
    closed := opened.Add(time.Hour)
    quota := int64(3)
    name := "Updated Qurban"
    input := UpdateInput{
        ExpectedVersion:      1,
        Name:                 &name,
        RegistrationOpensAt:  OptionalTime{Set: true},
        RegistrationClosesAt: OptionalTime{Set: true},
        ParticipantQuota:     OptionalInt64{Set: true},
    }

    tests := []struct {
        status  Status
        wantErr error
    }{
        {status: StatusDraft},
        {status: StatusPublished},
        {status: StatusSuspended},
        {status: StatusActive, wantErr: ErrConfigurationFrozen},
        {status: StatusClosed, wantErr: ErrImmutable},
        {status: StatusArchived, wantErr: ErrImmutable},
    }
    for _, test := range tests {
        t.Run(string(test.status), func(t *testing.T) {
            current := testEvent(test.status)
            current.RegistrationOpensAt = &opened
            current.RegistrationClosesAt = &closed
            current.ParticipantQuota = &quota

            mutation, err := Update(current, input)
            if test.wantErr != nil {
                if !errors.Is(err, test.wantErr) {
                    t.Fatalf("Update() error = %v, want %v", err, test.wantErr)
                }
                return
            }
            if err != nil {
                t.Fatalf("Update() error = %v", err)
            }
            if mutation.Event.Name != name || mutation.Event.RegistrationOpensAt != nil || mutation.Event.RegistrationClosesAt != nil || mutation.Event.ParticipantQuota != nil {
                t.Fatalf("Update() event = %#v", mutation.Event)
            }
            if mutation.Event.Version != 2 || mutation.Action != ActionUpdate || mutation.RetrySensitive || mutation.OutboxEventType != "" || mutation.Before == nil {
                t.Fatalf("Update() effects = %#v", mutation)
            }
        })
    }
}

func TestUpdateRejectsStaleVersionAndEmptyPatch(t *testing.T) {
    current := testEvent(StatusDraft)
    if _, err := Update(current, UpdateInput{ExpectedVersion: 2, Name: stringPointer("Changed")}); !errors.Is(err, ErrStaleVersion) {
        t.Fatalf("Update() error = %v, want stale version", err)
    }
    if _, err := Update(current, UpdateInput{ExpectedVersion: 1}); !errors.Is(err, ErrNoChanges) {
        t.Fatalf("Update() error = %v, want no changes", err)
    }
}

func TestUpdateReturnsCurrentEventForNormalizedNoOp(t *testing.T) {
    current := testEvent(StatusDraft)
    current.ID = "11111111-1111-1111-1111-111111111111"
    current.Name = "Qurban 2026"
    mutation, err := Update(current, UpdateInput{ExpectedVersion: 1, Name: stringPointer("  Qurban 2026  ")})
    if err != nil {
        t.Fatalf("Update() error = %v", err)
    }
    if mutation.Event.Version != 1 || mutation.Action != "" || mutation.Before != nil || mutation.After.Version != 1 {
        t.Fatalf("Update() no-op mutation = %#v", mutation)
    }
}

func TestLifecycleCommandsExposeExactEffects(t *testing.T) {
    statuses := []Status{StatusDraft, StatusPublished, StatusActive, StatusSuspended, StatusClosed, StatusArchived}
    commands := []struct {
        name    string
        command func(Event, TransitionInput) (Mutation, error)
        allowed map[Status]Status
        action  string
        outbox  string
    }{
        {name: "publish", command: Publish, allowed: map[Status]Status{StatusDraft: StatusPublished}, action: ActionPublish, outbox: OutboxEventPublished},
        {name: "activate", command: Activate, allowed: map[Status]Status{StatusPublished: StatusActive, StatusSuspended: StatusActive}, action: ActionActivate, outbox: OutboxEventActivated},
        {name: "suspend", command: Suspend, allowed: map[Status]Status{StatusActive: StatusSuspended}, action: ActionSuspend, outbox: OutboxEventSuspended},
        {name: "close", command: Close, allowed: map[Status]Status{StatusActive: StatusClosed, StatusSuspended: StatusClosed}, action: ActionClose, outbox: OutboxEventClosed},
        {name: "archive", command: Archive, allowed: map[Status]Status{StatusClosed: StatusArchived}, action: ActionArchive, outbox: OutboxEventArchived},
    }
    for _, command := range commands {
        for _, status := range statuses {
            t.Run(command.name+"_from_"+string(status), func(t *testing.T) {
                mutation, err := command.command(testEvent(status), TransitionInput{ExpectedVersion: 1})
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
                if mutation.Event.Status != target || mutation.Event.Version != 2 || mutation.Action != command.action || mutation.OutboxEventType != command.outbox || !mutation.RetrySensitive || mutation.Before == nil {
                    t.Fatalf("command result = %#v", mutation)
                }
            })
        }
    }

    if _, err := Publish(testEvent(StatusDraft), TransitionInput{ExpectedVersion: 2}); !errors.Is(err, ErrStaleVersion) {
        t.Fatalf("stale lifecycle error = %v", err)
    }
    if _, err := Publish(testEvent(StatusDraft), TransitionInput{}); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("zero expected version error = %v", err)
    }
    maximum := testEvent(StatusDraft)
    maximum.Version = MaxSafeInteger
    if _, err := Publish(maximum, TransitionInput{ExpectedVersion: MaxSafeInteger}); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("maximum version error = %v", err)
    }
}

func testEvent(status Status) Event {
    return Event{EventYear: 2026, Name: "Qurban 2026", Status: status, Version: 1}
}

func int64Pointer(value int64) *int64 {
    return &value
}

func stringPointer(value string) *string {
    return &value
}
