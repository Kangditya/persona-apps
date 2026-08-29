package app

import (
    "context"
    "fmt"
    "log/slog"
    "math"
    "net/http"
    "net/netip"
    "strconv"
    "strings"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    eventmodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/event"
    eventhttp "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/transport/http"
    offeringmodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering"
    offeringhttp "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/transport/http"
    purchasingmodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/cors"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/ratelimit"
    "github.com/gin-gonic/gin"
)

const (
    publicAPIPrefix     = "/api/public/v1"
    operationsAPIPrefix = "/api/operations/v1"
)

type readinessChecker interface {
    PingContext(context.Context) error
}

func NewServer(address string, logger *slog.Logger, public config.PublicConfig, dependencies Dependencies) (*http.Server, error) {
    router, err := newRouter(dependencies.Database, logger, public, dependencies)
    if err != nil {
        return nil, err
    }
    return &http.Server{
        Addr:              address,
        Handler:           router,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       60 * time.Second,
        ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
    }, nil
}

func newRouter(database readinessChecker, logger *slog.Logger, public config.PublicConfig, dependencies Dependencies) (*gin.Engine, error) {
    router := gin.New()
    router.RedirectTrailingSlash = false
    router.RedirectFixedPath = false
    router.HandleMethodNotAllowed = true
    if err := router.SetTrustedProxies(public.TrustedProxyCIDRs); err != nil {
        return nil, fmt.Errorf("set trusted proxies: %w", err)
    }
    router.NoRoute(func(c *gin.Context) {
        if isAPIPath(c.Request.URL.Path) {
            httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", httpx.RequestID(c.Request.Context()), nil)
            return
        }
        http.NotFound(c.Writer, c.Request)
    })
    router.NoMethod(func(c *gin.Context) {
        if isAPIPath(c.Request.URL.Path) {
            httpx.WriteError(c, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", httpx.RequestID(c.Request.Context()), nil)
            return
        }
        http.Error(c.Writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
    })
    limiter, err := ratelimit.NewPublicLimiter(public.RateLimitPerMinute, public.RateLimitBurst)
    if err != nil {
        return nil, fmt.Errorf("create public rate limiter: %w", err)
    }
    router.Use(httpx.RequestIDMiddleware(logger), httpx.RecoveryMiddleware(logger), operationsNoStore())

    registerSwaggerRoutes(router, resolveSwaggerSpecDirectory())
    registerHealthRoutes(router, database, logger)
    registerPublicRoutes(router, public, limiter, logger, dependencies.Event, dependencies.Offering, dependencies.Purchasing)
    registerOperationsRoutes(router, dependencies.Auth, dependencies.Event, dependencies.Offering, dependencies.Purchasing)
    return router, nil
}

func registerPublicRoutes(router *gin.Engine, public config.PublicConfig, limiter *ratelimit.PublicLimiter, logger *slog.Logger, events *eventmodule.Module, offerings *offeringmodule.Module, purchases *purchasingmodule.Module) {
    publicRouter := router.Group(publicAPIPrefix)
    methods := []string{http.MethodGet}
    headers := []string{"X-Request-ID"}
    if purchases != nil {
        methods = append(methods, http.MethodPost)
        headers = append(headers, "Content-Type", "Idempotency-Key")
    }
    publicRouter.Use(cors.Middleware(cors.Policy{Origins: public.StorefrontAllowedOrigins, Methods: methods, Headers: headers}), publicNoStore(), publicRateLimit(limiter))
    if events != nil {
        events.RegisterPublicRoutes(publicRouter)
    } else {
        eventhttp.RegisterPublicRoutes(publicRouter, nil, logger)
    }
    if offerings != nil {
        offerings.RegisterPublicRoutes(publicRouter)
    } else {
        offeringhttp.RegisterPublicRoutes(publicRouter, nil, logger)
    }
    if purchases != nil {
        purchases.RegisterPublicRoutes(publicRouter)
    }
}

func registerOperationsRoutes(router *gin.Engine, operationsAuth *auth.Service, events *eventmodule.Module, offerings *offeringmodule.Module, purchases *purchasingmodule.Module) {
    operations := router.Group(operationsAPIPrefix)
    if operationsAuth == nil {
        return
    }
    operations.Use(cors.Middleware(cors.Policy{
        Origins: operationsAuth.AllowedOrigins(), Methods: []string{http.MethodGet, http.MethodPost, http.MethodPatch},
        Headers: []string{"Content-Type", "X-CSRF-Token", "Idempotency-Key", "X-Request-ID"}, Credentials: true,
    }))
    operationsAuth.RegisterRoutes(operations)
    if events != nil {
        events.RegisterOperationsRoutes(operations, operationsAuth)
    }
    if offerings != nil {
        offerings.RegisterOperationsRoutes(operations, operationsAuth)
    }
    if purchases != nil {
        purchases.RegisterOperationsRoutes(operations, operationsAuth)
    }
}

func operationsNoStore() gin.HandlerFunc {
    return func(c *gin.Context) {
        if strings.HasPrefix(c.Request.URL.Path, operationsAPIPrefix+"/") {
            c.Header("Cache-Control", "no-store")
        }
        c.Next()
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

func isAPIPath(path string) bool {
    return strings.HasPrefix(path, "/api/")
}
