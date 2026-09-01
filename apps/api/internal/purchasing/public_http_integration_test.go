package purchasing

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

func TestPublicPurchaseCheckoutRetriesDuplicateReferencePostgreSQL(t *testing.T) {
    gin.SetMode(gin.TestMode)
    database := integrationDatabase(t)
    ctx := context.Background()
    acquireActiveEventIntegrationLock(t, ctx, database)
    prefix := fmt.Sprintf("public-reference-retry-%d", time.Now().UnixNano())
    eventID, offeringID, partyID := createReservationFixture(t, ctx, database, prefix, nil, nil)
    email := prefix + "@example.test"
    key := prefix + "-key"
    t.Cleanup(func() {
        _, _ = database.ExecContext(ctx, `DELETE FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, "storefront.purchase.create.guest", key)
        _, _ = database.ExecContext(ctx, `DELETE FROM outbox_events WHERE aggregate_id IN (SELECT id FROM purchases WHERE event_id = $1)`, eventID)
        cleanupReservationFixture(t, ctx, database, eventID, partyID)
        _, _ = database.ExecContext(ctx, `DELETE FROM parties WHERE email = $1`, email)
    })

    var eventYear int
    if err := database.QueryRowContext(ctx, `SELECT event_year FROM qurban_events WHERE id = $1`, eventID).Scan(&eventYear); err != nil {
        t.Fatal(err)
    }
    collidingReference, err := newPublicPurchaseReference(eventYear, bytes.NewReader(make([]byte, publicPurchaseReferenceBytes)))
    if err != nil {
        t.Fatal(err)
    }
    if _, _, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, collidingReference, time.Now().UTC()); err != nil {
        t.Fatal(err)
    }

    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "reference-retry", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)))
    randomBytes := make([]byte, 0, 2*(publicPurchaseReferenceBytes+publicPurchaseAccessTokenBytes))
    randomBytes = append(randomBytes, make([]byte, publicPurchaseReferenceBytes)...)
    randomBytes = append(randomBytes, bytes.Repeat([]byte{2}, publicPurchaseAccessTokenBytes)...)
    randomBytes = append(randomBytes, bytes.Repeat([]byte{1}, publicPurchaseReferenceBytes)...)
    randomBytes = append(randomBytes, bytes.Repeat([]byte{3}, publicPurchaseAccessTokenBytes)...)
    random := bytes.NewReader(randomBytes)
    handler.random = random
    router := gin.New()
    handler.RegisterRoutes(router.Group(""))

    body := fmt.Sprintf(`{"offering_id":%q,"purchaser":{"party_ref":"purchaser","display_name":"Retry purchaser","email":%q},"payer":{"party_ref":"purchaser"},"participants":[{"party_ref":"purchaser"}]}`, offeringID, email)
    request := httptest.NewRequest(http.MethodPost, "/purchases", bytes.NewBufferString(body))
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Idempotency-Key", key)
    response := httptest.NewRecorder()
    router.ServeHTTP(response, request)
    if response.Code != http.StatusCreated {
        t.Fatalf("checkout status = %d, want %d", response.Code, http.StatusCreated)
    }
    var payload struct {
        Data struct {
            Purchase struct {
                PurchaseRef string `json:"purchase_ref"`
            } `json:"purchase"`
        } `json:"data"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if payload.Data.Purchase.PurchaseRef == "" || payload.Data.Purchase.PurchaseRef == collidingReference {
        t.Fatalf("purchase reference = %q, collision = %q", payload.Data.Purchase.PurchaseRef, collidingReference)
    }
    if random.Len() != 0 {
        t.Fatalf("reference retry did not consume both entropy sequences: %d bytes remain", random.Len())
    }

    var purchaseID string
    if err := database.QueryRowContext(ctx, `SELECT id FROM purchases WHERE purchase_ref = $1`, payload.Data.Purchase.PurchaseRef).Scan(&purchaseID); err != nil {
        t.Fatal(err)
    }
    var purchases, parties, history, participants, reservations, outbox, records int
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM purchases WHERE event_id = $1`, eventID).Scan(&purchases); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM parties WHERE email = $1`, email).Scan(&parties); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM purchase_status_history WHERE purchase_id = $1`, purchaseID).Scan(&history); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM purchase_participants WHERE purchase_id = $1`, purchaseID).Scan(&participants); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM quota_reservations WHERE purchase_id = $1`, purchaseID).Scan(&reservations); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id = $1`, purchaseID).Scan(&outbox); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, "storefront.purchase.create.guest", key).Scan(&records); err != nil {
        t.Fatal(err)
    }
    if purchases != 2 || parties != 1 || history != 1 || participants != 1 || reservations != 1 || outbox != 2 || records != 1 {
        t.Fatalf("reference retry effects = purchases:%d parties:%d history:%d participants:%d reservations:%d outbox:%d idempotency_records:%d", purchases, parties, history, participants, reservations, outbox, records)
    }
}
