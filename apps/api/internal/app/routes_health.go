package app

import (
    "context"
    "io"
    "log/slog"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

const readinessTimeout = 2 * time.Second

func registerHealthRoutes(router *gin.Engine, database readinessChecker, logger *slog.Logger) {
    router.GET("/health", func(c *gin.Context) {
        writeStatus(c, http.StatusOK, "ok")
    })
    router.GET("/ready", func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
        defer cancel()

        if err := database.PingContext(ctx); err != nil {
            logger.Warn("readiness check failed", "error", err)
            writeStatus(c, http.StatusServiceUnavailable, "unavailable")
            return
        }

        writeStatus(c, http.StatusOK, "ready")
    })
}

func writeStatus(c *gin.Context, code int, status string) {
    c.Header("Content-Type", "application/json")
    c.Status(code)
    _, _ = io.WriteString(c.Writer, "{\"status\":\""+status+"\"}\n")
}
