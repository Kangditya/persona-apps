package event

import (
    "database/sql"
    "log/slog"

    eventapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/application"
    eventpersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/infrastructure/persistence"
    eventhttp "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/transport/http"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

type Module struct {
    service *eventapplication.Service
    http    *eventhttp.Handler
}

func NewModule(db *sql.DB, cipher idempotency.Cipher, logger *slog.Logger) *Module {
    service := eventapplication.NewService(db, func(database platformdb.DBTX) eventapplication.Repository {
        return eventpersistence.NewRepository(database)
    })
    return &Module{service: service, http: eventhttp.NewHandler(db, service, cipher, logger)}
}

func (module *Module) Service() *eventapplication.Service {
    return module.service
}

func (module *Module) RegisterPublicRoutes(router gin.IRouter) {
    module.http.RegisterPublicRoutes(router)
}

func (module *Module) RegisterOperationsRoutes(router gin.IRouter, authentication *platformauth.Service) {
    module.http.RegisterOperationsRoutes(router, authentication)
}
