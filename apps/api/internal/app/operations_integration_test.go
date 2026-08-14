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
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Kangditya/persona-apps/apps/api/internal/config"
	"github.com/Kangditya/persona-apps/apps/api/internal/event"
	"github.com/Kangditya/persona-apps/apps/api/internal/offering"
	platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
	"github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
	"github.com/gin-gonic/gin"
)

func TestOperationsCommandsPostgreSQL(t *testing.T) {
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

	prefix := fmt.Sprintf("w2ops%d", time.Now().UnixNano())
	token := strings.Repeat("s", 43)
	csrf := strings.Repeat("c", 43)
	operatorID := insertOperationsSession(t, database, prefix, token, csrf)
	eventYear := availableEventYear(t, database)
	eventName := "Operations " + prefix
	var eventID, offeringID string
	t.Cleanup(func() { cleanupOperationsFixture(t, database, prefix, operatorID, eventID, offeringID) })

	cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "integration", Value: make([]byte, 32)}})
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	authentication := newTestAuthentication(t, database)
	server, err := NewServer(":0", database, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20}, authentication, event.NewOperationsHandler(database, cipher, logger), offering.NewOperationsHandler(database, cipher, logger))
	if err != nil {
		t.Fatal(err)
	}

	createEventBody := fmt.Sprintf(`{"event_year":%d,"name":%q,"participant_quota":10}`, eventYear, eventName)
	createEventKey := prefix + ".event.create"
	createdEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events", createEventKey, createEventBody, http.StatusCreated)
	eventID = responseID(t, createdEvent)
	if responseVersion(t, createdEvent) != 1 || responseStatus(t, createdEvent) != "DRAFT" {
		t.Fatalf("created event = %s", createdEvent)
	}
	replay := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events", createEventKey, createEventBody, http.StatusCreated)
	if !bytes.Equal(createdEvent, replay) {
		t.Fatalf("create replay differs: first=%s replay=%s", createdEvent, replay)
	}
	operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events", createEventKey, fmt.Sprintf(`{"event_year":%d,"name":"different"}`, eventYear), http.StatusConflict)

	noOp := operationsRequest(t, server, token, csrf, http.MethodPatch, "/api/operations/v1/events/"+eventID, "", fmt.Sprintf(`{"expected_version":1,"name":"  %s  "}`, eventName), http.StatusOK)
	if responseVersion(t, noOp) != 1 {
		t.Fatalf("no-op patch changed version: %s", noOp)
	}
	patchedEvent := operationsRequest(t, server, token, csrf, http.MethodPatch, "/api/operations/v1/events/"+eventID, "", fmt.Sprintf(`{"expected_version":1,"name":%q,"participant_quota":12}`, eventName+" Updated"), http.StatusOK)
	if responseVersion(t, patchedEvent) != 2 {
		t.Fatalf("patched event = %s", patchedEvent)
	}
	publishEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/publish", prefix+".event.publish", `{"expected_version":2}`, http.StatusOK)
	if responseVersion(t, publishEvent) != 3 || responseStatus(t, publishEvent) != "PUBLISHED" {
		t.Fatalf("published event = %s", publishEvent)
	}

	createOffering := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/offerings", prefix+".offering.create", `{"code":"COW-7","name":"Cow share","offering_kind":"SHARE","price_minor":2500000,"currency_code":"IDR","participant_capacity":7,"participant_quota":7}`, http.StatusCreated)
	offeringID = responseID(t, createOffering)
	if responseVersion(t, createOffering) != 1 || responseStatus(t, createOffering) != "DRAFT" {
		t.Fatalf("created offering = %s", createOffering)
	}
	patchedOffering := operationsRequest(t, server, token, csrf, http.MethodPatch, "/api/operations/v1/offerings/"+offeringID, "", `{"expected_version":1,"name":"Cow share updated","price_minor":2750000}`, http.StatusOK)
	if responseVersion(t, patchedOffering) != 2 {
		t.Fatalf("patched offering = %s", patchedOffering)
	}
	publishOffering := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/offerings/"+offeringID+"/publish", prefix+".offering.publish", `{"expected_version":2}`, http.StatusOK)
	if responseVersion(t, publishOffering) != 3 || responseStatus(t, publishOffering) != "PUBLISHED" {
		t.Fatalf("published offering = %s", publishOffering)
	}
	activateEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/activate", prefix+".event.activate", `{"expected_version":3}`, http.StatusOK)
	if responseVersion(t, activateEvent) != 4 || responseStatus(t, activateEvent) != "ACTIVE" {
		t.Fatalf("activated event = %s", activateEvent)
	}

	activePublic := catalogueRequest(t, server, "/api/public/v1/events/active", http.StatusOK)
	assertPublicEvent(t, activePublic, eventID)
	publicList := catalogueRequest(t, server, "/api/public/v1/events/"+eventID+"/offerings", http.StatusOK)
	assertPublicOfferingList(t, publicList, eventID, offeringID)
	publicDetail := catalogueRequest(t, server, "/api/public/v1/offerings/"+offeringID, http.StatusOK)
	assertPublicOffering(t, publicDetail, eventID, offeringID)

	unavailableOffering := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/offerings/"+offeringID+"/unavailable", prefix+".offering.unavailable", `{"expected_version":3}`, http.StatusOK)
	if responseVersion(t, unavailableOffering) != 4 || responseStatus(t, unavailableOffering) != "UNAVAILABLE" {
		t.Fatalf("unavailable offering = %s", unavailableOffering)
	}
	catalogueRequest(t, server, "/api/public/v1/offerings/"+offeringID, http.StatusNotFound)
	republishedOffering := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/offerings/"+offeringID+"/publish", prefix+".offering.republish", `{"expected_version":4}`, http.StatusOK)
	if responseVersion(t, republishedOffering) != 5 || responseStatus(t, republishedOffering) != "PUBLISHED" {
		t.Fatalf("republished offering = %s", republishedOffering)
	}

	suspendedEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/suspend", prefix+".event.suspend", `{"expected_version":4}`, http.StatusOK)
	if responseVersion(t, suspendedEvent) != 5 || responseStatus(t, suspendedEvent) != "SUSPENDED" {
		t.Fatalf("suspended event = %s", suspendedEvent)
	}
	assertNoActiveEvent(t, catalogueRequest(t, server, "/api/public/v1/events/active", http.StatusOK))
	catalogueRequest(t, server, "/api/public/v1/offerings/"+offeringID, http.StatusNotFound)
	reactivatedEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/activate", prefix+".event.reactivate", `{"expected_version":5}`, http.StatusOK)
	if responseVersion(t, reactivatedEvent) != 6 || responseStatus(t, reactivatedEvent) != "ACTIVE" {
		t.Fatalf("reactivated event = %s", reactivatedEvent)
	}
	closedEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/close", prefix+".event.close", `{"expected_version":6}`, http.StatusOK)
	if responseVersion(t, closedEvent) != 7 || responseStatus(t, closedEvent) != "CLOSED" {
		t.Fatalf("closed event = %s", closedEvent)
	}
	archivedOffering := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/offerings/"+offeringID+"/archive", prefix+".offering.archive", `{"expected_version":5}`, http.StatusOK)
	if responseVersion(t, archivedOffering) != 6 || responseStatus(t, archivedOffering) != "ARCHIVED" {
		t.Fatalf("archived offering = %s", archivedOffering)
	}
	archivedEvent := operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events/"+eventID+"/archive", prefix+".event.archive", `{"expected_version":7}`, http.StatusOK)
	if responseVersion(t, archivedEvent) != 8 || responseStatus(t, archivedEvent) != "ARCHIVED" {
		t.Fatalf("archived event = %s", archivedEvent)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/operations/v1/events", strings.NewReader(createEventBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("X-CSRF-Token", csrf)
	request.Header.Set("Idempotency-Key", createEventKey)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated replay status = %d, body = %s", response.Code, response.Body.String())
	}

	deniedKey := prefix + ".event.denied"
	if _, err := database.Exec(`UPDATE operator_sessions SET permission_snapshot = '["event.read"]'::jsonb WHERE operator_user_id = $1`, operatorID); err != nil {
		t.Fatal(err)
	}
	operationsRequest(t, server, token, csrf, http.MethodPost, "/api/operations/v1/events", deniedKey, fmt.Sprintf(`{"event_year":%d,"name":"denied"}`, eventYear+1), http.StatusForbidden)
	if _, err := database.Exec(`UPDATE operator_sessions SET permission_snapshot = '["event.read","event.manage","offering.read","offering.manage"]'::jsonb WHERE operator_user_id = $1`, operatorID); err != nil {
		t.Fatal(err)
	}

	keys := []string{
		createEventKey,
		prefix + ".event.publish",
		prefix + ".offering.create",
		prefix + ".offering.publish",
		prefix + ".event.activate",
		prefix + ".offering.unavailable",
		prefix + ".offering.republish",
		prefix + ".event.suspend",
		prefix + ".event.reactivate",
		prefix + ".event.close",
		prefix + ".offering.archive",
		prefix + ".event.archive",
		deniedKey,
	}
	assertOperationsEffects(t, database, prefix, operatorID, eventID, offeringID, deniedKey, keys, token, csrf)
}

func operationsRequest(t *testing.T, server *http.Server, token, csrf, method, path, key, body string, wantStatus int) []byte {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("X-CSRF-Token", csrf)
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	request.AddCookie(&http.Cookie{Name: "__Host-operations_session", Value: token})
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s status = %d, want %d, body = %s", method, path, response.Code, wantStatus, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("%s %s missing private response headers", method, path)
	}
	return response.Body.Bytes()
}

func catalogueRequest(t *testing.T, server *http.Server, path string, wantStatus int) []byte {
	t.Helper()
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != wantStatus {
		t.Fatalf("GET %s status = %d, want %d, body = %s", path, response.Code, wantStatus, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("GET %s missing public response headers", path)
	}
	return response.Body.Bytes()
}

func assertPublicEvent(t *testing.T, body []byte, eventID string) {
	t.Helper()
	data := responseData(t, body)
	if data["id"] != eventID || data["status"] != "ACTIVE" {
		t.Fatalf("public Event = %s", body)
	}
	assertJSONKeys(t, data, "id", "event_year", "name", "status", "registration_opens_at", "registration_closes_at")
}

func assertNoActiveEvent(t *testing.T, body []byte) {
	t.Helper()
	var response struct {
		Data any `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data != nil {
		t.Fatalf("active Event response = %s, want null", body)
	}
}

func assertPublicOfferingList(t *testing.T, body []byte, eventID, offeringID string) {
	t.Helper()
	var response struct {
		Data []map[string]any `json:"data"`
		Page map[string]any   `json:"page"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data) != 1 {
		t.Fatalf("public Offering list = %s", body)
	}
	assertPublicOfferingData(t, response.Data[0], eventID, offeringID)
	assertJSONKeys(t, response.Page, "limit", "next_cursor")
}

func assertPublicOffering(t *testing.T, body []byte, eventID, offeringID string) {
	t.Helper()
	assertPublicOfferingData(t, responseData(t, body), eventID, offeringID)
}

func assertPublicOfferingData(t *testing.T, data map[string]any, eventID, offeringID string) {
	t.Helper()
	if data["id"] != offeringID || data["event_id"] != eventID || data["status"] != "PUBLISHED" || data["price_minor"] != float64(2_750_000) {
		t.Fatalf("public Offering = %#v", data)
	}
	assertJSONKeys(t, data, "id", "event_id", "code", "name", "offering_kind", "description", "price_minor", "currency_code", "participant_capacity", "available_participant_units", "status")
}

func assertJSONKeys(t *testing.T, data map[string]any, allowed ...string) {
	t.Helper()
	allowlist := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowlist[key] = struct{}{}
	}
	for key := range data {
		if _, allowed := allowlist[key]; !allowed {
			t.Fatalf("public response exposed unexpected field %q in %#v", key, data)
		}
	}
}

func responseData(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var response struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	return response.Data
}

func responseID(t *testing.T, body []byte) string {
	t.Helper()
	value, ok := responseData(t, body)["id"].(string)
	if !ok || value == "" {
		t.Fatalf("response ID missing: %s", body)
	}
	return value
}

func responseVersion(t *testing.T, body []byte) int64 {
	t.Helper()
	value, ok := responseData(t, body)["version"].(float64)
	if !ok {
		t.Fatalf("response version missing: %s", body)
	}
	return int64(value)
}

func responseStatus(t *testing.T, body []byte) string {
	t.Helper()
	value, _ := responseData(t, body)["status"].(string)
	return value
}

func insertOperationsSession(t *testing.T, database *sql.DB, prefix, token, csrf string) string {
	t.Helper()
	var operatorID string
	if err := database.QueryRow(`INSERT INTO operator_users (external_subject, display_name, status) VALUES ($1, $2, 'ACTIVE') RETURNING id`, prefix, "Integration Operator").Scan(&operatorID); err != nil {
		t.Fatal(err)
	}
	sessionHash, csrfHash := sha256.Sum256([]byte(token)), sha256.Sum256([]byte(csrf))
	if _, err := database.Exec(`INSERT INTO operator_sessions (operator_user_id, session_token_hash, csrf_token_hash, permission_snapshot, expires_at) VALUES ($1, $2, $3, '["event.read","event.manage","offering.read","offering.manage"]'::jsonb, now() + interval '1 hour')`, operatorID, sessionHash[:], csrfHash[:]); err != nil {
		t.Fatal(err)
	}
	return operatorID
}

func availableEventYear(t *testing.T, database *sql.DB) int {
	t.Helper()
	var year int
	if err := database.QueryRow(`SELECT candidate FROM generate_series(8000, 9999) AS candidate WHERE NOT EXISTS (SELECT 1 FROM qurban_events WHERE event_year = candidate) ORDER BY candidate LIMIT 1`).Scan(&year); err != nil {
		t.Fatal(err)
	}
	return year
}

func assertOperationsEffects(t *testing.T, database *sql.DB, prefix, operatorID, eventID, offeringID, deniedKey string, keys []string, token, csrf string) {
	t.Helper()
	var audits, outboxEvents, replayRecords, deniedRecords int
	if err := database.QueryRow(`SELECT count(*) FROM audit_log WHERE actor_operator_id = $1`, operatorID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT count(*) FROM outbox_events WHERE aggregate_id IN ($1, $2)`, eventID, offeringID).Scan(&outboxEvents); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT count(*) FROM idempotency_records WHERE idempotency_key LIKE $1`, prefix+".%").Scan(&replayRecords); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT count(*) FROM idempotency_records WHERE idempotency_key = $1`, deniedKey).Scan(&deniedRecords); err != nil {
		t.Fatal(err)
	}
	if audits != 14 || outboxEvents != 10 || replayRecords != 12 || deniedRecords != 0 {
		t.Fatalf("effects audit/outbox/replay/denied = %d/%d/%d/%d", audits, outboxEvents, replayRecords, deniedRecords)
	}

	rows, err := database.Query(`SELECT action, permission, source, host(client_ip), request_id, coalesce(before_data::text, ''), coalesce(after_data::text, '') FROM audit_log WHERE actor_operator_id = $1`, operatorID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var action, permission, source, clientIP, requestID, before, after string
		if err := rows.Scan(&action, &permission, &source, &clientIP, &requestID, &before, &after); err != nil {
			t.Fatal(err)
		}
		if action == "" || permission == "" || source != "operations-web" || clientIP != "192.0.2.1" || requestID == "" {
			t.Fatalf("unsafe audit metadata = %q/%q/%q/%q/%q", action, permission, source, clientIP, requestID)
		}
		secrets := append(append([]string{}, keys...), token, csrf)
		for _, secret := range secrets {
			if secret != "" && (strings.Contains(before, secret) || strings.Contains(after, secret)) {
				t.Fatalf("audit payload contains request secret")
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	assertAuditHistory(t, database, operatorID)
	assertEffectTypes(t, database, eventID, offeringID)
	assertPayloadsExcludeSecrets(t, database, prefix, eventID, offeringID, append(append([]string{}, keys...), token, csrf))
}

func assertAuditHistory(t *testing.T, database *sql.DB, operatorID string) {
	t.Helper()
	var eventBefore, eventAfter, offeringBefore, offeringAfter int64
	if err := database.QueryRow(`SELECT (before_data->>'participant_quota')::bigint, (after_data->>'participant_quota')::bigint FROM audit_log WHERE actor_operator_id = $1 AND action = 'event.update'`, operatorID).Scan(&eventBefore, &eventAfter); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT (before_data->>'price_minor')::bigint, (after_data->>'price_minor')::bigint FROM audit_log WHERE actor_operator_id = $1 AND action = 'offering.update'`, operatorID).Scan(&offeringBefore, &offeringAfter); err != nil {
		t.Fatal(err)
	}
	if eventBefore != 10 || eventAfter != 12 || offeringBefore != 2_500_000 || offeringAfter != 2_750_000 {
		t.Fatalf("historical Event/Offering audit = %d/%d and %d/%d", eventBefore, eventAfter, offeringBefore, offeringAfter)
	}
}

func assertEffectTypes(t *testing.T, database *sql.DB, eventID, offeringID string) {
	t.Helper()
	expectedAudits := map[string]int{
		"event.create": 1, "event.update": 1, "event.publish": 1, "event.activate": 2,
		"event.suspend": 1, "event.close": 1, "event.archive": 1,
		"offering.create": 1, "offering.update": 1, "offering.publish": 2,
		"offering.unavailable": 1, "offering.archive": 1,
	}
	rows, err := database.Query(`SELECT action, count(*) FROM audit_log WHERE target_id IN ($1, $2) GROUP BY action`, eventID, offeringID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		if expectedAudits[action] != count {
			rows.Close()
			t.Fatalf("audit action %q count = %d, want %d", action, count, expectedAudits[action])
		}
		delete(expectedAudits, action)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(expectedAudits) != 0 {
		t.Fatalf("missing audit action counts: %#v", expectedAudits)
	}

	expectedOutbox := map[string]int{
		"EventPublished": 1, "EventActivated": 2, "EventSuspended": 1,
		"EventClosed": 1, "EventArchived": 1, "OfferingPublished": 2,
		"OfferingUnavailable": 1, "OfferingArchived": 1,
	}
	rows, err = database.Query(`SELECT event_type, count(*) FROM outbox_events WHERE aggregate_id IN ($1, $2) GROUP BY event_type`, eventID, offeringID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		if expectedOutbox[eventType] != count {
			rows.Close()
			t.Fatalf("outbox type %q count = %d, want %d", eventType, count, expectedOutbox[eventType])
		}
		delete(expectedOutbox, eventType)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(expectedOutbox) != 0 {
		t.Fatalf("missing outbox type counts: %#v", expectedOutbox)
	}
}

func assertPayloadsExcludeSecrets(t *testing.T, database *sql.DB, prefix, eventID, offeringID string, secrets []string) {
	t.Helper()
	var outboxPayloads, replayBodies string
	if err := database.QueryRow(`SELECT coalesce(string_agg(payload::text, ' '), '') FROM outbox_events WHERE aggregate_id IN ($1, $2)`, eventID, offeringID).Scan(&outboxPayloads); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT coalesce(string_agg(response_body::text, ' '), '') FROM idempotency_records WHERE idempotency_key LIKE $1`, prefix+".%").Scan(&replayBodies); err != nil {
		t.Fatal(err)
	}
	for _, secret := range secrets {
		if secret != "" && (strings.Contains(outboxPayloads, secret) || strings.Contains(replayBodies, secret)) {
			t.Fatalf("outbox or encrypted replay payload contains a request secret")
		}
	}
}

func cleanupOperationsFixture(t *testing.T, database *sql.DB, prefix, operatorID, eventID, offeringID string) {
	t.Helper()
	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM idempotency_records WHERE idempotency_key LIKE $1`, []any{prefix + ".%"}},
		{`DELETE FROM outbox_events WHERE aggregate_id IN (NULLIF($1, '')::uuid, NULLIF($2, '')::uuid)`, []any{eventID, offeringID}},
		{`DELETE FROM audit_log WHERE actor_operator_id = $1`, []any{operatorID}},
		{`DELETE FROM offerings WHERE id = NULLIF($1, '')::uuid`, []any{offeringID}},
		{`DELETE FROM qurban_events WHERE id = NULLIF($1, '')::uuid`, []any{eventID}},
		{`DELETE FROM operator_sessions WHERE operator_user_id = $1`, []any{operatorID}},
		{`DELETE FROM operator_users WHERE id = $1`, []any{operatorID}},
	}
	for _, statement := range statements {
		if _, err := database.Exec(statement.query, statement.args...); err != nil {
			t.Errorf("cleanup Operations fixture: %v", err)
		}
	}
}

func acquireOperationsIntegrationLock(t *testing.T, database *sql.DB) {
	t.Helper()
	connection, err := database.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, int64(20260813)); err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(20260813))
		_ = connection.Close()
	})
}
