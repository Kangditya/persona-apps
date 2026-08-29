package http

import (
    "encoding/json"
    "strconv"

    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

const offeringCommandBodyLimit = 64 << 10

type createOfferingRequest struct {
    Code                string  `json:"code"`
    Name                string  `json:"name"`
    Kind                string  `json:"offering_kind"`
    Description         *string `json:"description"`
    PriceMinor          int64   `json:"price_minor"`
    CurrencyCode        string  `json:"currency_code"`
    ParticipantCapacity int64   `json:"participant_capacity"`
    ParticipantQuota    *int64  `json:"participant_quota"`
}

type patchOfferingRequest struct {
    ExpectedVersion  int64           `json:"expected_version"`
    Name             json.RawMessage `json:"name"`
    Description      json.RawMessage `json:"description"`
    PriceMinor       json.RawMessage `json:"price_minor"`
    ParticipantQuota json.RawMessage `json:"participant_quota"`
}

type offeringTransitionRequest struct {
    ExpectedVersion int64 `json:"expected_version"`
}

func (request createOfferingRequest) input(eventID string) offeringdomain.CreateInput {
    return offeringdomain.CreateInput{
        EventID: eventID, Code: request.Code, Name: request.Name, Kind: request.Kind,
        Description: request.Description, PriceMinor: request.PriceMinor, CurrencyCode: request.CurrencyCode,
        ParticipantCapacity: request.ParticipantCapacity, ParticipantQuota: request.ParticipantQuota,
    }
}

func (request patchOfferingRequest) input() (offeringdomain.UpdateInput, error) {
    input := offeringdomain.UpdateInput{ExpectedVersion: request.ExpectedVersion}
    fields := 0
    if request.Name != nil {
        fields++
        var value string
        if string(request.Name) == "null" || json.Unmarshal(request.Name, &value) != nil {
            return offeringdomain.UpdateInput{}, httpx.ErrInvalidJSON
        }
        input.Name = &value
    }
    if request.Description != nil {
        fields++
        value, err := requestNullableString(request.Description)
        if err != nil {
            return offeringdomain.UpdateInput{}, err
        }
        input.Description = offeringdomain.OptionalString{Set: true, Value: value}
    }
    if request.PriceMinor != nil {
        fields++
        var value int64
        if json.Unmarshal(request.PriceMinor, &value) != nil {
            return offeringdomain.UpdateInput{}, httpx.ErrInvalidJSON
        }
        input.PriceMinor = &value
    }
    if request.ParticipantQuota != nil {
        fields++
        value, err := requestNullableInt64(request.ParticipantQuota)
        if err != nil {
            return offeringdomain.UpdateInput{}, err
        }
        input.ParticipantQuota = offeringdomain.OptionalInt64{Set: true, Value: value}
    }
    if fields == 0 {
        return offeringdomain.UpdateInput{}, offeringdomain.ErrNoChanges
    }
    return input, nil
}

func requestNullableString(raw json.RawMessage) (*string, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value string
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    return &value, nil
}

func requestNullableInt64(raw json.RawMessage) (*int64, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value int64
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    return &value, nil
}

func offeringListInput(c *gin.Context) (offeringdomain.ListInput, error) {
    values := c.Request.URL.Query()
    for name := range values {
        if name != "cursor" && name != "limit" {
            return offeringdomain.ListInput{}, offeringdomain.ErrInvalidInput
        }
    }
    input := offeringdomain.ListInput{}
    if cursors, found := values["cursor"]; found {
        if len(cursors) != 1 || cursors[0] == "" {
            return offeringdomain.ListInput{}, offeringdomain.ErrInvalidCursor
        }
        input.Cursor = cursors[0]
    }
    if limits, found := values["limit"]; found {
        if len(limits) != 1 || limits[0] == "" {
            return offeringdomain.ListInput{}, offeringdomain.ErrInvalidInput
        }
        limit, err := strconv.Atoi(limits[0])
        if err != nil {
            return offeringdomain.ListInput{}, offeringdomain.ErrInvalidInput
        }
        input.Limit = limit
    }
    if _, _, err := offeringdomain.NormalizeListInput(input); err != nil {
        return offeringdomain.ListInput{}, err
    }
    return input, nil
}
