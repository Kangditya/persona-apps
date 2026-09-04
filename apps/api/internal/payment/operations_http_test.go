package payment

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "strconv"
    "strings"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/gin-gonic/gin"
)

type operationStore struct {
    body  []byte
    err   error
    opens *int
}

func (store operationStore) Put(context.Context, EvidenceObject) (StoredEvidence, error) {
    return StoredEvidence{}, ErrStorageUnavailable
}
func (store operationStore) Open(context.Context, string) (io.ReadCloser, error) {
    if store.opens != nil {
        *store.opens++
    }
    if store.err != nil {
        return nil, store.err
    }
    return io.NopCloser(bytes.NewReader(store.body)), nil
}

func TestOperationsEvidenceRejectsUnsafeDatabaseFilenameBeforeOpen(t *testing.T) {
    gin.SetMode(gin.TestMode)
    reference := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    digest := make([]byte, sha256.Size)
    for _, filename := range []string{"proof\\.pdf", "proof\r\n.pdf", "proof\x00.pdf"} {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatal(err)
        }
        logs := &bytes.Buffer{}
        opens := 0
        handler := NewOperationsHandler(db, operationStore{body: []byte("content-sentinel"), opens: &opens}, slog.New(slog.NewTextHandler(logs, nil)))
        router := gin.New()
        router.GET("/payments/:payment_id/evidence", handler.getEvidence)
        mock.ExpectQuery("SELECT evidence_reference").WithArgs("11111111-1111-1111-1111-111111111111").WillReturnRows(sqlmock.NewRows([]string{"evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256"}).AddRow(reference, filename, "application/pdf", 1, digest))
        response := httptest.NewRecorder()
        router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/payments/11111111-1111-1111-1111-111111111111/evidence", nil))
        if response.Code != http.StatusServiceUnavailable || opens != 0 || response.Header().Get("Content-Disposition") != "" || strings.Contains(response.Body.String(), filename) || strings.Contains(logs.String(), filename) || strings.Contains(logs.String(), reference) {
            t.Fatal("unsafe filename was not safely rejected")
        }
        _ = db.Close()
    }
}
func (store operationStore) Delete(context.Context, string) error { return nil }

func TestRepositoryEvidenceQueries(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    reference := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    digest := make([]byte, sha256.Size)
    mock.ExpectQuery("SELECT evidence_reference").WithArgs("11111111-1111-1111-1111-111111111111").WillReturnRows(sqlmock.NewRows([]string{"evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256"}).AddRow(reference, "proof.pdf", "application/pdf", 4, digest))
    got, err := NewRepository(db).GetEvidence(context.Background(), "11111111-1111-1111-1111-111111111111")
    if err != nil || got.Reference != reference {
        t.Fatal("evidence query failed")
    }
    mock.ExpectQuery("SELECT EXISTS").WithArgs(reference).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
    if found, err := NewRepository(db).ReferenceExists(context.Background(), reference); err != nil || !found {
        t.Fatal("reference existence query failed")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
    mock.ExpectQuery("SELECT evidence_reference").WithArgs("22222222-2222-2222-2222-222222222222").WillReturnRows(sqlmock.NewRows([]string{"evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256"}))
    if _, err := NewRepository(db).GetEvidence(context.Background(), "22222222-2222-2222-2222-222222222222"); !errors.Is(err, ErrEvidenceNotFound) {
        t.Fatal("no-evidence mapping failed")
    }
}

func TestOperationsEvidenceSafeHeadersAndTamper(t *testing.T) {
    gin.SetMode(gin.TestMode)
    data := []byte("%PDF-1.4\nproof")
    digest := sha256.Sum256(data)
    reference := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    newServer := func(body []byte, storeErr error) (*sql.DB, sqlmock.Sqlmock, *gin.Engine, *bytes.Buffer) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatal(err)
        }
        logs := &bytes.Buffer{}
        handler := NewOperationsHandler(db, operationStore{body: body, err: storeErr}, slog.New(slog.NewTextHandler(logs, nil)))
        router := gin.New()
        router.GET("/payments/:payment_id/evidence", handler.getEvidence)
        return db, mock, router, logs
    }
    db, mock, router, _ := newServer(data, nil)
    defer db.Close()
    mock.ExpectQuery("SELECT evidence_reference").WithArgs("11111111-1111-1111-1111-111111111111").WillReturnRows(sqlmock.NewRows([]string{"evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256"}).AddRow(reference, "proof.pdf", "application/pdf", len(data), digest[:]))
    response := httptest.NewRecorder()
    router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/payments/11111111-1111-1111-1111-111111111111/evidence", nil))
    if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), data) || response.Header().Get("Content-Type") != "application/pdf" || response.Header().Get("Content-Length") != strconv.Itoa(len(data)) || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("Content-Disposition") == "" {
        t.Fatal("safe success response mismatch")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
    db2, mock2, router2, logs := newServer([]byte("tampered"), nil)
    defer db2.Close()
    mock2.ExpectQuery("SELECT evidence_reference").WithArgs("11111111-1111-1111-1111-111111111111").WillReturnRows(sqlmock.NewRows([]string{"evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256"}).AddRow(reference, "proof.pdf", "application/pdf", len(data), digest[:]))
    tampered := httptest.NewRecorder()
    router2.ServeHTTP(tampered, httptest.NewRequest(http.MethodGet, "/payments/11111111-1111-1111-1111-111111111111/evidence", nil))
    if tampered.Code != http.StatusServiceUnavailable || bytes.Contains(tampered.Body.Bytes(), []byte(reference)) {
        t.Fatal("tampered object mapping exposed data")
    }
    if strings.Contains(logs.String(), reference) || strings.Contains(logs.String(), "proof.pdf") || strings.Contains(logs.String(), string(data)) {
        t.Fatal("failure logs leaked evidence metadata")
    }
    // A valid DB record whose object disappeared is a safe not-found result.
    db3, mock3, router3, logs3 := newServer(data, ErrEvidenceNotFound)
    defer db3.Close()
    mock3.ExpectQuery("SELECT evidence_reference").WithArgs("11111111-1111-1111-1111-111111111111").WillReturnRows(sqlmock.NewRows([]string{"evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256"}).AddRow(reference, "proof.pdf", "application/pdf", len(data), digest[:]))
    missing := httptest.NewRecorder()
    router3.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/payments/11111111-1111-1111-1111-111111111111/evidence", nil))
    if missing.Code != http.StatusNotFound || strings.Contains(missing.Body.String(), reference) || strings.Contains(logs3.String(), reference) {
        t.Fatal("missing object leaked metadata")
    }
}
