package offering

import (
    "database/sql"
    "log/slog"

    eventapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/application"
    offeringapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/application"
    offeringpersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/infrastructure/persistence"
    offeringhttp "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/transport/http"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

type Module struct {
    service *offeringapplication.Service
    http    *offeringhttp.Handler
}

func NewModule(db *sql.DB, events *eventapplication.Service, cipher idempotency.Cipher, logger *slog.Logger) *Module {
    service := offeringapplication.NewService(db, func(database platformdb.DBTX) offeringapplication.Repository {
        return offeringpersistence.NewRepository(database)
    }, events)
    return &Module{service: service, http: offeringhttp.NewHandler(db, service, cipher, logger)}
}

func (module *Module) Service() *offeringapplication.Service {
    return module.service
}

func (module *Module) RegisterPublicRoutes(router gin.IRouter) {
    module.http.RegisterPublicRoutes(router)
}

func (module *Module) RegisterOperationsRoutes(router gin.IRouter, authentication *platformauth.Service) {
    module.http.RegisterOperationsRoutes(router, authentication)
}
