package app

import (
    "log/slog"
    "math"
    "net/http"
    "net/netip"
    "strconv"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/ratelimit"
    "github.com/gin-gonic/gin"
)

func registerPublicRoutes(router *gin.Engine, events event.ActiveReader, offerings offering.PublicCatalogueReader, limiter *ratelimit.PublicLimiter, logger *slog.Logger) {
    public := router.Group(publicAPIPrefix)
    public.Use(publicNoStore(), publicRateLimit(limiter))
    event.RegisterPublicRoutes(public, events, logger)
    offering.RegisterPublicRoutes(public, offerings, logger)
}

func publicNoStore() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", "no-store")
        c.Next()
    }
}

func publicRateLimit(limiter *ratelimit.PublicLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        allowed, retryAfter := limiter.Allow(publicClientIPKey(c.ClientIP()))
        if allowed {
            c.Next()
            return
        }
        c.Header("Retry-After", strconv.Itoa(retryAfterSeconds(retryAfter)))
        httpx.WriteError(c, http.StatusTooManyRequests, "rate_limited", "request rate is temporarily limited", httpx.RequestID(c.Request.Context()), nil)
        c.Abort()
    }
}

func publicClientIPKey(value string) string {
    address, err := netip.ParseAddr(value)
    if err != nil {
        return ""
    }
    return address.Unmap().String()
}

func retryAfterSeconds(delay time.Duration) int {
    seconds := int(math.Ceil(delay.Seconds()))
    if seconds < 1 {
        return 1
    }
    if seconds > 60 {
        return 60
    }
    return seconds
}
