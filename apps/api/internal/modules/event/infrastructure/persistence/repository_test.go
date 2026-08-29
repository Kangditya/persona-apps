package persistence

import (
    "database/sql"
    "testing"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
)

func TestEventRowMapsNullableValuesAndUTC(t *testing.T) {
    location := time.FixedZone("test", 7*60*60)
    opened := time.Date(2026, time.August, 21, 10, 0, 0, 0, location)
    closed := opened.Add(time.Hour)
    row := eventRow{
        ID:                   "event-id",
        EventYear:            2026,
        Name:                 "Qurban 2026",
        Status:               "ACTIVE",
        RegistrationOpensAt:  sql.NullTime{Time: opened, Valid: true},
        RegistrationClosesAt: sql.NullTime{Time: closed, Valid: true},
        ParticipantQuota:     sql.NullInt64{Int64: 100, Valid: true},
        Version:              1,
        CreatedAt:            opened,
        UpdatedAt:            closed,
    }

    value := row.event()
    if value.ID != row.ID || value.Status != eventdomain.StatusActive || value.ParticipantQuota == nil || *value.ParticipantQuota != 100 {
        t.Fatalf("event = %#v", value)
    }
    if value.RegistrationOpensAt == nil || !value.RegistrationOpensAt.Equal(opened.UTC()) || value.RegistrationOpensAt.Location() != time.UTC {
        t.Fatalf("registration opening = %#v", value.RegistrationOpensAt)
    }
    if value.RegistrationClosesAt == nil || !value.RegistrationClosesAt.Equal(closed.UTC()) || value.RegistrationClosesAt.Location() != time.UTC {
        t.Fatalf("registration closing = %#v", value.RegistrationClosesAt)
    }
    if value.CreatedAt.Location() != time.UTC || value.UpdatedAt.Location() != time.UTC {
        t.Fatalf("timestamps = %v, %v", value.CreatedAt, value.UpdatedAt)
    }

    row.RegistrationOpensAt = sql.NullTime{}
    row.RegistrationClosesAt = sql.NullTime{}
    row.ParticipantQuota = sql.NullInt64{}
    value = row.event()
    if value.RegistrationOpensAt != nil || value.RegistrationClosesAt != nil || value.ParticipantQuota != nil {
        t.Fatalf("empty nullable fields = %#v", value)
    }
}
