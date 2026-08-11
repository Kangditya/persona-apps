package app

import (
    "log/slog"
    "net/http"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

const (
    publicAPIPrefix     = "/api/public/v1"
    operationsAPIPrefix = "/api/operations/v1"
)

func newRouter(database readinessChecker, logger *slog.Logger, operationsAuth *auth.Service) *gin.Engine {
    router := gin.New()
    router.RedirectTrailingSlash = false
    router.RedirectFixedPath = false
    router.HandleMethodNotAllowed = true
    router.NoRoute(func(c *gin.Context) {
        http.NotFound(c.Writer, c.Request)
    })
    router.NoMethod(func(c *gin.Context) {
        http.Error(c.Writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
    })
    router.Use(httpx.RequestIDMiddleware(logger))

    registerHealthRoutes(router, database, logger)
    registerPublicRoutes(router)
    registerOperationsRoutes(router, operationsAuth)

    return router
}
