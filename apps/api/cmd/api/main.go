package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/Kangditya/persona-apps/apps/api/internal/app"
    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    platformlogger "github.com/Kangditya/persona-apps/apps/api/internal/platform/logger"
)

func main() {
    log := platformlogger.New()
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
    application, err := app.New(context.Background(), configuration, log)
    if err != nil {
        return fmt.Errorf("create application: %w", err)
    }
    return application.Run()
}
