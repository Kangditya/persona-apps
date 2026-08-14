package app

import (
    "net/http"
    "strings"

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/cors"
    "github.com/gin-gonic/gin"
)

func registerOperationsRoutes(router *gin.Engine, operationsAuth *auth.Service, eventOperations *event.OperationsHandler, offeringOperations *offering.OperationsHandler) {
    operations := router.Group(operationsAPIPrefix)
    if operationsAuth == nil {
        return
    }
    operations.Use(cors.Middleware(cors.Policy{
        Origins: operationsAuth.AllowedOrigins(), Methods: []string{http.MethodGet, http.MethodPost, http.MethodPatch},
        Headers: []string{"Content-Type", "X-CSRF-Token", "Idempotency-Key", "X-Request-ID"}, Credentials: true,
    }))
    operationsAuth.RegisterRoutes(operations)
    if eventOperations != nil && offeringOperations != nil {
        eventOperations.RegisterRoutes(operations, operationsAuth)
        offeringOperations.RegisterRoutes(operations, operationsAuth)
    }
    for _, path := range []string{
        "/auth/login", "/auth/callback", "/auth/session", "/auth/logout", "/events", "/events/:event_id",
        "/events/:event_id/publish", "/events/:event_id/activate", "/events/:event_id/suspend", "/events/:event_id/close",
        "/events/:event_id/archive", "/events/:event_id/offerings", "/offerings/:offering_id",
        "/offerings/:offering_id/publish", "/offerings/:offering_id/unavailable", "/offerings/:offering_id/archive",
    } {
        operations.OPTIONS(path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
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
