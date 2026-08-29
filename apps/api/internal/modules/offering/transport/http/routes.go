package http

import (
    "log/slog"
    "net/http"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/gin-gonic/gin"
)

func (handler *Handler) RegisterPublicRoutes(router gin.IRouter) {
    RegisterPublicRoutes(router, handler.service, handler.logger)
}

func RegisterPublicRoutes(router gin.IRouter, reader PublicCatalogueReader, logger *slog.Logger) {
    public := publicHandler{reader: reader, logger: logger}
    router.GET("/events/:event_id/offerings", public.list)
    router.GET("/offerings/:offering_id", public.get)
    router.OPTIONS("/events/:event_id/offerings", func(c *gin.Context) { c.Status(http.StatusNoContent) })
    router.OPTIONS("/offerings/:offering_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}

func (handler *Handler) RegisterOperationsRoutes(router gin.IRouter, authentication *platformauth.Service) {
    router.GET("/events/:event_id/offerings", authentication.Require("offering.read"), handler.list)
    router.GET("/offerings/:offering_id", authentication.Require("offering.read"), handler.get)
    router.POST("/events/:event_id/offerings", authentication.Require("offering.manage"), handler.create)
    router.PATCH("/offerings/:offering_id", authentication.Require("offering.manage"), handler.patch)
    router.POST("/offerings/:offering_id/publish", authentication.Require("offering.manage"), handler.transition(offeringdomain.ActionPublish, func(parent eventdomain.Event, current offeringdomain.Offering, input offeringdomain.TransitionInput, now time.Time) (offeringdomain.Mutation, error) {
        return offeringdomain.Publish(parent, current, offeringdomain.PublishInput{ExpectedVersion: input.ExpectedVersion, OccurredAt: now})
    }))
    router.POST("/offerings/:offering_id/unavailable", authentication.Require("offering.manage"), handler.transition(offeringdomain.ActionUnavailable, func(_ eventdomain.Event, current offeringdomain.Offering, input offeringdomain.TransitionInput, _ time.Time) (offeringdomain.Mutation, error) {
        return offeringdomain.MarkUnavailable(current, input)
    }))
    router.POST("/offerings/:offering_id/archive", authentication.Require("offering.manage"), handler.transition(offeringdomain.ActionArchive, func(_ eventdomain.Event, current offeringdomain.Offering, input offeringdomain.TransitionInput, _ time.Time) (offeringdomain.Mutation, error) {
        return offeringdomain.Archive(current, input)
    }))
    for _, path := range []string{
        "/events/:event_id/offerings", "/offerings/:offering_id", "/offerings/:offering_id/publish",
        "/offerings/:offering_id/unavailable", "/offerings/:offering_id/archive",
    } {
        router.OPTIONS(path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
    }
}
