package payment

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "sync"
    "testing"
    "time"

    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

type memoryEvidenceStore struct {
    mutex   sync.Mutex
    objects map[string][]byte
    puts    int
    deletes int
    putErr  error
}

func (store *memoryEvidenceStore) Put(_ context.Context, object EvidenceObject) (string, error) {
    store.mutex.Lock()
    defer store.mutex.Unlock()
    if store.putErr != nil {
        return "", store.putErr
    }
    body, err := io.ReadAll(object.Body)
    if err != nil {
        return "", err
    }
    digest := sha256.Sum256(body)
    if int64(len(body)) != object.SizeBytes || !bytes.Equal(digest[:], object.SHA256) {
        return "", fmt.Errorf("invalid evidence object")
    }
    store.puts++
    reference := fmt.Sprintf("evidence/%d", store.puts)
    store.objects[reference] = append([]byte(nil), body...)
    return reference, nil
}

func (store *memoryEvidenceStore) Delete(_ context.Context, reference string) error {
    store.mutex.Lock()
    defer store.mutex.Unlock()
    store.deletes++
    delete(store.objects, reference)
    return nil
}

func TestPublicEvidenceSubmissionPostgreSQLReplayAndDuplicateGuard(t *testing.T) {
    gin.SetMode(gin.TestMode)
    database := paymentIntegrationDatabase(t)
    ctx := context.Background()
    acquirePaymentIntegrationLock(t, ctx, database)
    fixture := createPaymentFixture(t, ctx, database, "payment-submit", time.Now().UTC().Truncate(time.Microsecond))
    t.Cleanup(func() { cleanupPaymentFixture(t, ctx, database, fixture) })

    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "payment-test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    store := &memoryEvidenceStore{objects: map[string][]byte{}}
    handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)), store)
    handler.now = func() time.Time { return fixture.now }
    handler.random = bytes.NewReader(append(make([]byte, referenceEntropyBytes), bytes.Repeat([]byte{1}, referenceEntropyBytes)...))
    router := gin.New()
    handler.RegisterRoutes(router.Group(""))

    evidence := []byte{0xff, 0xd8, 0xff, 0x00, 0x01}
    first := submitEvidence(t, router, fixture, "payment-intent-0001", fixture.token, evidence)
    if first.Code != http.StatusCreated {
        t.Fatalf("first status/body = %d/%s", first.Code, first.Body.String())
    }
    replay := submitEvidence(t, router, fixture, "payment-intent-0001", fixture.token, evidence)
    if replay.Code != http.StatusCreated || replay.Body.String() != first.Body.String() {
        t.Fatalf("replay status/body = %d/%s, first %s", replay.Code, replay.Body.String(), first.Body.String())
    }
    changed := submitEvidence(t, router, fixture, "payment-intent-0001", fixture.token, append(evidence, 0x02))
    if changed.Code != http.StatusConflict {
        t.Fatalf("changed intent status/body = %d/%s", changed.Code, changed.Body.String())
    }
    wrongToken := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))
    forbidden := submitEvidence(t, router, fixture, "payment-intent-0002", wrongToken, evidence)
    if forbidden.Code != http.StatusForbidden {
        t.Fatalf("wrong token status/body = %d/%s", forbidden.Code, forbidden.Body.String())
    }
    parallel := submitEvidence(t, router, fixture, "payment-intent-0002", fixture.token, evidence)
    if parallel.Code != http.StatusConflict {
        t.Fatalf("parallel submitted status/body = %d/%s", parallel.Code, parallel.Body.String())
    }
    if store.puts != 1 || store.deletes != 0 || len(store.objects) != 1 {
        t.Fatalf("store effects after replay/conflicts = puts:%d deletes:%d objects:%d", store.puts, store.deletes, len(store.objects))
    }

    var response struct {
        Data struct {
            ID           string `json:"id"`
            PaymentRef   string `json:"payment_ref"`
            AmountMinor  int64  `json:"amount_minor"`
            CurrencyCode string `json:"currency_code"`
            Status       string `json:"status"`
        } `json:"data"`
    }
    if err := json.Unmarshal(first.Body.Bytes(), &response); err != nil {
        t.Fatal(err)
    }
    if response.Data.ID == "" || !strings.HasPrefix(response.Data.PaymentRef, "PAY-") || response.Data.AmountMinor != 100 || response.Data.CurrencyCode != "IDR" || response.Data.Status != "SUBMITTED" || strings.Contains(first.Body.String(), "evidence/") || strings.Contains(first.Body.String(), "proof.jpg") {
        t.Fatalf("public response = %s", first.Body.String())
    }

    namespace := "storefront.purchase.payment-evidence.submit." + fixture.purchaseID
    var payments, history, outboxRows, replayRows int
    var expiresAt sql.NullTime
    var evidenceReference, filename, mediaType string
    var size int64
    var digest []byte
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID).Scan(&payments); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_status_history WHERE payment_id = $1`, response.Data.ID).Scan(&history); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT evidence_reference, evidence_filename, evidence_media_type, evidence_size_bytes, evidence_sha256 FROM payment_records WHERE id = $1`, response.Data.ID).Scan(&evidenceReference, &filename, &mediaType, &size, &digest); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT expires_at FROM quota_reservations WHERE purchase_id = $1 AND status = 'RESERVED'`, fixture.purchaseID).Scan(&expiresAt); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id = $1 AND event_type = 'PaymentEvidenceSubmitted'`, response.Data.ID).Scan(&outboxRows); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2 AND expires_at IS NULL`, namespace, "payment-intent-0001").Scan(&replayRows); err != nil {
        t.Fatal(err)
    }
    expectedDigest := sha256.Sum256(evidence)
    if payments != 1 || history != 1 || outboxRows != 1 || replayRows != 1 || expiresAt.Valid || evidenceReference != "evidence/1" || filename != "proof.jpg" || mediaType != "image/jpeg" || size != int64(len(evidence)) || !bytes.Equal(digest, expectedDigest[:]) {
        t.Fatalf("stored effects = payments:%d history:%d outbox:%d replay:%d expiry:%v reference:%q filename:%q media:%q size:%d digest:%x", payments, history, outboxRows, replayRows, expiresAt, evidenceReference, filename, mediaType, size, digest)
    }
    var payload string
    if err := database.QueryRowContext(ctx, `SELECT payload::text FROM outbox_events WHERE aggregate_id = $1 AND event_type = 'PaymentEvidenceSubmitted'`, response.Data.ID).Scan(&payload); err != nil {
        t.Fatal(err)
    }
    if strings.Contains(payload, "evidence") || strings.Contains(payload, "proof.jpg") || strings.Contains(payload, fixture.token) {
        t.Fatalf("unsafe outbox payload = %s", payload)
    }

    if _, err := database.ExecContext(ctx, `UPDATE quota_reservations SET expires_at = $1, version = version + 1 WHERE purchase_id = $2 AND status = 'RESERVED'`, fixture.now.Add(time.Hour), fixture.purchaseID); err != nil {
        t.Fatal(err)
    }
    guarded := submitEvidence(t, router, fixture, "payment-intent-0003", fixture.token, evidence)
    if guarded.Code != http.StatusConflict {
        t.Fatalf("database guard status/body = %d/%s", guarded.Code, guarded.Body.String())
    }
    if store.puts != 2 || store.deletes != 1 || len(store.objects) != 1 {
        t.Fatalf("database guard cleanup = puts:%d deletes:%d objects:%d", store.puts, store.deletes, len(store.objects))
    }
}

func TestPublicEvidenceSubmissionPostgreSQLReacquiresAfterRejection(t *testing.T) {
    gin.SetMode(gin.TestMode)
    database := paymentIntegrationDatabase(t)
    ctx := context.Background()
    acquirePaymentIntegrationLock(t, ctx, database)
    fixture := createPaymentFixture(t, ctx, database, "payment-reacquire", time.Now().UTC().Truncate(time.Microsecond))
    t.Cleanup(func() { cleanupPaymentFixture(t, ctx, database, fixture) })
    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "payment-test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    store := &memoryEvidenceStore{objects: map[string][]byte{}}
    handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)), store)
    handler.now = func() time.Time { return fixture.now }
    handler.random = bytes.NewReader(append(make([]byte, referenceEntropyBytes), bytes.Repeat([]byte{1}, referenceEntropyBytes)...))
    router := gin.New()
    handler.RegisterRoutes(router.Group(""))

    first := submitEvidence(t, router, fixture, "payment-reacquire-0001", fixture.token, []byte("%PDF-1.4\nfirst"))
    if first.Code != http.StatusCreated {
        t.Fatalf("first status/body = %d/%s", first.Code, first.Body.String())
    }
    var paymentID string
    if err := database.QueryRowContext(ctx, `SELECT id FROM payment_records WHERE purchase_id = $1 AND status = 'SUBMITTED'`, fixture.purchaseID).Scan(&paymentID); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `UPDATE payment_records SET status = 'REJECTED', rejection_reason = 'invalid proof' WHERE id = $1`, paymentID); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `UPDATE quota_reservations SET status = 'RELEASED', released_at = now(), release_reason = 'PAYMENT_REJECTED', version = version + 1 WHERE purchase_id = $1 AND status = 'RESERVED'`, fixture.purchaseID); err != nil {
        t.Fatal(err)
    }

    second := submitEvidence(t, router, fixture, "payment-reacquire-0002", fixture.token, []byte("%PDF-1.4\nsecond"))
    if second.Code != http.StatusCreated {
        t.Fatalf("second status/body = %d/%s", second.Code, second.Body.String())
    }
    var payments, attempts, submitted int
    var currentExpiry sql.NullTime
    if err := database.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE status = 'SUBMITTED') FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID).Scan(&payments, &submitted); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM quota_reservations WHERE purchase_id = $1`, fixture.purchaseID).Scan(&attempts); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT expires_at FROM quota_reservations WHERE purchase_id = $1 AND status = 'RESERVED'`, fixture.purchaseID).Scan(&currentExpiry); err != nil {
        t.Fatal(err)
    }
    if payments != 2 || submitted != 1 || attempts != 2 || currentExpiry.Valid || store.puts != 2 || store.deletes != 0 {
        t.Fatalf("reacquire effects = payments:%d submitted:%d attempts:%d expiry:%v puts:%d deletes:%d", payments, submitted, attempts, currentExpiry, store.puts, store.deletes)
    }
}

func TestPublicEvidenceSubmissionPostgreSQLRollsBackStorageFailure(t *testing.T) {
    gin.SetMode(gin.TestMode)
    database := paymentIntegrationDatabase(t)
    ctx := context.Background()
    acquirePaymentIntegrationLock(t, ctx, database)
    fixture := createPaymentFixture(t, ctx, database, "payment-storage-failure", time.Now().UTC().Truncate(time.Microsecond))
    t.Cleanup(func() { cleanupPaymentFixture(t, ctx, database, fixture) })
    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "payment-test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    store := &memoryEvidenceStore{objects: map[string][]byte{}, putErr: ErrStorageUnavailable}
    handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)), store)
    handler.now = func() time.Time { return fixture.now }
    router := gin.New()
    handler.RegisterRoutes(router.Group(""))

    response := submitEvidence(t, router, fixture, "payment-storage-0001", fixture.token, []byte("%PDF-1.4\nproof"))
    if response.Code != http.StatusServiceUnavailable {
        t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
    }
    var payments, histories, outboxRows, replayRows int
    var expiresAt sql.NullTime
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID).Scan(&payments); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_status_history WHERE payment_id IN (SELECT id FROM payment_records WHERE purchase_id = $1)`, fixture.purchaseID).Scan(&histories); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id IN (SELECT id FROM payment_records WHERE purchase_id = $1)`, fixture.purchaseID).Scan(&outboxRows); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM idempotency_records WHERE namespace = $1`, "storefront.purchase.payment-evidence.submit."+fixture.purchaseID).Scan(&replayRows); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT expires_at FROM quota_reservations WHERE purchase_id = $1 AND status = 'RESERVED'`, fixture.purchaseID).Scan(&expiresAt); err != nil {
        t.Fatal(err)
    }
    if payments != 0 || histories != 0 || outboxRows != 0 || replayRows != 0 || !expiresAt.Valid {
        t.Fatalf("rollback effects = payments:%d histories:%d outbox:%d replay:%d expiry:%v", payments, histories, outboxRows, replayRows, expiresAt)
    }
}

func TestPublicEvidenceSubmissionPostgreSQLExpiredReservationHonorsQuota(t *testing.T) {
    gin.SetMode(gin.TestMode)
    database := paymentIntegrationDatabase(t)
    ctx := context.Background()
    acquirePaymentIntegrationLock(t, ctx, database)
    fixture := createPaymentFixture(t, ctx, database, "payment-expired-quota", time.Now().UTC().Truncate(time.Microsecond))
    t.Cleanup(func() { cleanupPaymentFixture(t, ctx, database, fixture) })
    if _, err := database.ExecContext(ctx, `UPDATE quota_reservations SET expires_at = $1 WHERE purchase_id = $2`, fixture.now.Add(-time.Minute), fixture.purchaseID); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `UPDATE qurban_events SET participant_quota = 0 WHERE id = $1`, fixture.eventID); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `UPDATE offerings SET participant_quota = 0 WHERE id = $1`, fixture.offeringID); err != nil {
        t.Fatal(err)
    }
    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "payment-test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    store := &memoryEvidenceStore{objects: map[string][]byte{}}
    handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)), store)
    handler.now = func() time.Time { return fixture.now }
    router := gin.New()
    handler.RegisterRoutes(router.Group(""))

    response := submitEvidence(t, router, fixture, "payment-expired-0001", fixture.token, []byte("%PDF-1.4\nproof"))
    if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "quota_unavailable") {
        t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
    }
    var payments, attempts int
    var status string
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID).Scan(&payments); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*), min(status) FROM quota_reservations WHERE purchase_id = $1`, fixture.purchaseID).Scan(&attempts, &status); err != nil {
        t.Fatal(err)
    }
    if payments != 0 || attempts != 1 || status != "RESERVED" || store.puts != 0 {
        t.Fatalf("quota rollback = payments:%d attempts:%d status:%s puts:%d", payments, attempts, status, store.puts)
    }
}

func TestPublicEvidenceSubmissionPostgreSQLRejectsAmountCurrencyAndZeroTotal(t *testing.T) {
    gin.SetMode(gin.TestMode)
    database := paymentIntegrationDatabase(t)
    ctx := context.Background()
    acquirePaymentIntegrationLock(t, ctx, database)
    fixture := createPaymentFixture(t, ctx, database, "payment-values", time.Now().UTC().Truncate(time.Microsecond))
    t.Cleanup(func() { cleanupPaymentFixture(t, ctx, database, fixture) })
    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "payment-test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    store := &memoryEvidenceStore{objects: map[string][]byte{}}
    handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)), store)
    handler.now = func() time.Time { return fixture.now }
    router := gin.New()
    handler.RegisterRoutes(router.Group(""))
    evidence := []byte("%PDF-1.4\nproof")

    wrongAmount := submitEvidenceValues(t, router, fixture, "payment-values-0001", fixture.token, "99", "IDR", evidence)
    wrongCurrency := submitEvidenceValues(t, router, fixture, "payment-values-0002", fixture.token, "100", "USD", evidence)
    if wrongAmount.Code != http.StatusUnprocessableEntity || wrongCurrency.Code != http.StatusUnprocessableEntity {
        t.Fatalf("amount/currency statuses = %d/%d", wrongAmount.Code, wrongCurrency.Code)
    }
    if _, err := database.ExecContext(ctx, `UPDATE purchases SET offering_unit_price_minor = 0, total_amount_minor = 0 WHERE id = $1`, fixture.purchaseID); err != nil {
        t.Fatal(err)
    }
    zeroTotal := submitEvidenceValues(t, router, fixture, "payment-values-0003", fixture.token, "1", "IDR", evidence)
    if zeroTotal.Code != http.StatusConflict || !strings.Contains(zeroTotal.Body.String(), "state_conflict") {
        t.Fatalf("zero-total status/body = %d/%s", zeroTotal.Code, zeroTotal.Body.String())
    }
    var payments, replays int
    var expiresAt sql.NullTime
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID).Scan(&payments); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM idempotency_records WHERE namespace = $1`, "storefront.purchase.payment-evidence.submit."+fixture.purchaseID).Scan(&replays); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `SELECT expires_at FROM quota_reservations WHERE purchase_id = $1 AND status = 'RESERVED'`, fixture.purchaseID).Scan(&expiresAt); err != nil {
        t.Fatal(err)
    }
    if payments != 0 || replays != 0 || !expiresAt.Valid || store.puts != 0 {
        t.Fatalf("validation effects = payments:%d replays:%d expiry:%v puts:%d", payments, replays, expiresAt, store.puts)
    }
}

func TestPublicEvidenceSubmissionPostgreSQLConcurrentRetries(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := []struct {
        name      string
        keys      [2]string
        wantCodes map[int]int
    }{
        {"same intent", [2]string{"payment-concurrent-0001", "payment-concurrent-0001"}, map[int]int{http.StatusCreated: 2}},
        {"different intents", [2]string{"payment-concurrent-0002", "payment-concurrent-0003"}, map[int]int{http.StatusCreated: 1, http.StatusConflict: 1}},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            database := paymentIntegrationDatabase(t)
            ctx := context.Background()
            acquirePaymentIntegrationLock(t, ctx, database)
            fixture := createPaymentFixture(t, ctx, database, "payment-"+strings.ReplaceAll(test.name, " ", "-"), time.Now().UTC().Truncate(time.Microsecond))
            t.Cleanup(func() { cleanupPaymentFixture(t, ctx, database, fixture) })
            cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "payment-test", Value: make([]byte, 32)}})
            if err != nil {
                t.Fatal(err)
            }
            store := &memoryEvidenceStore{objects: map[string][]byte{}}
            handler := NewPublicHandler(database, cipher, slog.New(slog.NewTextHandler(io.Discard, nil)), store)
            handler.now = func() time.Time { return fixture.now }
            handler.random = bytes.NewReader(bytes.Repeat([]byte{3}, 2*referenceEntropyBytes))
            router := gin.New()
            handler.RegisterRoutes(router.Group(""))

            start := make(chan struct{})
            responses := make(chan *httptest.ResponseRecorder, 2)
            for _, key := range test.keys {
                key := key
                go func() {
                    <-start
                    responses <- submitEvidence(t, router, fixture, key, fixture.token, []byte("%PDF-1.4\nproof"))
                }()
            }
            close(start)
            codes := map[int]int{}
            bodies := make([]string, 0, 2)
            for range 2 {
                response := <-responses
                codes[response.Code]++
                bodies = append(bodies, response.Body.String())
            }
            for code, count := range test.wantCodes {
                if codes[code] != count {
                    t.Fatalf("codes = %#v, want %d x %d", codes, count, code)
                }
            }
            if test.keys[0] == test.keys[1] && bodies[0] != bodies[1] {
                t.Fatalf("same-intent bodies differ: %q / %q", bodies[0], bodies[1])
            }
            var payments, histories, outboxRows int
            if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID).Scan(&payments); err != nil {
                t.Fatal(err)
            }
            if err := database.QueryRowContext(ctx, `SELECT count(*) FROM payment_status_history WHERE payment_id IN (SELECT id FROM payment_records WHERE purchase_id = $1)`, fixture.purchaseID).Scan(&histories); err != nil {
                t.Fatal(err)
            }
            if err := database.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id IN (SELECT id FROM payment_records WHERE purchase_id = $1)`, fixture.purchaseID).Scan(&outboxRows); err != nil {
                t.Fatal(err)
            }
            if payments != 1 || histories != 1 || outboxRows != 1 || store.puts != 1 || store.deletes != 0 {
                t.Fatalf("concurrent effects = payments:%d histories:%d outbox:%d puts:%d deletes:%d", payments, histories, outboxRows, store.puts, store.deletes)
            }
        })
    }
}

type paymentFixture struct {
    now        time.Time
    eventID    string
    offeringID string
    partyID    string
    purchaseID string
    token      string
}

func createPaymentFixture(t *testing.T, ctx context.Context, database *sql.DB, prefix string, now time.Time) paymentFixture {
    t.Helper()
    var value paymentFixture
    value.now = now
    var year int
    if err := database.QueryRowContext(ctx, `SELECT candidate FROM generate_series(8000, 8999) AS candidate WHERE NOT EXISTS (SELECT 1 FROM qurban_events WHERE event_year = candidate) ORDER BY candidate LIMIT 1`).Scan(&year); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status, participant_quota) VALUES ($1, $2, 'ACTIVE', 10) RETURNING id`, year, prefix).Scan(&value.eventID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, participant_quota, status) VALUES ($1, $2, 'One share', 'SHARE', 100, 'IDR', 1, 10, 'PUBLISHED') RETURNING id`, value.eventID, prefix).Scan(&value.offeringID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO parties (party_type, display_name) VALUES ('PERSON', $1) RETURNING id`, prefix).Scan(&value.partyID); err != nil {
        t.Fatal(err)
    }
    value.token = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
    tokenHash := sha256.Sum256([]byte(value.token))
    if err := database.QueryRowContext(ctx, `
INSERT INTO purchases (
    event_id, purchase_ref, channel, purchaser_party_id, payer_party_id, offering_id, participant_count,
    offering_name_snapshot, offering_kind_snapshot, offering_unit_price_minor, participant_capacity_snapshot,
    total_amount_minor, currency_code, status, access_token_hash
) VALUES ($1, $2, 'COMMON', $3, $3, $4, 1, 'One share', 'SHARE', 100, 1, 100, 'IDR', 'PENDING_PAYMENT', $5)
RETURNING id`, value.eventID, prefix, value.partyID, value.offeringID, tokenHash[:]).Scan(&value.purchaseID); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `INSERT INTO purchase_participants (event_id, purchase_id, party_id, sequence_no, display_name_snapshot) VALUES ($1, $2, $3, 1, $4)`, value.eventID, value.purchaseID, value.partyID, prefix); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `INSERT INTO purchase_status_history (purchase_id, from_status, to_status) VALUES ($1, 'DRAFT', 'PENDING_PAYMENT')`, value.purchaseID); err != nil {
        t.Fatal(err)
    }
    if _, err := database.ExecContext(ctx, `INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at) VALUES ($1, $2, $3, 1, 1, 'RESERVED', $4)`, value.eventID, value.purchaseID, value.offeringID, now.Add(time.Hour)); err != nil {
        t.Fatal(err)
    }
    value.now = time.Now().UTC().Truncate(time.Microsecond)
    return value
}

func submitEvidence(t *testing.T, router http.Handler, fixture paymentFixture, key, token string, data []byte) *httptest.ResponseRecorder {
    t.Helper()
    return submitEvidenceValues(t, router, fixture, key, token, "100", "IDR", data)
}

func submitEvidenceValues(t *testing.T, router http.Handler, fixture paymentFixture, key, token, amount, currency string, data []byte) *httptest.ResponseRecorder {
    t.Helper()
    body, contentType := evidenceBody(t, amount, currency, "proof.jpg", evidenceMediaType(data), data)
    request := httptest.NewRequest(http.MethodPost, "/purchases/"+fixture.purchaseID+"/payment-evidence", body)
    request.Header.Set("Content-Type", contentType)
    request.Header.Set("Authorization", "Bearer "+token)
    request.Header.Set("Idempotency-Key", key)
    response := httptest.NewRecorder()
    router.ServeHTTP(response, request)
    return response
}

func evidenceMediaType(data []byte) string {
    if bytes.HasPrefix(data, []byte("%PDF-")) {
        return "application/pdf"
    }
    return "image/jpeg"
}

func paymentIntegrationDatabase(t *testing.T) *sql.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is required for PostgreSQL payment coverage")
    }
    database, err := platformdb.Open(dsn)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    return database
}

func acquirePaymentIntegrationLock(t *testing.T, ctx context.Context, database *sql.DB) {
    t.Helper()
    connection, err := database.Conn(ctx)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := connection.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, int64(20260813)); err != nil {
        _ = connection.Close()
        t.Fatal(err)
    }
    t.Cleanup(func() {
        _, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(20260813))
        _ = connection.Close()
    })
}

func cleanupPaymentFixture(t *testing.T, ctx context.Context, database *sql.DB, fixture paymentFixture) {
    t.Helper()
    _, _ = database.ExecContext(ctx, `DELETE FROM payment_status_history WHERE payment_id IN (SELECT id FROM payment_records WHERE purchase_id = $1)`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM outbox_events WHERE aggregate_id IN (SELECT id FROM payment_records WHERE purchase_id = $1)`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM payment_records WHERE purchase_id = $1`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM idempotency_records WHERE namespace = $1`, "storefront.purchase.payment-evidence.submit."+fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM quota_reservations WHERE purchase_id = $1`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM purchase_status_history WHERE purchase_id = $1`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM purchase_participants WHERE purchase_id = $1`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM purchases WHERE id = $1`, fixture.purchaseID)
    _, _ = database.ExecContext(ctx, `DELETE FROM offerings WHERE id = $1`, fixture.offeringID)
    _, _ = database.ExecContext(ctx, `DELETE FROM qurban_events WHERE id = $1`, fixture.eventID)
    _, _ = database.ExecContext(ctx, `DELETE FROM parties WHERE id = $1`, fixture.partyID)
}
