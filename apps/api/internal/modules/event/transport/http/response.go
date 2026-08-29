package http

import (
    "encoding/json"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
)

type operationsEvent struct {
    ID                   string             `json:"id"`
    EventYear            int                `json:"event_year"`
    Name                 string             `json:"name"`
    Status               eventdomain.Status `json:"status"`
    RegistrationOpensAt  *time.Time         `json:"registration_opens_at,omitempty"`
    RegistrationClosesAt *time.Time         `json:"registration_closes_at,omitempty"`
    ParticipantQuota     *int64             `json:"participant_quota,omitempty"`
    Version              int64              `json:"version"`
    CreatedAt            time.Time          `json:"created_at"`
    UpdatedAt            time.Time          `json:"updated_at"`
}

type operationsEventResponse struct {
    Data operationsEvent `json:"data"`
}

type operationsEventListResponse struct {
    Data []operationsEvent `json:"data"`
    Page operationsPage    `json:"page"`
}

type operationsPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
}

func eventReplayResponse(status int, value eventdomain.Event) (idempotency.Response, error) {
    body, err := json.Marshal(operationsEventResponse{Data: toOperationsEvent(value)})
    if err != nil {
        return idempotency.Response{}, err
    }
    return idempotency.Response{Status: status, Body: body}, nil
}

func toOperationsEvent(value eventdomain.Event) operationsEvent {
    return operationsEvent{
        ID: value.ID, EventYear: value.EventYear, Name: value.Name, Status: value.Status,
        RegistrationOpensAt: copyTime(value.RegistrationOpensAt), RegistrationClosesAt: copyTime(value.RegistrationClosesAt),
        ParticipantQuota: copyInt64(value.ParticipantQuota), Version: value.Version,
        CreatedAt: value.CreatedAt.UTC(), UpdatedAt: value.UpdatedAt.UTC(),
    }
}

func copyTime(value *time.Time) *time.Time {
    if value == nil {
        return nil
    }
    copied := value.UTC()
    return &copied
}

func copyInt64(value *int64) *int64 {
    if value == nil {
        return nil
    }
    copied := *value
    return &copied
}
