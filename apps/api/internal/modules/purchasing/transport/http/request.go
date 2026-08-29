package http

import (
    "encoding/json"
    "errors"
    "strconv"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
    purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

const publicPurchaseBodyLimit = 64 << 10

type publicStringField struct {
    Value string
    Set   bool
}

type publicPartyDeclarationRequest struct {
    PartyRef    publicStringField `json:"party_ref"`
    DisplayName publicStringField `json:"display_name"`
    Email       publicStringField `json:"email"`
    Phone       publicStringField `json:"phone"`
}

type publicParticipantRequest struct {
    PartyRef    publicStringField `json:"party_ref"`
    DisplayName publicStringField `json:"display_name"`
}

type publicCreatePurchaseRequest struct {
    OfferingID   string                        `json:"offering_id"`
    Purchaser    publicPartyDeclarationRequest `json:"purchaser"`
    Payer        publicPartyDeclarationRequest `json:"payer"`
    Participants []publicParticipantRequest    `json:"participants"`
}

func (field *publicStringField) UnmarshalJSON(data []byte) error {
    field.Set = true
    if string(data) == "null" {
        return errors.New("string fields cannot be null")
    }
    return json.Unmarshal(data, &field.Value)
}

func (request publicCreatePurchaseRequest) relationships() (purchasingdomain.Relationships, error) {
    participants := make([]purchasingdomain.ParticipantInput, 0, len(request.Participants))
    for _, participant := range request.Participants {
        if participant.PartyRef.Set == participant.DisplayName.Set {
            return purchasingdomain.Relationships{}, purchasingdomain.ErrInvalidRelationships
        }
        participants = append(participants, purchasingdomain.ParticipantInput{PartyRef: participant.PartyRef.Value, DisplayName: participant.DisplayName.Value})
    }
    payer := purchasingdomain.PayerInput{Ref: request.Payer.PartyRef.Value}
    if request.Payer.DisplayName.Set || request.Payer.Email.Set || request.Payer.Phone.Set {
        declaration := request.Payer.declaration()
        payer = purchasingdomain.PayerInput{Declaration: &declaration}
    }
    return purchasingdomain.NewRelationships(purchasingdomain.RelationshipsInput{
        Purchaser: request.Purchaser.declaration(), Payer: payer, Participants: participants,
    })
}

func (request publicPartyDeclarationRequest) declaration() purchasingdomain.PartyDeclarationInput {
    return purchasingdomain.PartyDeclarationInput{
        Ref: request.PartyRef.Value,
        Party: identitydomain.CreateInput{
            Type: identitydomain.PartyTypePerson, DisplayName: request.DisplayName.Value,
            Email: publicOptionalString(request.Email), Phone: publicOptionalString(request.Phone),
        },
    }
}

func publicOptionalString(field publicStringField) *string {
    if !field.Set {
        return nil
    }
    value := field.Value
    return &value
}

func operationsListInput(c *gin.Context) (purchasingdomain.ListInput, error) {
    values := c.Request.URL.Query()
    for name := range values {
        if name != "cursor" && name != "limit" && name != "event_id" && name != "status" {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidInput
        }
    }
    input := purchasingdomain.ListInput{}
    if value, found := values["cursor"]; found {
        if len(value) != 1 || value[0] == "" {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidCursor
        }
        input.Cursor = value[0]
    }
    if value, found := values["limit"]; found {
        if len(value) != 1 || value[0] == "" {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidInput
        }
        limit, err := strconv.Atoi(value[0])
        if err != nil {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidInput
        }
        input.Limit = limit
    }
    if value, found := values["event_id"]; found {
        if len(value) != 1 {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidInput
        }
        eventID, err := httpx.CanonicalUUID(value[0])
        if err != nil {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidInput
        }
        input.EventID = eventID
    }
    if value, found := values["status"]; found {
        if len(value) != 1 || value[0] == "" {
            return purchasingdomain.ListInput{}, purchasingdomain.ErrInvalidInput
        }
        input.Status = purchasingdomain.Status(value[0])
    }
    if _, _, err := purchasingdomain.NormalizeListInput(input); err != nil {
        return purchasingdomain.ListInput{}, err
    }
    return input, nil
}
