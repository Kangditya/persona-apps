package app

import (
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/gin-gonic/gin"
)

func registerOperationsRoutes(router *gin.Engine, operationsAuth *auth.Service) {
    operations := router.Group(operationsAPIPrefix)
    if operationsAuth != nil {
        operationsAuth.RegisterRoutes(operations)
    }
}
