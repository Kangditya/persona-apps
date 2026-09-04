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
    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
    "github.com/Kangditya/persona-apps/apps/api/internal/payment"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/logger"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
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

    var store *payment.FilesystemStore
    if configuration.Evidence != nil {
        store, err = payment.NewFilesystemStore(configuration.Evidence.Root)
        if err != nil {
            return fmt.Errorf("create evidence storage: %w", err)
        }
        defer store.Close()
        startupContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
        defer cancel()
        if err := store.Reconcile(startupContext, payment.NewRepository(db)); err != nil {
            return fmt.Errorf("reconcile evidence storage: %w", err)
        }
    }

    var operationsAuth *auth.Service
    var publicPurchases *purchasing.PublicHandler
    var eventOperations *event.OperationsHandler
    var offeringOperations *offering.OperationsHandler
    var purchaseOperations *purchasing.OperationsHandler
    var publicPayments *payment.PublicHandler
    var paymentOperations *payment.OperationsHandler
    if configuration.Auth != nil {
        configuredAuth, authErr := auth.New(context.Background(), db, *configuration.Auth)
        if authErr != nil {
            return authErr
        }
        operationsAuth = configuredAuth
        cipher, cipherErr := idempotency.NewCipher(configuration.Auth.ResponseEncryptionKeys)
        if cipherErr != nil {
            return fmt.Errorf("create idempotency cipher: %w", cipherErr)
        }
        publicPurchases = purchasing.NewPublicHandler(db, cipher, log)
        eventOperations = event.NewOperationsHandler(db, cipher, log)
        offeringOperations = offering.NewOperationsHandler(db, cipher, log)
        purchaseOperations = purchasing.NewOperationsHandler(db, log)
        if store != nil {
            publicPayments = payment.NewPublicHandler(db, cipher, log, store)
            paymentOperations = payment.NewOperationsHandler(db, store, log)
        }
    }
    server, err := app.NewServer(configuration.HTTPAddress, db, log, configuration.Public, publicPurchases, publicPayments, operationsAuth, eventOperations, offeringOperations, purchaseOperations, paymentOperations)
    if err != nil {
        return fmt.Errorf("create HTTP server: %w", err)
    }
    serverErrors := make(chan error, 1)
    reconciliationContext, stopReconciliation := context.WithCancel(context.Background())
    reconciliationDone := make(chan struct{})
    defer func() {
        stopReconciliation()
        if store != nil {
            <-reconciliationDone
        }
    }()
    go func() {
        log.Info("API listening", "address", configuration.HTTPAddress)
        serverErrors <- server.ListenAndServe()
    }()
    if store != nil {
        go func() {
            defer close(reconciliationDone)
            reconcileEvidence(reconciliationContext, time.Hour, log, store, payment.NewRepository(db))
        }()
    } else {
        close(reconciliationDone)
    }

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

func reconcileEvidence(ctx context.Context, interval time.Duration, log *slog.Logger, store *payment.FilesystemStore, repository *payment.Repository) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
        }
        if err := store.Reconcile(ctx, repository); err != nil && log != nil {
            log.Warn("evidence reconciliation failed", "error_type", fmt.Sprintf("%T", err))
        }
    }
}
