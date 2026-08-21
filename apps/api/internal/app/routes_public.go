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
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/cors"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/ratelimit"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
    "github.com/gin-gonic/gin"
)

func registerPublicRoutes(router *gin.Engine, events event.ActiveReader, offerings offering.PublicCatalogueReader, purchases *purchasing.PublicHandler, limiter *ratelimit.PublicLimiter, logger *slog.Logger, allowedOrigins map[string]struct{}) {
    public := router.Group(publicAPIPrefix)
    methods := []string{http.MethodGet}
    headers := []string{"X-Request-ID"}
    if purchases != nil {
        methods = append(methods, http.MethodPost)
        headers = append(headers, "Content-Type", "Idempotency-Key")
    }
    public.Use(cors.Middleware(cors.Policy{Origins: allowedOrigins, Methods: methods, Headers: headers}), publicNoStore(), publicRateLimit(limiter))
    event.RegisterPublicRoutes(public, events, logger)
    offering.RegisterPublicRoutes(public, offerings, logger)
    if purchases != nil {
        purchases.RegisterRoutes(public)
    }
    for _, path := range []string{"/events/active", "/events/:event_id/offerings", "/offerings/:offering_id"} {
        public.OPTIONS(path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
    }
    if purchases != nil {
        public.OPTIONS("/purchases", func(c *gin.Context) { c.Status(http.StatusNoContent) })
    }
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
