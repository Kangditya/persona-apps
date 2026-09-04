package payment

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"
    "fmt"
    "io"
    "log/slog"
    "mime"
    "net/http"
    "path"
    "strings"
    "unicode/utf8"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type OperationsHandler struct {
    database *sql.DB
    store    EvidenceStore
    logger   *slog.Logger
}

func NewOperationsHandler(database *sql.DB, store EvidenceStore, logger *slog.Logger) *OperationsHandler {
    return &OperationsHandler{database: database, store: store, logger: logger}
}

func (handler *OperationsHandler) RegisterRoutes(group *gin.RouterGroup, authentication *auth.Service) {
    group.GET("/payments/:payment_id/evidence", authentication.Require("payment.verify"), handler.getEvidence)
}

func (handler *OperationsHandler) getEvidence(c *gin.Context) {
    id, err := httpx.CanonicalUUID(c.Param("payment_id"))
    if err != nil {
        writeOperationsPaymentError(c, handler.logger, ErrInvalidInput)
        return
    }
    if handler.database == nil || handler.store == nil {
        writeOperationsPaymentError(c, handler.logger, ErrStorageUnavailable)
        return
    }
    evidence, err := NewRepository(handler.database).GetEvidence(c.Request.Context(), id)
    if err != nil {
        writeOperationsPaymentError(c, handler.logger, err)
        return
    }
    if !validEvidenceMetadata(evidence) {
        writeOperationsPaymentError(c, handler.logger, ErrStorageUnavailable)
        return
    }
    if err := handler.verifyEvidence(c.Request.Context(), evidence); err != nil {
        writeOperationsPaymentError(c, handler.logger, err)
        return
    }
    disposition := mime.FormatMediaType("attachment", map[string]string{"filename": evidence.Filename})
    if disposition == "" {
        writeOperationsPaymentError(c, handler.logger, ErrStorageUnavailable)
        return
    }
    file, err := handler.store.Open(c.Request.Context(), evidence.Reference)
    if err != nil {
        writeOperationsPaymentError(c, handler.logger, err)
        return
    }
    defer file.Close()
    c.Header("Cache-Control", "no-store")
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("Content-Type", evidence.MediaType)
    c.Header("Content-Length", fmt.Sprintf("%d", evidence.SizeBytes))
    c.Header("Content-Disposition", disposition)
    c.Status(http.StatusOK)
    _, _ = io.Copy(c.Writer, file)
}

func validEvidenceMetadata(evidence Evidence) bool {
    return validEvidenceKey(evidence.Reference) == nil && validMediaType(evidence.MediaType) && evidence.SizeBytes >= 1 && evidence.SizeBytes <= MaxEvidenceBytes && len(evidence.SHA256) == sha256.Size && strings.TrimSpace(evidence.Filename) != "" && !strings.ContainsAny(evidence.Filename, "\\\r\n\x00") && path.Base(evidence.Filename) == evidence.Filename && utf8.RuneCountInString(evidence.Filename) <= MaxEvidenceFilename
}

func (handler *OperationsHandler) verifyEvidence(ctx context.Context, evidence Evidence) error {
    file, err := handler.store.Open(ctx, evidence.Reference)
    if err != nil {
        return err
    }
    defer file.Close()
    hash := sha256.New()
    count, err := io.Copy(hash, file)
    if err != nil || count != evidence.SizeBytes || !bytes.Equal(hash.Sum(nil), evidence.SHA256) {
        return ErrStorageUnavailable
    }
    return nil
}

func writeOperationsPaymentError(c *gin.Context, logger *slog.Logger, err error) {
    switch {
    case errors.Is(err, ErrInvalidInput):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid payment request", httpx.RequestID(c.Request.Context()), nil)
    case errors.Is(err, ErrEvidenceNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", httpx.RequestID(c.Request.Context()), nil)
    case errors.Is(err, ErrStorageUnavailable), httpx.IsDependencyUnavailable(err):
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service unavailable", httpx.RequestID(c.Request.Context()), nil)
    default:
        if logger != nil {
            logger.Error("payment evidence retrieval failed", "error_type", fmt.Sprintf("%T", err))
        }
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", httpx.RequestID(c.Request.Context()), nil)
    }
}
