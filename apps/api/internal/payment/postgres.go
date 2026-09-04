package payment

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Repository struct {
    database DBTX
}

type scanner interface {
    Scan(...any) error
}

func NewRepository(database DBTX) *Repository {
    return &Repository{database: database}
}

func (repository *Repository) CreateSubmission(ctx context.Context, value Payment) (Payment, error) {
    created, err := scanPayment(repository.database.QueryRowContext(ctx, `
INSERT INTO payment_records (
    event_id, payment_ref, payer_party_id, purchase_id, amount_minor, currency_code, method, status,
    evidence_reference, evidence_filename, evidence_media_type, evidence_size_bytes, evidence_sha256, submitted_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, event_id, payment_ref, payer_party_id, purchase_id, amount_minor, currency_code, method, status,
    evidence_reference, evidence_filename, evidence_media_type, evidence_size_bytes, evidence_sha256,
    submitted_at, verified_at, rejection_reason, created_at, updated_at`,
        value.EventID, value.Reference, value.PayerPartyID, value.PurchaseID, value.AmountMinor,
        value.CurrencyCode, value.Method, value.Status, value.Evidence.Reference, value.Evidence.Filename,
        value.Evidence.MediaType, value.Evidence.SizeBytes, value.Evidence.SHA256, value.SubmittedAt))
    if err != nil {
        return Payment{}, fmt.Errorf("insert payment submission: %w", translateError(err))
    }
    if _, err := repository.database.ExecContext(ctx, `
INSERT INTO payment_status_history (payment_id, from_status, to_status, changed_at)
VALUES ($1, NULL, $2, $3)`, created.ID, StatusSubmitted, created.SubmittedAt); err != nil {
        return Payment{}, fmt.Errorf("insert payment submission history: %w", translateError(err))
    }
    return created, nil
}

// GetEvidence returns only the database-owned evidence metadata for private retrieval.
func (repository *Repository) GetEvidence(ctx context.Context, paymentID string) (Evidence, error) {
    var evidence Evidence
    err := repository.database.QueryRowContext(ctx, `
SELECT evidence_reference, evidence_filename, evidence_media_type, evidence_size_bytes, evidence_sha256
FROM payment_records WHERE id = $1 AND evidence_reference IS NOT NULL`, paymentID).Scan(&evidence.Reference, &evidence.Filename, &evidence.MediaType, &evidence.SizeBytes, &evidence.SHA256)
    if errors.Is(err, sql.ErrNoRows) {
        return Evidence{}, ErrEvidenceNotFound
    }
    if err != nil {
        return Evidence{}, fmt.Errorf("get payment evidence: %w", translateError(err))
    }
    return evidence, nil
}

func (repository *Repository) ReferenceExists(ctx context.Context, reference string) (bool, error) {
    var found bool
    if err := repository.database.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM payment_records WHERE evidence_reference = $1)`, reference).Scan(&found); err != nil {
        return false, fmt.Errorf("check payment evidence reference: %w", err)
    }
    return found, nil
}

func scanPayment(value scanner) (Payment, error) {
    var payment Payment
    var status string
    var verifiedAt sql.NullTime
    var rejectionReason sql.NullString
    if err := value.Scan(
        &payment.ID, &payment.EventID, &payment.Reference, &payment.PayerPartyID, &payment.PurchaseID,
        &payment.AmountMinor, &payment.CurrencyCode, &payment.Method, &status,
        &payment.Evidence.Reference, &payment.Evidence.Filename, &payment.Evidence.MediaType,
        &payment.Evidence.SizeBytes, &payment.Evidence.SHA256, &payment.SubmittedAt, &verifiedAt,
        &rejectionReason, &payment.CreatedAt, &payment.UpdatedAt,
    ); err != nil {
        return Payment{}, err
    }
    payment.Status = Status(status)
    payment.SubmittedAt = payment.SubmittedAt.UTC()
    payment.CreatedAt = payment.CreatedAt.UTC()
    payment.UpdatedAt = payment.UpdatedAt.UTC()
    if verifiedAt.Valid {
        value := verifiedAt.Time.UTC()
        payment.VerifiedAt = &value
    }
    if rejectionReason.Valid {
        payment.RejectionReason = rejectionReason.String
    }
    return payment, nil
}

func translateError(err error) error {
    var postgresError *pgconn.PgError
    if !errors.As(err, &postgresError) {
        return err
    }
    switch postgresError.Code {
    case "23505":
        if postgresError.ConstraintName == "uq_payment_records_one_submitted_purchase" {
            return ErrStateConflict
        }
        return ErrStateConflict
    case "23502", "23503", "23514", "22P02":
        return ErrInvalidInput
    default:
        return err
    }
}
