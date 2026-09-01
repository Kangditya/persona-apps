package app

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "os"
    "strings"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/identity"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
    "github.com/gin-gonic/gin"
)

func TestOperationsPurchaseReadsPostgreSQL(t *testing.T) {
    gin.SetMode(gin.TestMode)
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is required for PostgreSQL Operations coverage")
    }
    database, err := platformdb.Open(dsn)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    acquireOperationsIntegrationLock(t, database)

    ctx := context.Background()
    prefix := fmt.Sprintf("w3purchase%d", time.Now().UnixNano())
    token, csrf := strings.Repeat("p", 43), strings.Repeat("c", 43)
    operatorID := insertOperationsSession(t, database, prefix, token, csrf)
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    authentication := newTestAuthentication(t, database)
    deniedServer, err := NewServer(":0", database, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20}, nil, authentication, nil, nil, purchasing.NewOperationsHandler(nil, logger))
    if err != nil {
        t.Fatal(err)
    }
    operationsRequest(t, deniedServer, "", "", http.MethodGet, "/api/operations/v1/purchases", "", "", http.StatusUnauthorized)
    operationsRequest(t, deniedServer, token, csrf, http.MethodGet, "/api/operations/v1/purchases", "", "", http.StatusForbidden)
    if _, err := database.Exec(`UPDATE operator_sessions SET permission_snapshot = '["purchase.read"]'::jsonb WHERE operator_user_id = $1`, operatorID); err != nil {
        t.Fatal(err)
    }
    var eventID, offeringID, purchaserID, payerID string
    t.Cleanup(func() {
        cleanupPurchaseReadFixture(t, ctx, database, operatorID, eventID, offeringID, purchaserID, payerID)
    })

    parties := identity.NewRepository(database)
    purchaser, err := identity.NewParty(identity.CreateInput{Type: identity.PartyTypePerson, DisplayName: "Purchaser"})
    if err != nil {
        t.Fatal(err)
    }
    purchaser, err = parties.Create(ctx, purchaser)
    if err != nil {
        t.Fatal(err)
    }
    purchaserID = purchaser.ID
    payer, err := identity.NewParty(identity.CreateInput{Type: identity.PartyTypePerson, DisplayName: "Payer"})
    if err != nil {
        t.Fatal(err)
    }
    payer, err = parties.Create(ctx, payer)
    if err != nil {
        t.Fatal(err)
    }
    payerID = payer.ID
    if err := database.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status) VALUES ($1, $2, 'DRAFT') RETURNING id`, availableEventYear(t, database), prefix).Scan(&eventID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, status) VALUES ($1, 'SHARE', 'Stored share', 'SHARE', 100, 'IDR', 1, 'DRAFT') RETURNING id`, eventID).Scan(&offeringID); err != nil {
        t.Fatal(err)
    }
    secretSentinel := "owned-raw-purchase-token-" + prefix
    accessHash := sha256.Sum256([]byte(secretSentinel))
    purchase, err := purchasing.NewPurchase(purchasing.CreateInput{
        EventID: eventID, PurchaseRef: prefix, PurchaserPartyID: purchaserID, PayerPartyID: payerID,
        OfferingID: offeringID, OfferingNameSnapshot: "Stored share", OfferingKindSnapshot: "SHARE",
        OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: 1, TotalAmountMinor: 100, CurrencyCode: "IDR",
        Participants:    []purchasing.PurchaseParticipantInput{{PartyID: &purchaserID, DisplayName: "Stored participant"}},
        AccessTokenHash: accessHash[:],
    })
    if err != nil {
        t.Fatal(err)
    }
    var purchaseID string
    if err := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        created, createErr := purchasing.NewRepository(transaction).Create(ctx, purchase)
        purchaseID = created.ID
        return createErr
    }); err != nil {
        t.Fatal(err)
    }

    server, err := NewServer(":0", database, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20}, nil, authentication, nil, nil, purchasing.NewOperationsHandler(database, logger))
    if err != nil {
        t.Fatal(err)
    }
    listed := operationsRequest(t, server, token, csrf, http.MethodGet, "/api/operations/v1/purchases?event_id="+eventID, "", "", http.StatusOK)
    var page struct {
        Data []map[string]any `json:"data"`
    }
    if err := json.Unmarshal(listed, &page); err != nil {
        t.Fatal(err)
    }
    if len(page.Data) != 1 || page.Data[0]["purchase_ref"] != prefix || page.Data[0]["offering_name_snapshot"] != "Stored share" {
        t.Fatal("Operations Purchase list does not match the stored fixture")
    }
    assertExactJSONKeys(t, page.Data[0], "id", "event_id", "purchase_ref", "channel", "purchaser", "payer", "offering_id", "offering_name_snapshot", "participant_count", "total_amount_minor", "currency_code", "status", "created_at", "updated_at")
    assertExactJSONKeys(t, page.Data[0]["purchaser"].(map[string]any), "id", "display_name")
    assertExactJSONKeys(t, page.Data[0]["payer"].(map[string]any), "id", "display_name")
    detailed := operationsRequest(t, server, token, csrf, http.MethodGet, "/api/operations/v1/purchases/"+purchaseID, "", "", http.StatusOK)
    data := responseData(t, detailed)
    participants, ok := data["participants"].([]any)
    if !ok || len(participants) != 1 || data["purchaser"].(map[string]any)["display_name"] != "Purchaser" || data["payer"].(map[string]any)["display_name"] != "Payer" {
        t.Fatal("Operations Purchase detail does not match the stored fixture")
    }
    assertExactJSONKeys(t, data, "id", "event_id", "purchase_ref", "channel", "purchaser", "payer", "offering_id", "offering_name_snapshot", "participant_count", "participants", "total_amount_minor", "currency_code", "status", "created_at", "updated_at")
    assertExactJSONKeys(t, participants[0].(map[string]any), "sequence_no", "display_name")
    if bytes.Contains(listed, []byte(secretSentinel)) || bytes.Contains(detailed, []byte(secretSentinel)) {
        t.Fatal("Operations Purchase response leaked an owned token sentinel")
    }
    if _, err := database.Exec(`UPDATE operator_sessions SET created_at = now() - interval '2 hours', last_seen_at = now() - interval '2 hours', expires_at = now() - interval '1 hour' WHERE operator_user_id = $1`, operatorID); err != nil {
        t.Fatal(err)
    }
    operationsRequest(t, deniedServer, token, csrf, http.MethodGet, "/api/operations/v1/purchases", "", "", http.StatusUnauthorized)
    if _, err := database.Exec(`UPDATE operator_sessions SET expires_at = now() + interval '1 hour', revoked_at = now() WHERE operator_user_id = $1`, operatorID); err != nil {
        t.Fatal(err)
    }
    operationsRequest(t, deniedServer, token, csrf, http.MethodGet, "/api/operations/v1/purchases", "", "", http.StatusUnauthorized)
}

func cleanupPurchaseReadFixture(t *testing.T, ctx context.Context, database *sql.DB, operatorID, eventID, offeringID, purchaserID, payerID string) {
    t.Helper()
    if eventID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM purchase_status_history WHERE purchase_id IN (SELECT id FROM purchases WHERE event_id = $1)`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchase_participants WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchases WHERE event_id = $1`, eventID)
    }
    if offeringID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM offerings WHERE id = $1`, offeringID)
    }
    if eventID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM qurban_events WHERE id = $1`, eventID)
    }
    for _, partyID := range []string{purchaserID, payerID} {
        if partyID != "" {
            _, _ = database.ExecContext(ctx, `DELETE FROM parties WHERE id = $1`, partyID)
        }
    }
    if operatorID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM operator_sessions WHERE operator_user_id = $1`, operatorID)
        _, _ = database.ExecContext(ctx, `DELETE FROM operator_users WHERE id = $1`, operatorID)
    }
}
