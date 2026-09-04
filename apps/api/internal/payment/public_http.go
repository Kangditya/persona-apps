package payment

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log/slog"
    "mime"
    "mime/multipart"
    "net/http"
    "os"
    "path"
    "strconv"
    "strings"
    "time"
    "unicode/utf8"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
    "github.com/gin-gonic/gin"
)

const publicEvidenceBodyLimit = MaxEvidenceBytes + 64<<10

type PublicHandler struct {
    database *sql.DB
    cipher   idempotency.Cipher
    logger   *slog.Logger
    store    EvidenceStore
    now      func() time.Time
    random   io.Reader
}

type publicEvidenceRequest struct {
    AmountMinor  int64
    CurrencyCode string
    Filename     string
    MediaType    string
    Spool        *os.File
    SizeBytes    int64
    SHA256       []byte
}

type publicPayment struct {
    ID              string    `json:"id"`
    Reference       string    `json:"payment_ref"`
    AmountMinor     int64     `json:"amount_minor"`
    CurrencyCode    string    `json:"currency_code"`
    Status          Status    `json:"status"`
    SubmittedAt     time.Time `json:"submitted_at"`
    RejectionReason string    `json:"rejection_reason,omitempty"`
}

type publicPaymentResponse struct {
    Data publicPayment `json:"data"`
}

func NewPublicHandler(database *sql.DB, cipher idempotency.Cipher, logger *slog.Logger, store EvidenceStore) *PublicHandler {
    return &PublicHandler{database: database, cipher: cipher, logger: logger, store: store, now: time.Now, random: rand.Reader}
}

func (handler *PublicHandler) RegisterRoutes(group *gin.RouterGroup) {
    group.POST("/purchases/:purchase_id/payment-evidence", handler.submit)
}

func (handler *PublicHandler) submit(c *gin.Context) {
    purchaseID, err := httpx.CanonicalUUID(c.Param("purchase_id"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid Purchase identifier", httpx.RequestID(c.Request.Context()), nil)
        return
    }
    token, err := purchaseBearer(c.Request)
    if err != nil {
        writePublicPaymentError(c, handler.logger, err)
        return
    }
    if handler.database == nil || handler.store == nil {
        writePublicPaymentError(c, handler.logger, httpx.ErrDependencyUnavailable)
        return
    }
    if err := purchasing.AuthorizePurchaseToken(c.Request.Context(), handler.database, purchaseID, token); err != nil {
        writePublicPaymentError(c, handler.logger, err)
        return
    }
    key, err := httpx.IdempotencyKey(c.GetHeader("Idempotency-Key"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid Idempotency-Key", httpx.RequestID(c.Request.Context()), nil)
        return
    }
    request, err := decodePublicEvidence(c)
    if err != nil {
        writePublicPaymentError(c, handler.logger, err)
        return
    }
    defer closeSpool(request.Spool)
    defer clear(request.SHA256)
    requestHash, err := idempotency.RequestHash(struct {
        PurchaseID   string `json:"purchase_id"`
        AmountMinor  int64  `json:"amount_minor"`
        CurrencyCode string `json:"currency_code"`
        Filename     string `json:"filename"`
        MediaType    string `json:"media_type"`
        SizeBytes    int64  `json:"size_bytes"`
        SHA256       string `json:"sha256"`
    }{
        PurchaseID: purchaseID, AmountMinor: request.AmountMinor, CurrencyCode: request.CurrencyCode,
        Filename: request.Filename, MediaType: request.MediaType, SizeBytes: request.SizeBytes,
        SHA256: base64.RawURLEncoding.EncodeToString(request.SHA256),
    })
    if err != nil {
        writePublicPaymentError(c, handler.logger, err)
        return
    }
    namespace, _ := idempotency.Namespace("storefront", "purchase.payment-evidence.submit", purchaseID)
    command := idempotency.Command{Namespace: namespace, Key: key, RequestHash: requestHash, Retention: 0}
    occurredAt := handler.clock()

    submissionContext, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
    defer cancel()
    var storedReference string
    armed := false
    defer func() {
        if recovered := recover(); recovered != nil {
            if armed {
                handler.cleanupEvidence(c.Request.Context(), storedReference)
            }
            panic(recovered)
        }
        handler.cleanupAfterSubmission(c.Request.Context(), armed, storedReference, err)
    }()
    response, _, err := idempotency.Execute(submissionContext, handler.database, handler.cipher, command, func(transaction *sql.Tx) (idempotency.Response, error) {
        source, prepareErr := purchasing.PreparePaymentSubmission(submissionContext, transaction, purchaseID, occurredAt)
        if prepareErr != nil {
            return idempotency.Response{}, prepareErr
        }
        if request.AmountMinor != source.Purchase.TotalAmountMinor || request.CurrencyCode != source.Purchase.CurrencyCode {
            return idempotency.Response{}, ErrInvalidInput
        }
        reference, referenceErr := NewEvidenceReference()
        if referenceErr != nil {
            return idempotency.Response{}, referenceErr
        }
        if _, seekErr := request.Spool.Seek(0, io.SeekStart); seekErr != nil {
            return idempotency.Response{}, fmt.Errorf("seek payment evidence: %w", seekErr)
        }
        stored, storeErr := handler.store.Put(submissionContext, EvidenceObject{
            Reference: reference, MediaType: request.MediaType, SizeBytes: request.SizeBytes, SHA256: request.SHA256,
            Body: request.Spool,
        })
        if storeErr != nil {
            return idempotency.Response{}, fmt.Errorf("store payment evidence: %w", storeErr)
        }
        storedReference = reference
        armed = true
        if strings.TrimSpace(stored.Reference) != reference || stored.SizeBytes != request.SizeBytes {
            return idempotency.Response{}, ErrStorageUnavailable
        }
        paymentReference, referenceErr := NewReference(handler.randomSource())
        if referenceErr != nil {
            return idempotency.Response{}, referenceErr
        }
        payerID := ""
        if source.Purchase.PayerPartyID != nil {
            payerID = *source.Purchase.PayerPartyID
        }
        submission, submissionErr := NewSubmission(SubmissionInput{
            Reference: paymentReference, EventID: source.Purchase.EventID, PayerPartyID: payerID,
            PurchaseID: source.Purchase.ID, AmountMinor: request.AmountMinor, CurrencyCode: request.CurrencyCode,
            Evidence:    Evidence{Reference: storedReference, Filename: request.Filename, MediaType: request.MediaType, SizeBytes: request.SizeBytes, SHA256: request.SHA256},
            SubmittedAt: occurredAt,
        })
        if submissionErr != nil {
            return idempotency.Response{}, submissionErr
        }
        created, createErr := NewRepository(transaction).CreateSubmission(submissionContext, submission)
        if createErr != nil {
            return idempotency.Response{}, createErr
        }
        if outboxErr := outbox.Write(submissionContext, transaction, outbox.Event{
            AggregateType: "Payment", AggregateID: created.ID, EventType: "PaymentEvidenceSubmitted",
            Payload: map[string]string{"payment_id": created.ID, "purchase_id": created.PurchaseID}, OccurredAt: created.SubmittedAt,
        }); outboxErr != nil {
            return idempotency.Response{}, outboxErr
        }
        body, marshalErr := json.Marshal(publicPaymentResponse{Data: toPublicPayment(created)})
        if marshalErr != nil {
            return idempotency.Response{}, fmt.Errorf("marshal public payment response: %w", marshalErr)
        }
        return idempotency.Response{Status: http.StatusCreated, Body: body}, nil
    })
    if err != nil {
        writePublicPaymentError(c, handler.logger, err)
        return
    }
    armed = false
    c.Data(response.Status, "application/json", response.Body)
}

func decodePublicEvidence(c *gin.Context) (publicEvidenceRequest, error) {
    mediaType, parameters, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
    if err != nil || mediaType != "multipart/form-data" || parameters["boundary"] == "" {
        return publicEvidenceRequest{}, ErrUnsupportedMedia
    }
    c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, publicEvidenceBodyLimit)
    reader := multipart.NewReader(c.Request.Body, parameters["boundary"])
    var result publicEvidenceRequest
    keepSpool := false
    defer func() {
        if !keepSpool {
            closeSpool(result.Spool)
        }
    }()
    seen := map[string]bool{}
    for {
        part, nextErr := reader.NextPart()
        if errors.Is(nextErr, io.EOF) {
            break
        }
        if nextErr != nil {
            var tooLarge *http.MaxBytesError
            if errors.As(nextErr, &tooLarge) {
                return publicEvidenceRequest{}, ErrEvidenceTooLarge
            }
            return publicEvidenceRequest{}, ErrInvalidInput
        }
        name := part.FormName()
        if name == "" || seen[name] {
            _ = part.Close()
            return publicEvidenceRequest{}, ErrInvalidInput
        }
        seen[name] = true
        switch name {
        case "amount_minor":
            value, readErr := readTextPart(part, 32)
            if readErr != nil {
                return publicEvidenceRequest{}, readErr
            }
            result.AmountMinor, err = strconv.ParseInt(value, 10, 64)
            if err != nil || result.AmountMinor <= 0 || result.AmountMinor > 9_007_199_254_740_991 {
                return publicEvidenceRequest{}, ErrInvalidInput
            }
        case "currency_code":
            result.CurrencyCode, err = readTextPart(part, 3)
            if err != nil || len(result.CurrencyCode) != 3 || strings.ToUpper(result.CurrencyCode) != result.CurrencyCode {
                return publicEvidenceRequest{}, ErrInvalidInput
            }
        case "evidence_file":
            result, err = readEvidencePart(c.Request.Context(), part, result)
            if err != nil {
                return publicEvidenceRequest{}, err
            }
        default:
            _ = part.Close()
            return publicEvidenceRequest{}, ErrInvalidInput
        }
    }
    if !seen["amount_minor"] || !seen["currency_code"] || !seen["evidence_file"] {
        return publicEvidenceRequest{}, ErrInvalidInput
    }
    keepSpool = true
    return result, nil
}

func readTextPart(part *multipart.Part, maximum int64) (string, error) {
    defer part.Close()
    if part.FileName() != "" {
        return "", ErrInvalidInput
    }
    value, err := io.ReadAll(io.LimitReader(part, maximum+1))
    if err != nil || int64(len(value)) > maximum {
        return "", ErrInvalidInput
    }
    result := strings.TrimSpace(string(value))
    if result == "" {
        return "", ErrInvalidInput
    }
    return result, nil
}

func readEvidencePart(ctx context.Context, part *multipart.Part, result publicEvidenceRequest) (publicEvidenceRequest, error) {
    defer part.Close()
    filename := strings.TrimSpace(path.Base(strings.ReplaceAll(part.FileName(), "\\", "/")))
    if filename == "" || filename == "." || utf8.RuneCountInString(filename) > MaxEvidenceFilename {
        return publicEvidenceRequest{}, ErrInvalidInput
    }
    spool, err := os.CreateTemp("", "payment-evidence-")
    if err != nil {
        return publicEvidenceRequest{}, ErrStorageUnavailable
    }
    if err := spool.Chmod(0o600); err != nil || os.Remove(spool.Name()) != nil {
        _ = spool.Close()
        return publicEvidenceRequest{}, ErrStorageUnavailable
    }
    keep := false
    defer func() {
        if !keep {
            _ = spool.Close()
        }
    }()
    hash := sha256.New()
    sample, buffer := make([]byte, 0, 512), make([]byte, 32<<10)
    source := io.LimitReader(part, MaxEvidenceBytes+1)
    var size int64
    for {
        if err := ctx.Err(); err != nil {
            return publicEvidenceRequest{}, err
        }
        read, readErr := source.Read(buffer)
        if read > 0 {
            size += int64(read)
            if size > MaxEvidenceBytes {
                return publicEvidenceRequest{}, ErrEvidenceTooLarge
            }
            if len(sample) < 512 {
                take := min(512-len(sample), read)
                sample = append(sample, buffer[:take]...)
            }
            if _, err := hash.Write(buffer[:read]); err != nil {
                return publicEvidenceRequest{}, ErrInvalidInput
            }
            if _, err := spool.Write(buffer[:read]); err != nil {
                return publicEvidenceRequest{}, ErrStorageUnavailable
            }
        }
        if errors.Is(readErr, io.EOF) {
            break
        }
        if readErr != nil {
            return publicEvidenceRequest{}, ErrInvalidInput
        }
    }
    if size == 0 {
        return publicEvidenceRequest{}, ErrInvalidInput
    }
    declared, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
    detected := http.DetectContentType(sample)
    if err != nil || !validMediaType(declared) || declared != detected {
        return publicEvidenceRequest{}, ErrUnsupportedMedia
    }
    result.Filename = filename
    result.MediaType = detected
    result.Spool = spool
    result.SizeBytes = size
    result.SHA256 = hash.Sum(nil)
    keep = true
    return result, nil
}

func closeSpool(file *os.File) {
    if file != nil {
        _ = file.Close()
    }
}

func purchaseBearer(request *http.Request) (string, error) {
    values := request.Header.Values("Authorization")
    if len(values) != 1 {
        return "", purchasing.ErrUnauthenticated
    }
    parts := strings.Fields(values[0])
    if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
        return "", purchasing.ErrUnauthenticated
    }
    return parts[1], nil
}

func toPublicPayment(value Payment) publicPayment {
    return publicPayment{
        ID: value.ID, Reference: value.Reference, AmountMinor: value.AmountMinor,
        CurrencyCode: value.CurrencyCode, Status: value.Status, SubmittedAt: value.SubmittedAt,
        RejectionReason: value.RejectionReason,
    }
}

func (handler *PublicHandler) clock() time.Time {
    if handler.now == nil {
        return time.Now().UTC()
    }
    return handler.now().UTC()
}

func (handler *PublicHandler) randomSource() io.Reader {
    if handler.random == nil {
        return rand.Reader
    }
    return handler.random
}

func (handler *PublicHandler) cleanupEvidence(ctx context.Context, reference string) {
    cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
    defer cancel()
    if cleanupErr := handler.store.Delete(cleanupContext, reference); cleanupErr != nil && handler.logger != nil {
        handler.logger.Warn("Payment evidence cleanup failed", "request_id", httpx.RequestID(ctx), "error_type", fmt.Sprintf("%T", cleanupErr))
    }
}

func (handler *PublicHandler) cleanupAfterSubmission(ctx context.Context, armed bool, reference string, result error) {
    if armed && !errors.Is(result, database.ErrCommitUncertain) {
        handler.cleanupEvidence(ctx, reference)
    }
}

func writePublicPaymentError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, purchasing.ErrUnauthenticated):
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "authentication required", requestID, nil)
    case errors.Is(err, purchasing.ErrForbidden):
        httpx.WriteError(c, http.StatusForbidden, "forbidden", "request is not authorized", requestID, nil)
    case errors.Is(err, purchasing.ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", requestID, nil)
    case errors.Is(err, ErrEvidenceTooLarge):
        httpx.WriteError(c, http.StatusRequestEntityTooLarge, "evidence_too_large", "payment evidence is too large", requestID, nil)
    case errors.Is(err, ErrUnsupportedMedia):
        httpx.WriteError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "payment evidence media type is unsupported", requestID, nil)
    case errors.Is(err, ErrInvalidInput), errors.Is(err, purchasing.ErrInvalidInput):
        httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "payment evidence validation failed", requestID, nil)
    case errors.Is(err, ErrStateConflict), errors.Is(err, purchasing.ErrStateConflict):
        httpx.WriteError(c, http.StatusConflict, "state_conflict", "payment state conflicts with the request", requestID, nil)
    case errors.Is(err, purchasing.ErrQuotaUnavailable):
        httpx.WriteError(c, http.StatusConflict, "quota_unavailable", "participant quota is unavailable", requestID, nil)
    case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
        httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
    case errors.Is(err, database.ErrCommitUncertain), errors.Is(err, ErrStorageCollision), errors.Is(err, ErrStorageUnavailable), httpx.IsDependencyUnavailable(err):
        logPublicPaymentError(logger, slog.LevelWarn, requestID, err)
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
    default:
        logPublicPaymentError(logger, slog.LevelError, requestID, err)
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", requestID, nil)
    }
}

func logPublicPaymentError(logger *slog.Logger, level slog.Level, requestID string, err error) {
    if logger == nil {
        return
    }
    logger.Log(context.Background(), level, "Public payment request failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
}
