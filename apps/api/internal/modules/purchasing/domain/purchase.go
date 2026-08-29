package domain

import (
    "errors"
    "fmt"
    "strings"
    "time"
    "unicode/utf8"
)

const (
    MaxSafeInteger              = int64(9_007_199_254_740_991)
    MaxParticipantCapacity      = int64(2_147_483_647)
    MaxOfferingNameLength       = 255
    MaxOfferingKindLength       = 64
    MaxParticipantDisplayLength = 255
)

var (
    ErrNotFound           = errors.New("purchase not found")
    ErrInvalidInput       = errors.New("invalid purchase input")
    ErrInvalidCursor      = errors.New("invalid purchase cursor")
    ErrConflict           = errors.New("purchase conflict")
    ErrDuplicateReference = errors.New("duplicate purchase reference")
    ErrStateConflict      = errors.New("purchase state conflict")
    ErrQuotaUnavailable   = errors.New("purchase quota unavailable")
)

type Channel string

const ChannelCommon Channel = "COMMON"

type Status string

const (
    StatusDraft          Status = "DRAFT"
    StatusPendingPayment Status = "PENDING_PAYMENT"
    StatusPaid           Status = "PAID"
    StatusEligible       Status = "ELIGIBLE"
    StatusAllocated      Status = "ALLOCATED"
    StatusCompleted      Status = "COMPLETED"
    StatusCancelled      Status = "CANCELLED"
)

type PurchaseParticipantInput struct {
    PartyID     *string
    DisplayName string
}

type Participant struct {
    ID          string
    PartyID     *string
    SequenceNo  int
    DisplayName string
}

type CreateInput struct {
    EventID                     string
    PurchaseRef                 string
    PurchaserPartyID            string
    PayerPartyID                string
    OfferingID                  string
    OfferingNameSnapshot        string
    OfferingKindSnapshot        string
    OfferingUnitPriceMinor      int64
    ParticipantCapacitySnapshot int64
    TotalAmountMinor            int64
    CurrencyCode                string
    Participants                []PurchaseParticipantInput
    AccessTokenHash             []byte
}

type Purchase struct {
    ID                          string
    EventID                     string
    PurchaseRef                 string
    Channel                     Channel
    PurchaserPartyID            string
    PayerPartyID                *string
    OfferingID                  string
    OfferingNameSnapshot        string
    OfferingKindSnapshot        string
    OfferingUnitPriceMinor      int64
    ParticipantCapacitySnapshot int32
    ParticipantCount            int
    TotalAmountMinor            int64
    CurrencyCode                string
    Status                      Status
    Version                     int64
    Participants                []Participant
    CreatedAt                   time.Time
    UpdatedAt                   time.Time
    accessTokenHash             []byte
}

func (purchase Purchase) AccessTokenHash() []byte {
    return append([]byte(nil), purchase.accessTokenHash...)
}

const (
    DefaultListLimit = 50
    MaxListLimit     = 100
)

type PartySummary struct {
    ID          string
    DisplayName string
}

type Detail struct {
    Purchase  Purchase
    Purchaser PartySummary
    Payer     *PartySummary
}

type ListInput struct {
    EventID string
    Status  Status
    Cursor  string
    Limit   int
}

type ListResult struct {
    Purchases  []Detail
    NextCursor string
    Limit      int
}

func NewPurchase(input CreateInput) (Purchase, error) {
    if input.ParticipantCapacitySnapshot < 1 || input.ParticipantCapacitySnapshot > MaxParticipantCapacity {
        return Purchase{}, fmt.Errorf("%w: participant capacity must be between 1 and %d", ErrInvalidInput, MaxParticipantCapacity)
    }
    purchase := Purchase{
        EventID:                     strings.TrimSpace(input.EventID),
        PurchaseRef:                 strings.TrimSpace(input.PurchaseRef),
        Channel:                     ChannelCommon,
        PurchaserPartyID:            strings.TrimSpace(input.PurchaserPartyID),
        OfferingID:                  strings.TrimSpace(input.OfferingID),
        OfferingNameSnapshot:        strings.TrimSpace(input.OfferingNameSnapshot),
        OfferingKindSnapshot:        strings.TrimSpace(input.OfferingKindSnapshot),
        OfferingUnitPriceMinor:      input.OfferingUnitPriceMinor,
        ParticipantCapacitySnapshot: int32(input.ParticipantCapacitySnapshot),
        ParticipantCount:            len(input.Participants),
        TotalAmountMinor:            input.TotalAmountMinor,
        CurrencyCode:                input.CurrencyCode,
        Status:                      StatusPendingPayment,
        Version:                     1,
        accessTokenHash:             append([]byte(nil), input.AccessTokenHash...),
    }
    payerID := strings.TrimSpace(input.PayerPartyID)
    purchase.PayerPartyID = &payerID
    if err := validatePurchase(&purchase); err != nil {
        return Purchase{}, err
    }
    purchase.Participants = make([]Participant, 0, len(input.Participants))
    for index, inputParticipant := range input.Participants {
        participant, err := newParticipant(index+1, inputParticipant)
        if err != nil {
            return Purchase{}, err
        }
        purchase.Participants = append(purchase.Participants, participant)
    }
    return purchase, nil
}

func validatePurchase(purchase *Purchase) error {
    if purchase.EventID == "" || purchase.OfferingID == "" {
        return fmt.Errorf("%w: event and offering IDs are required", ErrInvalidInput)
    }
    if purchase.PurchaseRef == "" {
        return fmt.Errorf("%w: purchase reference is required", ErrInvalidInput)
    }
    if purchase.PurchaserPartyID == "" || purchase.PayerPartyID == nil || *purchase.PayerPartyID == "" {
        return fmt.Errorf("%w: purchaser and payer are required for a common purchase", ErrInvalidInput)
    }
    if purchase.OfferingNameSnapshot == "" || utf8.RuneCountInString(purchase.OfferingNameSnapshot) > MaxOfferingNameLength || purchase.OfferingKindSnapshot == "" || utf8.RuneCountInString(purchase.OfferingKindSnapshot) > MaxOfferingKindLength {
        return fmt.Errorf("%w: offering snapshots are required", ErrInvalidInput)
    }
    if purchase.OfferingUnitPriceMinor < 0 || purchase.OfferingUnitPriceMinor > MaxSafeInteger || purchase.TotalAmountMinor < 0 || purchase.TotalAmountMinor > MaxSafeInteger {
        return fmt.Errorf("%w: money values must be between 0 and %d", ErrInvalidInput, MaxSafeInteger)
    }
    expectedTotal, err := calculateTotalAmount(purchase.OfferingUnitPriceMinor, purchase.ParticipantCount)
    if err != nil {
        return err
    }
    if purchase.TotalAmountMinor != expectedTotal {
        return fmt.Errorf("%w: total amount must equal unit price times participant count", ErrInvalidInput)
    }
    if purchase.ParticipantCapacitySnapshot < 1 || int64(purchase.ParticipantCapacitySnapshot) > MaxParticipantCapacity || purchase.ParticipantCount < 1 || int64(purchase.ParticipantCount) > int64(purchase.ParticipantCapacitySnapshot) {
        return fmt.Errorf("%w: participant count must fit the offering capacity", ErrInvalidInput)
    }
    if !validCurrency(purchase.CurrencyCode) {
        return fmt.Errorf("%w: currency code must be three uppercase letters", ErrInvalidInput)
    }
    if len(purchase.accessTokenHash) != 32 {
        return fmt.Errorf("%w: common purchase access-token hash must be 32 bytes", ErrInvalidInput)
    }
    return nil
}

func calculateTotalAmount(unitPrice int64, participantCount int) (int64, error) {
    if unitPrice < 0 || unitPrice > MaxSafeInteger || participantCount < 1 || int64(participantCount) > MaxSafeInteger {
        return 0, ErrInvalidInput
    }
    count := int64(participantCount)
    if unitPrice != 0 && count > MaxSafeInteger/unitPrice {
        return 0, fmt.Errorf("%w: total amount exceeds %d", ErrInvalidInput, MaxSafeInteger)
    }
    return unitPrice * count, nil
}

func newParticipant(sequenceNo int, input PurchaseParticipantInput) (Participant, error) {
    displayName := strings.TrimSpace(input.DisplayName)
    if displayName == "" || utf8.RuneCountInString(displayName) > MaxParticipantDisplayLength {
        return Participant{}, fmt.Errorf("%w: participant display name must contain 1 to %d characters", ErrInvalidInput, MaxParticipantDisplayLength)
    }
    var partyID *string
    if input.PartyID != nil {
        value := strings.TrimSpace(*input.PartyID)
        if value == "" {
            return Participant{}, fmt.Errorf("%w: participant party ID must not be blank", ErrInvalidInput)
        }
        partyID = &value
    }
    return Participant{PartyID: partyID, SequenceNo: sequenceNo, DisplayName: displayName}, nil
}

func validCurrency(value string) bool {
    if len(value) != 3 {
        return false
    }
    for _, character := range value {
        if character < 'A' || character > 'Z' {
            return false
        }
    }
    return true
}
