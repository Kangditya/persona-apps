package payment

import (
    "context"
    "crypto/rand"
    "encoding/base32"
    "encoding/base64"
    "errors"
    "fmt"
    "io"
    "strings"
    "time"
    "unicode/utf8"
)

const (
    MaxEvidenceBytes               = int64(10 << 20)
    MaxEvidenceFilename            = 255
    referenceEntropyBytes          = 16
    evidenceKeyEntropyBytes        = 32
    MethodManualTransfer           = "MANUAL_TRANSFER"
    StatusSubmitted         Status = "SUBMITTED"
)

var (
    ErrInvalidInput       = errors.New("invalid payment input")
    ErrStateConflict      = errors.New("payment state conflict")
    ErrEvidenceTooLarge   = errors.New("payment evidence is too large")
    ErrUnsupportedMedia   = errors.New("unsupported payment evidence media type")
    ErrStorageUnavailable = errors.New("payment evidence storage unavailable")
    ErrEvidenceNotFound   = errors.New("payment evidence not found")
    ErrStorageCollision   = errors.New("payment evidence storage collision")
)

type Status string

type Evidence struct {
    Reference string
    Filename  string
    MediaType string
    SizeBytes int64
    SHA256    []byte
}

type EvidenceObject struct {
    Reference string
    MediaType string
    SizeBytes int64
    SHA256    []byte
    Body      io.Reader
}

type StoredEvidence struct {
    Reference string
    SizeBytes int64
}

type EvidenceStore interface {
    Put(context.Context, EvidenceObject) (StoredEvidence, error)
    Open(context.Context, string) (io.ReadCloser, error)
    Delete(context.Context, string) error
}

type SubmissionInput struct {
    Reference    string
    EventID      string
    PayerPartyID string
    PurchaseID   string
    AmountMinor  int64
    CurrencyCode string
    Evidence     Evidence
    SubmittedAt  time.Time
}

type Payment struct {
    ID              string
    EventID         string
    Reference       string
    PayerPartyID    string
    PurchaseID      string
    AmountMinor     int64
    CurrencyCode    string
    Method          string
    Status          Status
    Evidence        Evidence
    SubmittedAt     time.Time
    VerifiedAt      *time.Time
    RejectionReason string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

func NewSubmission(input SubmissionInput) (Payment, error) {
    value := Payment{
        EventID: strings.TrimSpace(input.EventID), Reference: strings.TrimSpace(input.Reference),
        PayerPartyID: strings.TrimSpace(input.PayerPartyID), PurchaseID: strings.TrimSpace(input.PurchaseID),
        AmountMinor: input.AmountMinor, CurrencyCode: input.CurrencyCode, Method: MethodManualTransfer,
        Status: StatusSubmitted, Evidence: input.Evidence, SubmittedAt: input.SubmittedAt.UTC(),
    }
    value.Evidence.Reference = strings.TrimSpace(value.Evidence.Reference)
    value.Evidence.Filename = strings.TrimSpace(value.Evidence.Filename)
    if err := validateSubmission(value); err != nil {
        return Payment{}, err
    }
    value.Evidence.SHA256 = append([]byte(nil), value.Evidence.SHA256...)
    return value, nil
}

func NewReference(random io.Reader) (string, error) {
    if random == nil {
        return "", ErrInvalidInput
    }
    entropy := make([]byte, referenceEntropyBytes)
    if _, err := io.ReadFull(random, entropy); err != nil {
        return "", fmt.Errorf("generate payment reference: %w", err)
    }
    return "PAY-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entropy), nil
}

// NewEvidenceReference creates a flat, opaque, caller-owned storage identity.
func NewEvidenceReference() (string, error) {
    entropy := make([]byte, evidenceKeyEntropyBytes)
    if _, err := io.ReadFull(rand.Reader, entropy); err != nil {
        return "", fmt.Errorf("generate evidence reference: %w", err)
    }
    return base64.RawURLEncoding.EncodeToString(entropy), nil
}

func validateSubmission(value Payment) error {
    if value.EventID == "" || value.PayerPartyID == "" || value.PurchaseID == "" || !validReference(value.Reference) {
        return ErrInvalidInput
    }
    if value.AmountMinor <= 0 || value.AmountMinor > 9_007_199_254_740_991 || !validCurrency(value.CurrencyCode) || value.SubmittedAt.IsZero() {
        return ErrInvalidInput
    }
    if value.Evidence.Reference == "" || value.Evidence.Filename == "" || utf8.RuneCountInString(value.Evidence.Filename) > MaxEvidenceFilename || value.Evidence.SizeBytes < 1 || value.Evidence.SizeBytes > MaxEvidenceBytes || len(value.Evidence.SHA256) != 32 || !validMediaType(value.Evidence.MediaType) {
        return ErrInvalidInput
    }
    return nil
}

func validReference(value string) bool {
    if len(value) != 30 || !strings.HasPrefix(value, "PAY-") {
        return false
    }
    for _, character := range value[4:] {
        if (character < 'A' || character > 'Z') && (character < '2' || character > '7') {
            return false
        }
    }
    return true
}

func validCurrency(value string) bool {
    if len(value) != 3 {
        return false
    }
    for _, character := range value {
        if character < 'A' || character > 'Z' {
            return false
        }
    }
    return true
}

func validMediaType(value string) bool {
    switch value {
    case "image/jpeg", "image/png", "application/pdf":
        return true
    default:
        return false
    }
}
