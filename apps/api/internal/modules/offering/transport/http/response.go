package http

import (
    "encoding/json"
    "time"

    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
)

type operationsOffering struct {
    ID                        string                `json:"id"`
    EventID                   string                `json:"event_id"`
    Code                      string                `json:"code"`
    Name                      string                `json:"name"`
    Kind                      string                `json:"offering_kind"`
    Description               *string               `json:"description,omitempty"`
    PriceMinor                int64                 `json:"price_minor"`
    CurrencyCode              string                `json:"currency_code"`
    ParticipantCapacity       int32                 `json:"participant_capacity"`
    ParticipantQuota          *int64                `json:"participant_quota,omitempty"`
    AvailableParticipantUnits *int64                `json:"available_participant_units"`
    Status                    offeringdomain.Status `json:"status"`
    PublishedAt               *time.Time            `json:"published_at,omitempty"`
    Version                   int64                 `json:"version"`
    CreatedAt                 time.Time             `json:"created_at"`
    UpdatedAt                 time.Time             `json:"updated_at"`
}

type operationsOfferingResponse struct {
    Data operationsOffering `json:"data"`
}

type operationsOfferingListResponse struct {
    Data []operationsOffering `json:"data"`
    Page operationsPage       `json:"page"`
}

type operationsPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
}

func offeringReplayResponse(status int, value offeringdomain.CatalogueOffering) (idempotency.Response, error) {
    body, err := json.Marshal(operationsOfferingResponse{Data: toOperationsOffering(value)})
    if err != nil {
        return idempotency.Response{}, err
    }
    return idempotency.Response{Status: status, Body: body}, nil
}

func toOperationsOffering(value offeringdomain.CatalogueOffering) operationsOffering {
    offering := value.Offering
    return operationsOffering{
        ID: offering.ID, EventID: offering.EventID, Code: offering.Code, Name: offering.Name, Kind: offering.Kind,
        Description: copyString(offering.Description), PriceMinor: offering.PriceMinor, CurrencyCode: offering.CurrencyCode,
        ParticipantCapacity: offering.ParticipantCapacity, ParticipantQuota: copyInt64(offering.ParticipantQuota),
        AvailableParticipantUnits: copyInt64(value.AvailableParticipantUnits), Status: offering.Status,
        PublishedAt: copyTime(offering.PublishedAt), Version: offering.Version,
        CreatedAt: offering.CreatedAt.UTC(), UpdatedAt: offering.UpdatedAt.UTC(),
    }
}

func copyString(value *string) *string {
    if value == nil {
        return nil
    }
    copied := *value
    return &copied
}

func copyInt64(value *int64) *int64 {
    if value == nil {
        return nil
    }
    copied := *value
    return &copied
}

func copyTime(value *time.Time) *time.Time {
    if value == nil {
        return nil
    }
    copied := value.UTC()
    return &copied
}
