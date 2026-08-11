package app

import (
    "context"
    "log/slog"
    "net/http"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
)

const readinessTimeout = 2 * time.Second

type readinessChecker interface {
    PingContext(context.Context) error
}

func NewServer(address string, database readinessChecker, logger *slog.Logger, operationsAuth ...*auth.Service) *http.Server {
    var authService *auth.Service
    if len(operationsAuth) > 0 {
        authService = operationsAuth[0]
    }

    return &http.Server{
        Addr:              address,
        Handler:           newRouter(database, logger, authService),
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       60 * time.Second,
        ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
    }
}
