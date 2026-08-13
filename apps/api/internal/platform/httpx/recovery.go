package httpx

import (
    "fmt"
    "log/slog"
    "net/http"

    "github.com/gin-gonic/gin"
)

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            recovered := recover()
            if recovered == nil {
                return
            }
            if logger != nil {
                logger.Error("HTTP handler panicked", "method", c.Request.Method, "path", c.Request.URL.Path, "request_id", RequestID(c.Request.Context()), "panic_type", fmt.Sprintf("%T", recovered))
            }
            if !c.Writer.Written() {
                WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", RequestID(c.Request.Context()), nil)
            }
            c.Abort()
        }()
        c.Next()
    }
}
