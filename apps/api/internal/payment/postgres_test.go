package payment

import (
    "context"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryCreateSubmissionAppendsInitialHistory(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer database.Close()
    now := time.Now().UTC().Truncate(time.Microsecond)
    value, err := NewSubmission(SubmissionInput{
        Reference: "PAY-AAAAAAAAAAAAAAAAAAAAAAAAAA", EventID: "event", PayerPartyID: "payer", PurchaseID: "purchase",
        AmountMinor: 100, CurrencyCode: "IDR", SubmittedAt: now,
        Evidence: Evidence{Reference: "evidence/key", Filename: "proof.pdf", MediaType: "application/pdf", SizeBytes: 4, SHA256: make([]byte, 32)},
    })
    if err != nil {
        t.Fatal(err)
    }
    mock.ExpectQuery("INSERT INTO payment_records").
        WithArgs("event", value.Reference, "payer", "purchase", int64(100), "IDR", MethodManualTransfer, StatusSubmitted, "evidence/key", "proof.pdf", "application/pdf", int64(4), value.Evidence.SHA256, now).
        WillReturnRows(sqlmock.NewRows([]string{
            "id", "event_id", "payment_ref", "payer_party_id", "purchase_id", "amount_minor", "currency_code", "method", "status",
            "evidence_reference", "evidence_filename", "evidence_media_type", "evidence_size_bytes", "evidence_sha256",
            "submitted_at", "verified_at", "rejection_reason", "created_at", "updated_at",
        }).AddRow("payment", "event", value.Reference, "payer", "purchase", 100, "IDR", MethodManualTransfer, StatusSubmitted,
            "evidence/key", "proof.pdf", "application/pdf", 4, value.Evidence.SHA256, now, nil, nil, now, now))
    mock.ExpectExec("INSERT INTO payment_status_history").WithArgs("payment", StatusSubmitted, now).WillReturnResult(sqlmock.NewResult(1, 1))

    created, err := NewRepository(database).CreateSubmission(context.Background(), value)
    if err != nil {
        t.Fatal(err)
    }
    if created.ID != "payment" || created.Status != StatusSubmitted || created.Evidence.Reference != "evidence/key" {
        t.Fatalf("created = %#v", created)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}
