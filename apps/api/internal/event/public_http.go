package event

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type ActiveReader interface {
    Active(context.Context) (*Event, error)
}

type publicEvent struct {
    ID                   string     `json:"id"`
    EventYear            int        `json:"event_year"`
    Name                 string     `json:"name"`
    Status               Status     `json:"status"`
    RegistrationOpensAt  *time.Time `json:"registration_opens_at,omitempty"`
    RegistrationClosesAt *time.Time `json:"registration_closes_at,omitempty"`
}

type publicEventResponse struct {
    Data *publicEvent `json:"data"`
}

func RegisterPublicRoutes(group *gin.RouterGroup, reader ActiveReader, logger *slog.Logger) {
    group.GET("/events/active", func(c *gin.Context) {
        if reader == nil {
            writeActiveEventError(c, logger, httpx.ErrDependencyUnavailable)
            return
        }
        current, err := reader.Active(c.Request.Context())
        if err != nil {
            writeActiveEventError(c, logger, err)
            return
        }
        if current == nil {
            c.JSON(http.StatusOK, publicEventResponse{})
            return
        }
        if current.Status != StatusActive {
            writeActiveEventError(c, logger, fmt.Errorf("unexpected public event status %q", current.Status))
            return
        }
        c.JSON(http.StatusOK, publicEventResponse{Data: toPublicEvent(*current)})
    })
}

func toPublicEvent(value Event) *publicEvent {
    return &publicEvent{
        ID:                   value.ID,
        EventYear:            value.EventYear,
        Name:                 value.Name,
        Status:               value.Status,
        RegistrationOpensAt:  copyTime(value.RegistrationOpensAt),
        RegistrationClosesAt: copyTime(value.RegistrationClosesAt),
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
    logger.Log(context.Background(), level, "public event query failed", "request_id", requestID, "error_type", errorType(err))
}

func errorType(err error) string {
    if err == nil {
        return ""
    }
    return fmt.Sprintf("%T", err)
}
