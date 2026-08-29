package http

import (
    "net/http"

    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/gin-gonic/gin"
)

func (handler *Handler) RegisterPublicRoutes(router gin.IRouter) {
    router.POST("/purchases", handler.create)
    router.OPTIONS("/purchases", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}

func (handler *Handler) RegisterOperationsRoutes(router gin.IRouter, authentication *platformauth.Service) {
    router.GET("/purchases", authentication.Require("purchase.read"), handler.list)
    router.GET("/purchases/:purchase_id", authentication.Require("purchase.read"), handler.get)
    router.OPTIONS("/purchases", func(c *gin.Context) { c.Status(http.StatusNoContent) })
    router.OPTIONS("/purchases/:purchase_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}
