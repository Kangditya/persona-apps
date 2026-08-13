package main

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/app"
    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/logger"
)

const shutdownTimeout = 10 * time.Second

func main() {
    log := logger.New()
    if err := run(log); err != nil {
        log.Error("API stopped", "error", err)
        os.Exit(1)
    }
}

func run(log *slog.Logger) error {
    configuration, err := config.LoadAPI()
    if err != nil {
        return fmt.Errorf("load configuration: %w", err)
    }

    db, err := database.Open(configuration.DatabaseURL)
    if err != nil {
        return err
    }
    defer db.Close()

    var operationsAuth *auth.Service
    if configuration.Auth != nil {
        configuredAuth, authErr := auth.New(context.Background(), db, *configuration.Auth)
        if authErr != nil {
            return authErr
        }
        operationsAuth = configuredAuth
    }
    server, err := app.NewServer(configuration.HTTPAddress, db, log, configuration.Public, operationsAuth)
    if err != nil {
        return fmt.Errorf("create HTTP server: %w", err)
    }
    serverErrors := make(chan error, 1)
    go func() {
        log.Info("API listening", "address", configuration.HTTPAddress)
        serverErrors <- server.ListenAndServe()
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
    if err := server.Shutdown(shutdownContext); err != nil {
        return fmt.Errorf("shutdown HTTP server: %w", err)
    }

    return nil
}
