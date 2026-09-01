package purchasing

type SnapshotPurchaseInput struct {
    Source           CheckoutSource
    PurchaseRef      string
    PurchaserPartyID string
    PayerPartyID     string
    Participants     []PurchaseParticipantInput
    AccessTokenHash  []byte
}

// NewSnapshotPurchase captures the locked Event and Offering values needed for
// a canonical Common Purchase before the caller persists it.
func NewSnapshotPurchase(input SnapshotPurchaseInput) (Purchase, error) {
    if err := validateCheckoutSource(input.Source); err != nil {
        return Purchase{}, err
    }
    total, err := calculateTotalAmount(input.Source.Offering.PriceMinor, len(input.Participants))
    if err != nil {
        return Purchase{}, err
    }
    return NewPurchase(CreateInput{
        EventID:                     input.Source.Event.ID,
        PurchaseRef:                 input.PurchaseRef,
        PurchaserPartyID:            input.PurchaserPartyID,
        PayerPartyID:                input.PayerPartyID,
        OfferingID:                  input.Source.Offering.ID,
        OfferingNameSnapshot:        input.Source.Offering.Name,
        OfferingKindSnapshot:        input.Source.Offering.Kind,
        OfferingUnitPriceMinor:      input.Source.Offering.PriceMinor,
        ParticipantCapacitySnapshot: int64(input.Source.Offering.ParticipantCapacity),
        TotalAmountMinor:            total,
        CurrencyCode:                input.Source.Offering.CurrencyCode,
        Participants:                input.Participants,
        AccessTokenHash:             input.AccessTokenHash,
    })
}
