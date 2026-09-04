package main

import (
    "bytes"
    "context"
    "crypto/sha256"
    "errors"
    "io"
    "log/slog"
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/Kangditya/persona-apps/apps/api/internal/payment"
)

func TestEvidenceReaperKeepsYoungThenRemovesStale(t *testing.T) {
    root := t.TempDir()
    if err := os.Chmod(root, 0o700); err != nil {
        t.Fatal(err)
    }
    store, err := payment.NewFilesystemStore(root)
    if err != nil {
        t.Fatal(err)
    }
    defer store.Close()
    data := []byte("%PDF-1.4\nproof")
    digest := sha256.Sum256(data)
    reference, _ := payment.NewEvidenceReference()
    if _, err := store.Put(context.Background(), payment.EvidenceObject{Reference: reference, MediaType: "application/pdf", SizeBytes: int64(len(data)), SHA256: digest[:], Body: bytes.NewReader(data)}); err != nil {
        t.Fatal(err)
    }
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    repository := payment.NewRepository(db)
    if err := store.Reconcile(context.Background(), repository); err != nil {
        t.Fatal(err)
    }
    file, err := store.Open(context.Background(), reference)
    if err != nil {
        t.Fatal(err)
    }
    _ = file.Close()
    old := time.Now().Add(-25 * time.Hour)
    if err := os.Chtimes(filepath.Join(root, "objects", reference), old, old); err != nil {
        t.Fatal(err)
    }
    mock.ExpectQuery("SELECT EXISTS").WithArgs(reference).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})
    go func() {
        defer close(done)
        reconcileEvidence(ctx, time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil)), store, repository)
    }()
    time.Sleep(20 * time.Millisecond)
    cancel()
    <-done
    if _, err := store.Open(context.Background(), reference); !errors.Is(err, payment.ErrEvidenceNotFound) {
        t.Fatalf("stale = %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}
