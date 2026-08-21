package purchasing

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
)

func TestOperationsListInput(t *testing.T) {
    tests := []struct {
        target string
        want   ListInput
        valid  bool
    }{
        {target: "/purchases", want: ListInput{}, valid: true},
        {target: "/purchases?limit=10&status=PENDING_PAYMENT&event_id=11111111-1111-1111-1111-111111111111", want: ListInput{Limit: 10, Status: StatusPendingPayment, EventID: "11111111-1111-1111-1111-111111111111"}, valid: true},
        {target: "/purchases?status=DRAFT", valid: false},
        {target: "/purchases?event_id=invalid", valid: false},
        {target: "/purchases?unknown=value", valid: false},
    }
    for _, test := range tests {
        request := httptest.NewRequest(http.MethodGet, test.target, nil)
        context, _ := gin.CreateTestContext(httptest.NewRecorder())
        context.Request = request
        input, err := operationsListInput(context)
        if (err == nil) != test.valid {
            t.Fatalf("operationsListInput(%q) error = %v", test.target, err)
        }
        if err == nil && input != test.want {
            t.Fatalf("operationsListInput(%q) = %#v, want %#v", test.target, input, test.want)
        }
    }
}

func TestToOperationsPurchaseUsesStoredSnapshots(t *testing.T) {
    now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
    detail := Detail{
        Purchase: Purchase{
            ID: "purchase-id", EventID: "event-id", PurchaseRef: "purchase-ref", Channel: ChannelCommon,
            OfferingID: "offering-id", OfferingNameSnapshot: "Stored share", ParticipantCount: 1,
            TotalAmountMinor: 200, CurrencyCode: "IDR", Status: StatusPendingPayment, CreatedAt: now, UpdatedAt: now,
            Participants: []Participant{{SequenceNo: 1, DisplayName: "Stored participant"}},
        },
        Purchaser: PartySummary{ID: "purchaser-id", DisplayName: "Current purchaser"},
        Payer:     &PartySummary{ID: "payer-id", DisplayName: "Current payer"},
    }
    result := toOperationsPurchase(detail, true)
    if result.OfferingNameSnapshot != "Stored share" || len(result.Participants) != 1 || result.Participants[0].DisplayName != "Stored participant" || result.Payer == nil || result.Payer.DisplayName != "Current payer" {
        t.Fatalf("operations purchase = %#v", result)
    }
}
