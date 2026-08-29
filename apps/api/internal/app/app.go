package app

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
)

const shutdownTimeout = 10 * time.Second

type App struct {
    server *http.Server
    db     *sql.DB
    logger *slog.Logger
}

func New(ctx context.Context, configuration config.APIConfig, logger *slog.Logger) (*App, error) {
    dependencies, db, err := newDependencies(ctx, configuration, logger)
    if err != nil {
        return nil, err
    }
    server, err := NewServer(configuration.HTTPAddress, logger, configuration.Public, dependencies)
    if err != nil {
        _ = db.Close()
        return nil, fmt.Errorf("create HTTP server: %w", err)
    }
    return &App{server: server, db: db, logger: logger}, nil
}

func (application *App) Run() error {
    defer application.db.Close()
    serverErrors := make(chan error, 1)
    go func() {
        application.logger.Info("API listening", "address", application.server.Addr)
        serverErrors <- application.server.ListenAndServe()
    }()

    signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    select {
    case err := <-serverErrors:
        if errors.Is(err, http.ErrServerClosed) {
            return nil
        }
        return fmt.Errorf("serve HTTP: %w", err)
    case <-signalContext.Done():
    }

    shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
    defer cancel()
    if err := application.server.Shutdown(shutdownContext); err != nil {
        return fmt.Errorf("shutdown HTTP server: %w", err)
    }
    return nil
}
