package httpx

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "log/slog"
    "net/http"
    "regexp"
    "time"

    "github.com/gin-gonic/gin"
)

const RequestIDHeader = "X-Request-ID"

var requestIDPattern = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$")

type requestIDContextKey struct{}

func RequestID(ctx context.Context) string {
    value, _ := ctx.Value(requestIDContextKey{}).(string)
    return value
}

func RequestIDMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader(RequestIDHeader)
        if requestID != "" && !requestIDPattern.MatchString(requestID) {
            requestID = newRequestID()
            c.Header(RequestIDHeader, requestID)
            WriteError(c, http.StatusBadRequest, "invalid_request", "invalid X-Request-ID", requestID, nil)
            c.Abort()
            return
        }
        if requestID == "" {
            requestID = newRequestID()
        }

        c.Header(RequestIDHeader, requestID)
        c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDContextKey{}, requestID))
        started := time.Now()
        c.Next()
        logger.Info("HTTP request completed", "method", c.Request.Method, "path", c.Request.URL.Path, "request_id", requestID, "duration", time.Since(started))
    }
}

func newRequestID() string {
    var bytes [16]byte
    if _, err := rand.Read(bytes[:]); err != nil {
        panic("crypto/rand unavailable")
    }
    return hex.EncodeToString(bytes[:])
}
