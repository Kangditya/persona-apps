package app

import (
    "context"
    "log/slog"
    "net/http"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
    "github.com/Kangditya/persona-apps/apps/api/internal/payment"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/purchasing"
)

const readinessTimeout = 2 * time.Second

type readinessChecker interface {
    PingContext(context.Context) error
}

func NewServer(address string, database readinessChecker, logger *slog.Logger, public config.PublicConfig, publicPurchases *purchasing.PublicHandler, publicPayments *payment.PublicHandler, operationsAuth *auth.Service, eventOperations *event.OperationsHandler, offeringOperations *offering.OperationsHandler, purchaseOperations *purchasing.OperationsHandler, paymentOperations *payment.OperationsHandler) (*http.Server, error) {
    events, offerings := publicReaders(database)
    router, err := newRouter(database, logger, public, publicPurchases, publicPayments, operationsAuth, events, offerings, eventOperations, offeringOperations, purchaseOperations, paymentOperations)
    if err != nil {
        return nil, err
    }
    return &http.Server{
        Addr:              address,
        Handler:           router,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       60 * time.Second,
        ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
    }, nil
}

func publicReaders(database readinessChecker) (event.ActiveReader, offering.PublicCatalogueReader) {
    queries, ok := database.(event.DBTX)
    if !ok {
        return nil, nil
    }
    return event.NewRepository(queries), offering.NewRepository(queries)
}
