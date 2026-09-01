package app

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "os"
    "regexp"
    "strings"
    "sync"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
    "github.com/gin-gonic/gin"
)

const publicQuotaRaceRepetitions = 3

func TestPublicCommonPurchaseCheckoutPostgreSQL(t *testing.T) {
    gin.SetMode(gin.TestMode)
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is required for PostgreSQL public checkout coverage")
    }
    database, err := platformdb.Open(dsn)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    acquireOperationsIntegrationLock(t, database)

    ctx := context.Background()
    prefix := fmt.Sprintf("w3public%d", time.Now().UnixNano())
    email := prefix + "@example.test"
    var eventID, offeringID string
    t.Cleanup(func() { cleanupPublicCheckoutFixture(t, ctx, database, prefix, eventID) })
    if err := database.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status, participant_quota) VALUES ($1, $2, 'ACTIVE', $3) RETURNING id`, availableEventYear(t, database), prefix, 2+publicQuotaRaceRepetitions).Scan(&eventID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, participant_quota, status) VALUES ($1, 'SHARE', 'Public share', 'SHARE', 100, 'IDR', 2, 2, 'PUBLISHED') RETURNING id`, eventID).Scan(&offeringID); err != nil {
        t.Fatal(err)
    }

    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "public-test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    logOutput := &bytes.Buffer{}
    logger := slog.New(slog.NewTextHandler(logOutput, nil))
    const origin = "https://storefront.example.test"
    server, err := NewServer(":0", database, logger, config.PublicConfig{
        RateLimitPerMinute: 600, RateLimitBurst: 100, StorefrontAllowedOrigins: map[string]struct{}{origin: {}},
    }, purchasing.NewPublicHandler(database, cipher, logger), nil, nil, nil, nil)
    if err != nil {
        t.Fatal(err)
    }
    body := fmt.Sprintf(`{"offering_id":%q,"purchaser":{"party_ref":"purchaser","display_name":"Siti Aminah","email":%q},"payer":{"party_ref":"purchaser"},"participants":[{"party_ref":"purchaser"},{"display_name":"Ahmad"}]}`, offeringID, email)
    key := prefix + ".checkout"
    validationCases := []struct {
        name        string
        key         string
        body        string
        contentType string
        status      int
        code        string
    }{
        {name: "missing idempotency key", body: body, contentType: "application/json", status: http.StatusBadRequest, code: "invalid_request"},
        {name: "unknown JSON field", key: prefix + ".invalid-json", body: strings.Replace(body, `"participants":`, `"unexpected":true,"participants":`, 1), contentType: "application/json", status: http.StatusBadRequest, code: "invalid_request"},
        {name: "bounded body", key: prefix + ".too-large", body: `{"padding":"` + strings.Repeat("x", 65<<10) + `"}`, contentType: "application/json", status: http.StatusRequestEntityTooLarge, code: "request_too_large"},
        {name: "JSON media type", key: prefix + ".media", body: body, contentType: "text/plain", status: http.StatusUnsupportedMediaType, code: "unsupported_media_type"},
        {name: "canonical UUID", key: prefix + ".uuid", body: strings.Replace(body, offeringID, "ABCDEFAB-CDEF-4ABC-8DEF-ABCDEFABCDEF", 1), contentType: "application/json", status: http.StatusUnprocessableEntity, code: "validation_failed"},
        {name: "exclusive participant relationship", key: prefix + ".relationship", body: strings.Replace(body, `{"party_ref":"purchaser"},{"display_name":"Ahmad"}`, `{"party_ref":"purchaser","display_name":"Siti Aminah"},{"display_name":"Ahmad"}`, 1), contentType: "application/json", status: http.StatusUnprocessableEntity, code: "validation_failed"},
    }
    for _, testCase := range validationCases {
        t.Run("validation/"+testCase.name, func(t *testing.T) {
            response := publicCheckoutRequestWithContentType(server, testCase.key, testCase.body, testCase.contentType, origin)
            assertPublicCheckoutError(t, response, testCase.status, testCase.code)
            assertNoPublicCheckoutEffects(t, ctx, database, eventID, email, prefix)
        })
    }

    type result struct {
        body   []byte
        status int
        header http.Header
    }
    start := make(chan struct{})
    results := make(chan result, 2)
    var group sync.WaitGroup
    for range 2 {
        group.Add(1)
        go func() {
            defer group.Done()
            <-start
            response := publicCheckoutRequest(server, key, body, origin)
            results <- result{body: response.Body.Bytes(), status: response.Code, header: response.Header().Clone()}
        }()
    }
    close(start)
    group.Wait()
    close(results)

    var first []byte
    for result := range results {
        if result.status != http.StatusCreated {
            t.Fatalf("concurrent checkout status = %d, want %d", result.status, http.StatusCreated)
        }
        assertPublicCheckoutHeaders(t, result.header, origin)
        if first == nil {
            first = result.body
        } else if !bytes.Equal(first, result.body) {
            t.Fatal("concurrent replay response bytes differ")
        }
    }
    semanticBody := fmt.Sprintf(`{
        "participants": [{"party_ref": "purchaser"}, {"display_name": "Ahmad"}],
        "payer": {"party_ref": "purchaser"},
        "purchaser": {"email": %q, "display_name": "Siti Aminah", "party_ref": "purchaser"},
        "offering_id": %q
    }`, email, offeringID)
    replay := publicCheckoutRequest(server, key, semanticBody, origin)
    if replay.Code != http.StatusCreated || !bytes.Equal(first, replay.Body.Bytes()) {
        t.Fatalf("semantic replay status = %d or response bytes differ", replay.Code)
    }
    assertPublicCheckoutHeaders(t, replay.Header(), origin)

    changed := publicCheckoutRequest(server, key, strings.Replace(semanticBody, "Ahmad", "Budi", 1), origin)
    assertPublicCheckoutError(t, changed, http.StatusConflict, "idempotency_conflict")
    quota := publicCheckoutRequest(server, prefix+".quota", body, origin)
    assertPublicCheckoutError(t, quota, http.StatusConflict, "quota_unavailable")

    preflight := httptest.NewRecorder()
    preflightRequest := httptest.NewRequest(http.MethodOptions, "/api/public/v1/purchases", nil)
    preflightRequest.Header.Set("Origin", origin)
    preflightRequest.Header.Set("Access-Control-Request-Method", http.MethodPost)
    preflightRequest.Header.Set("Access-Control-Request-Headers", "content-type, idempotency-key, x-request-id")
    server.Handler.ServeHTTP(preflight, preflightRequest)
    if preflight.Code != http.StatusNoContent || preflight.Header().Get("Access-Control-Allow-Origin") != origin || preflight.Header().Get("Access-Control-Allow-Methods") != "GET, POST" || preflight.Header().Get("Access-Control-Allow-Headers") != "X-Request-ID, Content-Type, Idempotency-Key" || preflight.Header().Get("Access-Control-Allow-Credentials") != "" {
        t.Fatalf("public preflight = %d %#v", preflight.Code, preflight.Header())
    }

    var response struct {
        Data struct {
            AccessToken string `json:"access_token"`
            Purchase    struct {
                ID               string `json:"id"`
                PurchaseRef      string `json:"purchase_ref"`
                ParticipantCount int    `json:"participant_count"`
                TotalAmountMinor int64  `json:"total_amount_minor"`
                Status           string `json:"status"`
                Participants     []struct {
                    DisplayName string `json:"display_name"`
                } `json:"participants"`
            } `json:"purchase"`
        } `json:"data"`
    }
    if err := json.Unmarshal(first, &response); err != nil {
        t.Fatal(err)
    }
    if len(response.Data.AccessToken) != 43 || response.Data.Purchase.ID == "" || response.Data.Purchase.ParticipantCount != 2 || response.Data.Purchase.TotalAmountMinor != 200 || response.Data.Purchase.Status != "PENDING_PAYMENT" || len(response.Data.Purchase.Participants) != 2 || response.Data.Purchase.Participants[0].DisplayName != "Siti Aminah" || response.Data.Purchase.Participants[1].DisplayName != "Ahmad" {
        t.Fatal("checkout response does not match the safe public contract")
    }
    assertPublicCheckoutResponseShape(t, first)

    var purchaseID, purchaseRef, status, currency string
    var total int64
    var capacity int32
    var tokenHash []byte
    if err := database.QueryRowContext(ctx, `SELECT id, purchase_ref, status, total_amount_minor, currency_code, participant_capacity_snapshot, access_token_hash FROM purchases WHERE event_id = $1`, eventID).Scan(&purchaseID, &purchaseRef, &status, &total, &currency, &capacity, &tokenHash); err != nil {
        t.Fatal(err)
    }
    if !regexp.MustCompile(`^QRB-[0-9]{4}-[A-Z2-7]{16}$`).MatchString(purchaseRef) || purchaseRef != response.Data.Purchase.PurchaseRef || status != "PENDING_PAYMENT" || total != 200 || currency != "IDR" || capacity != 2 {
        t.Fatalf("stored purchase = %q/%q/%d/%q/%d", purchaseRef, status, total, currency, capacity)
    }
    wantTokenHash := sha256.Sum256([]byte(response.Data.AccessToken))
    if !bytes.Equal(tokenHash, wantTokenHash[:]) {
        t.Fatalf("stored access-token hash = %x", tokenHash)
    }

    var purchases, parties, history, reservations int
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM purchases WHERE event_id = $1`, eventID).Scan(&purchases); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM parties WHERE email = $1`, email).Scan(&parties); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM purchase_status_history WHERE purchase_id = $1`, purchaseID).Scan(&history); err != nil {
        t.Fatal(err)
    }
    var participants, linkedParticipants int
    if err := database.QueryRowContext(ctx, `SELECT count(*), count(party_id) FROM purchase_participants WHERE purchase_id = $1`, purchaseID).Scan(&participants, &linkedParticipants); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM quota_reservations WHERE purchase_id = $1 AND status = 'RESERVED' AND participant_units = 2`, purchaseID).Scan(&reservations); err != nil {
        t.Fatal(err)
    }
    if purchases != 1 || parties != 1 || history != 1 || participants != 2 || linkedParticipants != 1 || reservations != 1 {
        t.Fatalf("purchase effects = purchases:%d parties:%d history:%d participants:%d links:%d reservations:%d", purchases, parties, history, participants, linkedParticipants, reservations)
    }

    rows, err := database.QueryContext(ctx, `SELECT event_type, payload::text FROM outbox_events WHERE aggregate_id = $1 ORDER BY event_type`, purchaseID)
    if err != nil {
        t.Fatal(err)
    }
    defer rows.Close()
    eventTypes := make([]string, 0, 2)
    for rows.Next() {
        var eventType, payload string
        if err := rows.Scan(&eventType, &payload); err != nil {
            t.Fatal(err)
        }
        if strings.Contains(payload, response.Data.AccessToken) || strings.Contains(payload, email) || strings.Contains(payload, key) {
            t.Fatal("outbox payload leaked an owned secret sentinel")
        }
        eventTypes = append(eventTypes, eventType)
    }
    if err := rows.Err(); err != nil {
        t.Fatal(err)
    }
    if strings.Join(eventTypes, ",") != "PurchasePendingPayment,PurchaseReserved" {
        t.Fatalf("outbox event types = %v", eventTypes)
    }

    var expiresAt sql.NullTime
    var encrypted string
    if err := database.QueryRowContext(ctx, `SELECT expires_at, response_body::text FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, "storefront.purchase.create.guest", key).Scan(&expiresAt, &encrypted); err != nil {
        t.Fatal(err)
    }
    if expiresAt.Valid || strings.Contains(encrypted, response.Data.AccessToken) || strings.Contains(encrypted, email) || strings.Contains(encrypted, key) {
        t.Fatal("durable replay record expired or leaked an owned secret sentinel")
    }
    runPublicQuotaContention(t, ctx, database, server, eventID, prefix, origin)
    if strings.Contains(logOutput.String(), response.Data.AccessToken) || strings.Contains(logOutput.String(), prefix) {
        t.Fatal("public checkout logs leaked an owned secret sentinel")
    }
}

func publicCheckoutRequest(server *http.Server, key, body, origin string) *httptest.ResponseRecorder {
    return publicCheckoutRequestWithContentType(server, key, body, "application/json", origin)
}

func publicCheckoutRequestWithContentType(server *http.Server, key, body, contentType, origin string) *httptest.ResponseRecorder {
    request := httptest.NewRequest(http.MethodPost, "/api/public/v1/purchases", strings.NewReader(body))
    if contentType != "" {
        request.Header.Set("Content-Type", contentType)
    }
    if key != "" {
        request.Header.Set("Idempotency-Key", key)
    }
    if origin != "" {
        request.Header.Set("Origin", origin)
    }
    response := httptest.NewRecorder()
    server.Handler.ServeHTTP(response, request)
    return response
}

func assertPublicCheckoutHeaders(t *testing.T, header http.Header, origin string) {
    t.Helper()
    if header.Get("Cache-Control") != "no-store" || header.Get("X-Request-ID") == "" || header.Get("Access-Control-Allow-Origin") != origin || header.Get("Access-Control-Allow-Credentials") != "" {
        t.Fatalf("public checkout headers = %#v", header)
    }
}

func assertPublicCheckoutError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
    t.Helper()
    if response.Code != status {
        t.Fatalf("status = %d, want %d", response.Code, status)
    }
    var payload struct {
        Error struct {
            Code string `json:"code"`
        } `json:"error"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if payload.Error.Code != code {
        t.Fatalf("error code = %q, want %q", payload.Error.Code, code)
    }
}

func assertNoPublicCheckoutEffects(t *testing.T, ctx context.Context, database *sql.DB, eventID, email, prefix string) {
    t.Helper()
    var purchases, parties, history, participants, reservations, outbox, replays int
    err := database.QueryRowContext(ctx, `
        SELECT
            (SELECT count(*) FROM purchases WHERE event_id = $1),
            (SELECT count(*) FROM parties WHERE email = $2),
            (SELECT count(*) FROM purchase_status_history WHERE purchase_id IN (SELECT id FROM purchases WHERE event_id = $1)),
            (SELECT count(*) FROM purchase_participants WHERE event_id = $1),
            (SELECT count(*) FROM quota_reservations WHERE event_id = $1),
            (SELECT count(*) FROM outbox_events WHERE aggregate_id IN (SELECT id FROM purchases WHERE event_id = $1)),
            (SELECT count(*) FROM idempotency_records WHERE namespace = 'storefront.purchase.create.guest' AND idempotency_key LIKE $3)
    `, eventID, email, prefix+"%").Scan(&purchases, &parties, &history, &participants, &reservations, &outbox, &replays)
    if err != nil {
        t.Fatal(err)
    }
    if purchases != 0 || parties != 0 || history != 0 || participants != 0 || reservations != 0 || outbox != 0 || replays != 0 {
        t.Fatalf("rejected checkout effects = purchases:%d parties:%d history:%d participants:%d reservations:%d outbox:%d replays:%d", purchases, parties, history, participants, reservations, outbox, replays)
    }
}

func assertPublicCheckoutResponseShape(t *testing.T, body []byte) {
    t.Helper()
    var envelope map[string]any
    if err := json.Unmarshal(body, &envelope); err != nil {
        t.Fatal(err)
    }
    assertExactJSONKeys(t, envelope, "data")
    data, ok := envelope["data"].(map[string]any)
    if !ok {
        t.Fatal("public checkout data is not an object")
    }
    assertExactJSONKeys(t, data, "purchase", "access_token")
    purchase, ok := data["purchase"].(map[string]any)
    if !ok {
        t.Fatal("public checkout Purchase is not an object")
    }
    assertExactJSONKeys(t, purchase, "id", "purchase_ref", "channel", "offering", "participant_count", "participants", "total_amount_minor", "currency_code", "status", "reservation_expires_at", "created_at")
    offering, ok := purchase["offering"].(map[string]any)
    if !ok {
        t.Fatal("public checkout Offering is not an object")
    }
    assertExactJSONKeys(t, offering, "id", "name", "kind", "unit_price_minor", "participant_capacity")
    participants, ok := purchase["participants"].([]any)
    if !ok || len(participants) == 0 {
        t.Fatal("public checkout participants are missing")
    }
    for _, value := range participants {
        participant, ok := value.(map[string]any)
        if !ok {
            t.Fatal("public checkout participant is not an object")
        }
        assertExactJSONKeys(t, participant, "sequence_no", "display_name")
    }
}

func runPublicQuotaContention(t *testing.T, ctx context.Context, database *sql.DB, server *http.Server, eventID, prefix, origin string) {
    t.Helper()
    for repetition := 0; repetition < publicQuotaRaceRepetitions; repetition++ {
        var offeringID string
        code := fmt.Sprintf("RACE-%d", repetition)
        if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, participant_quota, status) VALUES ($1, $2, 'Last unit', 'SHARE', 100, 'IDR', 1, 1, 'PUBLISHED') RETURNING id`, eventID, code).Scan(&offeringID); err != nil {
            t.Fatal(err)
        }

        emailA := fmt.Sprintf("%s.race%d.a@example.test", prefix, repetition)
        emailB := fmt.Sprintf("%s.race%d.b@example.test", prefix, repetition)
        keyA := fmt.Sprintf("%s.race.%d.a", prefix, repetition)
        keyB := fmt.Sprintf("%s.race.%d.b", prefix, repetition)
        bodyA := fmt.Sprintf(`{"offering_id":%q,"purchaser":{"party_ref":"purchaser","display_name":"Race A","email":%q},"payer":{"party_ref":"purchaser"},"participants":[{"party_ref":"purchaser"}]}`, offeringID, emailA)
        bodyB := fmt.Sprintf(`{"offering_id":%q,"purchaser":{"party_ref":"purchaser","display_name":"Race B","email":%q},"payer":{"party_ref":"purchaser"},"participants":[{"party_ref":"purchaser"}]}`, offeringID, emailB)
        start := make(chan struct{})
        responses := make(chan *httptest.ResponseRecorder, 2)
        for _, intent := range []struct{ key, body string }{{keyA, bodyA}, {keyB, bodyB}} {
            intent := intent
            go func() {
                <-start
                responses <- publicCheckoutRequest(server, intent.key, intent.body, origin)
            }()
        }
        close(start)

        successes, conflicts := 0, 0
        for range 2 {
            select {
            case response := <-responses:
                switch response.Code {
                case http.StatusCreated:
                    successes++
                    assertPublicCheckoutHeaders(t, response.Header(), origin)
                case http.StatusConflict:
                    conflicts++
                    assertPublicCheckoutError(t, response, http.StatusConflict, "quota_unavailable")
                default:
                    t.Fatalf("quota race status = %d", response.Code)
                }
            case <-time.After(10 * time.Second):
                t.Fatal("quota race exceeded 10 seconds")
            }
        }
        if successes != 1 || conflicts != 1 {
            t.Fatalf("quota race successes/conflicts = %d/%d, want 1/1", successes, conflicts)
        }

        var purchases, parties, history, participants, reservations, outbox, replays, usedUnits int
        err := database.QueryRowContext(ctx, `
            SELECT
                (SELECT count(*) FROM purchases WHERE offering_id = $1),
                (SELECT count(*) FROM parties WHERE email IN ($2, $3)),
                (SELECT count(*) FROM purchase_status_history WHERE purchase_id IN (SELECT id FROM purchases WHERE offering_id = $1)),
                (SELECT count(*) FROM purchase_participants WHERE purchase_id IN (SELECT id FROM purchases WHERE offering_id = $1)),
                (SELECT count(*) FROM quota_reservations WHERE offering_id = $1),
                (SELECT count(*) FROM outbox_events WHERE aggregate_id IN (SELECT id FROM purchases WHERE offering_id = $1)),
                (SELECT count(*) FROM idempotency_records WHERE namespace = 'storefront.purchase.create.guest' AND idempotency_key IN ($4, $5)),
                (SELECT COALESCE(sum(participant_units), 0) FROM quota_reservations WHERE offering_id = $1 AND status IN ('RESERVED', 'CONSUMED'))
        `, offeringID, emailA, emailB, keyA, keyB).Scan(&purchases, &parties, &history, &participants, &reservations, &outbox, &replays, &usedUnits)
        if err != nil {
            t.Fatal(err)
        }
        if purchases != 1 || parties != 1 || history != 1 || participants != 1 || reservations != 1 || outbox != 2 || replays != 1 || usedUnits != 1 {
            t.Fatalf("quota race effects = purchases:%d parties:%d history:%d participants:%d reservations:%d outbox:%d replays:%d units:%d", purchases, parties, history, participants, reservations, outbox, replays, usedUnits)
        }
    }
}

func cleanupPublicCheckoutFixture(t *testing.T, ctx context.Context, database *sql.DB, prefix, eventID string) {
    t.Helper()
    statements := []struct {
        query string
        args  []any
    }{
        {`DELETE FROM idempotency_records WHERE idempotency_key LIKE $1`, []any{prefix + "%"}},
        {`DELETE FROM outbox_events WHERE aggregate_id IN (SELECT id FROM purchases WHERE event_id = $1)`, []any{eventID}},
        {`DELETE FROM quota_reservations WHERE event_id = $1`, []any{eventID}},
        {`DELETE FROM purchase_status_history WHERE purchase_id IN (SELECT id FROM purchases WHERE event_id = $1)`, []any{eventID}},
        {`DELETE FROM purchase_participants WHERE event_id = $1`, []any{eventID}},
        {`DELETE FROM purchases WHERE event_id = $1`, []any{eventID}},
        {`DELETE FROM parties WHERE email LIKE $1`, []any{prefix + "%"}},
        {`DELETE FROM offerings WHERE event_id = $1`, []any{eventID}},
        {`DELETE FROM qurban_events WHERE id = $1`, []any{eventID}},
    }
    for _, statement := range statements {
        if _, err := database.ExecContext(ctx, statement.query, statement.args...); err != nil {
            t.Errorf("cleanup public checkout fixture: %v", err)
        }
    }
}
