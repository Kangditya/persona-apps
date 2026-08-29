package app

import (
    "context"
    "database/sql"
    "log/slog"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    eventmodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/event"
    identitymodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity"
    offeringmodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering"
    purchasingmodule "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
)

type Dependencies struct {
    Database   readinessChecker
    Event      *eventmodule.Module
    Offering   *offeringmodule.Module
    Purchasing *purchasingmodule.Module
    Auth       *auth.Service
}

func newDependencies(ctx context.Context, configuration config.APIConfig, logger *slog.Logger) (Dependencies, *sql.DB, error) {
    db, err := platformdb.Open(configuration.DatabaseURL)
    if err != nil {
        return Dependencies{}, nil, err
    }
    var operationsAuth *auth.Service
    var cipher idempotency.Cipher
    if configuration.Auth != nil {
        operationsAuth, err = auth.New(ctx, db, *configuration.Auth)
        if err != nil {
            _ = db.Close()
            return Dependencies{}, nil, err
        }
        cipher, err = idempotency.NewCipher(configuration.Auth.ResponseEncryptionKeys)
        if err != nil {
            _ = db.Close()
            return Dependencies{}, nil, err
        }
    }
    events := eventmodule.NewModule(db, cipher, logger)
    identity := identitymodule.NewModule()
    dependencies := Dependencies{Database: db, Auth: operationsAuth, Event: events, Offering: offeringmodule.NewModule(db, events.Service(), cipher, logger)}
    if operationsAuth != nil {
        dependencies.Purchasing = purchasingmodule.NewModule(db, identity.Service(), cipher, logger)
    }
    return dependencies, db, nil
}
