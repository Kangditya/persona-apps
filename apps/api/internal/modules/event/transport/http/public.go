package http

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type ActiveReader interface {
    Active(context.Context) (*eventdomain.Event, error)
}

type publicEvent struct {
    ID                   string             `json:"id"`
    EventYear            int                `json:"event_year"`
    Name                 string             `json:"name"`
    Status               eventdomain.Status `json:"status"`
    RegistrationOpensAt  *time.Time         `json:"registration_opens_at,omitempty"`
    RegistrationClosesAt *time.Time         `json:"registration_closes_at,omitempty"`
}

type publicEventResponse struct {
    Data *publicEvent `json:"data"`
}

type publicHandler struct {
    reader ActiveReader
    logger *slog.Logger
}

func (handler publicHandler) active(c *gin.Context) {
    if handler.reader == nil {
        writeActiveEventError(c, handler.logger, httpx.ErrDependencyUnavailable)
        return
    }
    current, err := handler.reader.Active(c.Request.Context())
    if err != nil {
        writeActiveEventError(c, handler.logger, err)
        return
    }
    if current == nil {
        c.JSON(http.StatusOK, publicEventResponse{})
        return
    }
    if current.Status != eventdomain.StatusActive {
        writeActiveEventError(c, handler.logger, fmt.Errorf("unexpected public event status %q", current.Status))
        return
    }
    c.JSON(http.StatusOK, publicEventResponse{Data: toPublicEvent(*current)})
}

func toPublicEvent(value eventdomain.Event) *publicEvent {
    return &publicEvent{
        ID: value.ID, EventYear: value.EventYear, Name: value.Name, Status: value.Status,
        RegistrationOpensAt: copyTime(value.RegistrationOpensAt), RegistrationClosesAt: copyTime(value.RegistrationClosesAt),
    }
}

func writeActiveEventError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    if httpx.IsDependencyUnavailable(err) {
        logPublicEventError(logger, slog.LevelWarn, requestID, err)
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
        return
    }
    logPublicEventError(logger, slog.LevelError, requestID, err)
    httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", requestID, nil)
}

func logPublicEventError(logger *slog.Logger, level slog.Level, requestID string, err error) {
    if logger == nil {
        return
    }
    logger.Log(context.Background(), level, "public event query failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
}
