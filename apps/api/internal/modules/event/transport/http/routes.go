package http

import (
    "log/slog"
    "net/http"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/gin-gonic/gin"
)

func (handler *Handler) RegisterPublicRoutes(router gin.IRouter) {
    RegisterPublicRoutes(router, handler.service, handler.logger)
}

func RegisterPublicRoutes(router gin.IRouter, reader ActiveReader, logger *slog.Logger) {
    public := publicHandler{reader: reader, logger: logger}
    router.GET("/events/active", public.active)
    router.OPTIONS("/events/active", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}

func (handler *Handler) RegisterOperationsRoutes(router gin.IRouter, authentication *platformauth.Service) {
    router.GET("/events", authentication.Require("event.read"), handler.list)
    router.GET("/events/:event_id", authentication.Require("event.read"), handler.get)
    router.POST("/events", authentication.Require("event.manage"), handler.create)
    router.PATCH("/events/:event_id", authentication.Require("event.manage"), handler.patch)
    router.POST("/events/:event_id/publish", authentication.Require("event.manage"), handler.transition(eventdomain.ActionPublish, eventdomain.Publish))
    router.POST("/events/:event_id/activate", authentication.Require("event.manage"), handler.transition(eventdomain.ActionActivate, eventdomain.Activate))
    router.POST("/events/:event_id/suspend", authentication.Require("event.manage"), handler.transition(eventdomain.ActionSuspend, eventdomain.Suspend))
    router.POST("/events/:event_id/close", authentication.Require("event.manage"), handler.transition(eventdomain.ActionClose, eventdomain.Close))
    router.POST("/events/:event_id/archive", authentication.Require("event.manage"), handler.transition(eventdomain.ActionArchive, eventdomain.Archive))

    for _, path := range []string{
        "/events", "/events/:event_id", "/events/:event_id/publish", "/events/:event_id/activate",
        "/events/:event_id/suspend", "/events/:event_id/close", "/events/:event_id/archive",
    } {
        router.OPTIONS(path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
    }
}
