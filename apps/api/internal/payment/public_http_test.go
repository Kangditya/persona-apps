package payment

import (
    "bytes"
    "errors"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "net/textproto"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestDecodePublicEvidenceNormalizesAndHashesTrustedContent(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := []struct {
        filename  string
        mediaType string
        data      []byte
    }{
        {`C:\fakepath\proof.pdf`, "application/pdf", []byte("%PDF-1.4\nproof")},
        {"proof.jpg", "image/jpeg", []byte{0xff, 0xd8, 0xff, 0x00}},
        {"proof.png", "image/png", append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, make([]byte, 8)...)},
    }
    for _, test := range tests {
        body, contentType := evidenceBody(t, "125000", "IDR", test.filename, test.mediaType, test.data)
        request := httptest.NewRequest(http.MethodPost, "/", body)
        request.Header.Set("Content-Type", contentType)
        context, _ := gin.CreateTestContext(httptest.NewRecorder())
        context.Request = request

        value, err := decodePublicEvidence(context)
        if err != nil {
            t.Fatal(err)
        }
        if value.AmountMinor != 125000 || value.CurrencyCode != "IDR" || value.MediaType != test.mediaType || len(value.SHA256) != 32 {
            t.Fatalf("evidence = %#v", value)
        }
    }
}

func TestDecodePublicEvidenceAcceptsExactSizeLimit(t *testing.T) {
    gin.SetMode(gin.TestMode)
    data := append([]byte{0xff, 0xd8, 0xff}, make([]byte, MaxEvidenceBytes-3)...)
    body, contentType := evidenceBody(t, "1", "IDR", "proof.jpg", "image/jpeg", data)
    request := httptest.NewRequest(http.MethodPost, "/", body)
    request.Header.Set("Content-Type", contentType)
    context, _ := gin.CreateTestContext(httptest.NewRecorder())
    context.Request = request
    value, err := decodePublicEvidence(context)
    if err != nil || int64(len(value.Data)) != MaxEvidenceBytes {
        t.Fatalf("size/error = %d/%v", len(value.Data), err)
    }
}

func TestDecodePublicEvidenceRejectsDuplicateParts(t *testing.T) {
    gin.SetMode(gin.TestMode)
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    _ = writer.WriteField("amount_minor", "1")
    _ = writer.WriteField("amount_minor", "1")
    _ = writer.WriteField("currency_code", "IDR")
    header := make(textproto.MIMEHeader)
    header.Set("Content-Disposition", `form-data; name="evidence_file"; filename="proof.pdf"`)
    header.Set("Content-Type", "application/pdf")
    part, _ := writer.CreatePart(header)
    _, _ = part.Write([]byte("%PDF-1.4\nproof"))
    _ = writer.Close()
    request := httptest.NewRequest(http.MethodPost, "/", body)
    request.Header.Set("Content-Type", writer.FormDataContentType())
    context, _ := gin.CreateTestContext(httptest.NewRecorder())
    context.Request = request
    if _, err := decodePublicEvidence(context); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("duplicate part error = %v", err)
    }
}

func TestDecodePublicEvidenceRejectsMismatchedAndOversizedContent(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tests := []struct {
        name      string
        mediaType string
        data      []byte
        want      error
    }{
        {"mismatch", "image/png", []byte("%PDF-1.4\nproof"), ErrUnsupportedMedia},
        {"empty", "application/pdf", nil, ErrInvalidInput},
        {"oversized", "image/jpeg", append([]byte{0xff, 0xd8, 0xff}, make([]byte, MaxEvidenceBytes-2)...), ErrEvidenceTooLarge},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            body, contentType := evidenceBody(t, "1", "IDR", "proof", test.mediaType, test.data)
            request := httptest.NewRequest(http.MethodPost, "/", body)
            request.Header.Set("Content-Type", contentType)
            context, _ := gin.CreateTestContext(httptest.NewRecorder())
            context.Request = request
            _, err := decodePublicEvidence(context)
            if !errors.Is(err, test.want) {
                t.Fatalf("error = %v, want %v", err, test.want)
            }
        })
    }
}

func TestPurchaseBearerRequiresOneBearerCredential(t *testing.T) {
    request := httptest.NewRequest(http.MethodPost, "/", nil)
    if _, err := purchaseBearer(request); err == nil {
        t.Fatal("missing credential accepted")
    }
    request.Header.Set("Authorization", "Bearer token")
    if token, err := purchaseBearer(request); err != nil || token != "token" {
        t.Fatalf("token/error = %q/%v", token, err)
    }
    request.Header.Add("Authorization", "Bearer other")
    if _, err := purchaseBearer(request); err == nil {
        t.Fatal("multiple credentials accepted")
    }
}

func evidenceBody(t *testing.T, amount, currency, filename, mediaType string, data []byte) (*bytes.Buffer, string) {
    t.Helper()
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    if err := writer.WriteField("amount_minor", amount); err != nil {
        t.Fatal(err)
    }
    if err := writer.WriteField("currency_code", currency); err != nil {
        t.Fatal(err)
    }
    header := make(textproto.MIMEHeader)
    header.Set("Content-Disposition", `form-data; name="evidence_file"; filename="`+filename+`"`)
    header.Set("Content-Type", mediaType)
    part, err := writer.CreatePart(header)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := part.Write(data); err != nil {
        t.Fatal(err)
    }
    if err := writer.Close(); err != nil {
        t.Fatal(err)
    }
    return body, writer.FormDataContentType()
}
