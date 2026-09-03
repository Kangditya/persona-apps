package payment

import (
    "bytes"
    "errors"
    "testing"
    "time"
)

func TestNewReferenceAndSubmission(t *testing.T) {
    reference, err := NewReference(bytes.NewReader(make([]byte, referenceEntropyBytes)))
    if err != nil {
        t.Fatal(err)
    }
    if reference != "PAY-AAAAAAAAAAAAAAAAAAAAAAAAAA" {
        t.Fatalf("reference = %q", reference)
    }
    digest := bytes.Repeat([]byte{1}, 32)
    submittedAt := time.Date(2026, time.September, 3, 12, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
    value, err := NewSubmission(SubmissionInput{
        Reference: reference, EventID: "event", PayerPartyID: "payer", PurchaseID: "purchase",
        AmountMinor: 125_000, CurrencyCode: "IDR", SubmittedAt: submittedAt,
        Evidence: Evidence{Reference: "evidence/key", Filename: "proof.pdf", MediaType: "application/pdf", SizeBytes: 4, SHA256: digest},
    })
    if err != nil {
        t.Fatal(err)
    }
    if value.Method != MethodManualTransfer || value.Status != StatusSubmitted || value.SubmittedAt.Location() != time.UTC || value.Reference != reference {
        t.Fatalf("submission = %#v", value)
    }
    digest[0] = 9
    if value.Evidence.SHA256[0] != 1 {
        t.Fatal("submission retained caller-owned digest memory")
    }
}

func TestNewSubmissionRejectsInvalidEvidenceAndMoney(t *testing.T) {
    valid := SubmissionInput{
        Reference: "PAY-AAAAAAAAAAAAAAAAAAAAAAAAAA", EventID: "event", PayerPartyID: "payer", PurchaseID: "purchase",
        AmountMinor: 1, CurrencyCode: "IDR", SubmittedAt: time.Now(),
        Evidence: Evidence{Reference: "evidence/key", Filename: "proof.jpg", MediaType: "image/jpeg", SizeBytes: 1, SHA256: make([]byte, 32)},
    }
    tests := []struct {
        name   string
        mutate func(*SubmissionInput)
    }{
        {"zero amount", func(value *SubmissionInput) { value.AmountMinor = 0 }},
        {"lowercase currency", func(value *SubmissionInput) { value.CurrencyCode = "idr" }},
        {"bad reference", func(value *SubmissionInput) { value.Reference = "PAY-1" }},
        {"empty filename", func(value *SubmissionInput) { value.Evidence.Filename = "" }},
        {"oversize", func(value *SubmissionInput) { value.Evidence.SizeBytes = MaxEvidenceBytes + 1 }},
        {"bad digest", func(value *SubmissionInput) { value.Evidence.SHA256 = []byte{1} }},
        {"bad media", func(value *SubmissionInput) { value.Evidence.MediaType = "text/plain" }},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            input := valid
            input.Evidence.SHA256 = append([]byte(nil), valid.Evidence.SHA256...)
            test.mutate(&input)
            if _, err := NewSubmission(input); !errors.Is(err, ErrInvalidInput) {
                t.Fatalf("error = %v", err)
            }
        })
    }
}
