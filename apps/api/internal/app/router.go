package app

import (
    "fmt"
    "log/slog"
    "net/http"
    "strings"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
    "github.com/Kangditya/persona-apps/apps/api/internal/payment"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/ratelimit"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
    "github.com/gin-gonic/gin"
)

const (
    publicAPIPrefix     = "/api/public/v1"
    operationsAPIPrefix = "/api/operations/v1"
)

func newRouter(database readinessChecker, logger *slog.Logger, public config.PublicConfig, publicPurchases *purchasing.PublicHandler, publicPayments *payment.PublicHandler, operationsAuth *auth.Service, events event.ActiveReader, offerings offering.PublicCatalogueReader, eventOperations *event.OperationsHandler, offeringOperations *offering.OperationsHandler, purchaseOperations *purchasing.OperationsHandler, paymentOperations *payment.OperationsHandler) (*gin.Engine, error) {
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

    registerHealthRoutes(router, database, logger)
    registerPublicRoutes(router, events, offerings, publicPurchases, publicPayments, limiter, logger, public.StorefrontAllowedOrigins)
    registerOperationsRoutes(router, operationsAuth, eventOperations, offeringOperations, purchaseOperations, paymentOperations)

    return router, nil
}

func isAPIPath(path string) bool {
    return strings.HasPrefix(path, "/api/")
}
