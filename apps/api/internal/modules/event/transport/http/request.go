package http

import (
    "encoding/json"
    "strconv"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type createEventRequest struct {
    EventYear            int        `json:"event_year"`
    Name                 string     `json:"name"`
    RegistrationOpensAt  *time.Time `json:"registration_opens_at"`
    RegistrationClosesAt *time.Time `json:"registration_closes_at"`
    ParticipantQuota     *int64     `json:"participant_quota"`
}

type patchEventRequest struct {
    ExpectedVersion      int64           `json:"expected_version"`
    Name                 json.RawMessage `json:"name"`
    RegistrationOpensAt  json.RawMessage `json:"registration_opens_at"`
    RegistrationClosesAt json.RawMessage `json:"registration_closes_at"`
    ParticipantQuota     json.RawMessage `json:"participant_quota"`
}

type eventTransitionRequest struct {
    ExpectedVersion int64 `json:"expected_version"`
}

func (request createEventRequest) input() eventdomain.CreateInput {
    return eventdomain.CreateInput{
        EventYear: request.EventYear, Name: request.Name,
        RegistrationOpensAt: request.RegistrationOpensAt, RegistrationClosesAt: request.RegistrationClosesAt,
        ParticipantQuota: request.ParticipantQuota,
    }
}

func (request patchEventRequest) input() (eventdomain.UpdateInput, error) {
    input := eventdomain.UpdateInput{ExpectedVersion: request.ExpectedVersion}
    fields := 0
    if request.Name != nil {
        fields++
        var value string
        if string(request.Name) == "null" || json.Unmarshal(request.Name, &value) != nil {
            return eventdomain.UpdateInput{}, httpx.ErrInvalidJSON
        }
        input.Name = &value
    }
    if request.RegistrationOpensAt != nil {
        fields++
        value, err := nullableTime(request.RegistrationOpensAt)
        if err != nil {
            return eventdomain.UpdateInput{}, err
        }
        input.RegistrationOpensAt = eventdomain.OptionalTime{Set: true, Value: value}
    }
    if request.RegistrationClosesAt != nil {
        fields++
        value, err := nullableTime(request.RegistrationClosesAt)
        if err != nil {
            return eventdomain.UpdateInput{}, err
        }
        input.RegistrationClosesAt = eventdomain.OptionalTime{Set: true, Value: value}
    }
    if request.ParticipantQuota != nil {
        fields++
        value, err := nullableInt64(request.ParticipantQuota)
        if err != nil {
            return eventdomain.UpdateInput{}, err
        }
        input.ParticipantQuota = eventdomain.OptionalInt64{Set: true, Value: value}
    }
    if fields == 0 {
        return eventdomain.UpdateInput{}, eventdomain.ErrNoChanges
    }
    return input, nil
}

func nullableTime(raw json.RawMessage) (*time.Time, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value time.Time
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    value = value.UTC()
    return &value, nil
}

func nullableInt64(raw json.RawMessage) (*int64, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value int64
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    return &value, nil
}

func operationsListInput(c *gin.Context) (eventdomain.ListInput, error) {
    values := c.Request.URL.Query()
    for name := range values {
        if name != "cursor" && name != "limit" {
            return eventdomain.ListInput{}, eventdomain.ErrInvalidInput
        }
    }
    input := eventdomain.ListInput{}
    if cursors, found := values["cursor"]; found {
        if len(cursors) != 1 || cursors[0] == "" {
            return eventdomain.ListInput{}, eventdomain.ErrInvalidCursor
        }
        input.Cursor = cursors[0]
    }
    if limits, found := values["limit"]; found {
        if len(limits) != 1 || limits[0] == "" {
            return eventdomain.ListInput{}, eventdomain.ErrInvalidInput
        }
        limit, err := strconv.Atoi(limits[0])
        if err != nil {
            return eventdomain.ListInput{}, eventdomain.ErrInvalidInput
        }
        input.Limit = limit
    }
    if input.Limit < 0 || input.Limit > 100 {
        return eventdomain.ListInput{}, eventdomain.ErrInvalidInput
    }
    return input, nil
}
