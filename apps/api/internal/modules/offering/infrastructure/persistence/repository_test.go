package persistence

import (
    "database/sql"
    "testing"
    "time"

    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
)

func TestOfferingRowMapsNullableValuesAndUTC(t *testing.T) {
    location := time.FixedZone("test", 7*60*60)
    published := time.Date(2026, time.August, 21, 10, 0, 0, 0, location)
    row := offeringRow{
        ID:                  "offering-id",
        EventID:             "event-id",
        Code:                "SHARE",
        Name:                "One share",
        Kind:                "SHARE",
        Description:         sql.NullString{String: "Description", Valid: true},
        PriceMinor:          100,
        CurrencyCode:        "IDR",
        ParticipantCapacity: 1,
        ParticipantQuota:    sql.NullInt64{Int64: 100, Valid: true},
        Status:              "PUBLISHED",
        PublishedAt:         sql.NullTime{Time: published, Valid: true},
        Version:             1,
        CreatedAt:           sql.NullTime{Time: published, Valid: true},
        UpdatedAt:           sql.NullTime{Time: published.Add(time.Minute), Valid: true},
    }

    value := row.offering()
    if value.ID != row.ID || value.Status != offeringdomain.StatusPublished || value.Description == nil || *value.Description != "Description" || value.ParticipantQuota == nil || *value.ParticipantQuota != 100 {
        t.Fatalf("offering = %#v", value)
    }
    if value.PublishedAt == nil || !value.PublishedAt.Equal(published.UTC()) || value.PublishedAt.Location() != time.UTC {
        t.Fatalf("published at = %#v", value.PublishedAt)
    }
    if value.CreatedAt.Location() != time.UTC || value.UpdatedAt.Location() != time.UTC {
        t.Fatalf("timestamps = %v, %v", value.CreatedAt, value.UpdatedAt)
    }

    row.Description = sql.NullString{}
    row.ParticipantQuota = sql.NullInt64{}
    row.PublishedAt = sql.NullTime{}
    row.CreatedAt = sql.NullTime{}
    row.UpdatedAt = sql.NullTime{}
    value = row.offering()
    if value.Description != nil || value.ParticipantQuota != nil || value.PublishedAt != nil || !value.CreatedAt.IsZero() || !value.UpdatedAt.IsZero() {
        t.Fatalf("empty nullable fields = %#v", value)
    }
}
