package purchasing

import (
    "database/sql"
    "log/slog"

    identityapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/application"
    purchasingapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/application"
    purchasingpersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/infrastructure/persistence"
    purchasinghttp "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/transport/http"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

type Module struct {
    service *purchasingapplication.Service
    http    *purchasinghttp.Handler
}

func NewModule(db *sql.DB, parties *identityapplication.Service, cipher idempotency.Cipher, logger *slog.Logger) *Module {
    service := purchasingapplication.NewService(
        db,
        func(database platformdb.DBTX) purchasingapplication.Repository {
            return purchasingpersistence.NewRepository(database)
        },
        parties,
        purchasingpersistence.LockCheckoutForOffering,
        purchasingpersistence.Reserve,
    )
    return &Module{service: service, http: purchasinghttp.NewHandler(db, service, cipher, logger)}
}

func (module *Module) Service() *purchasingapplication.Service {
    return module.service
}

func (module *Module) RegisterPublicRoutes(router gin.IRouter) {
    module.http.RegisterPublicRoutes(router)
}

func (module *Module) RegisterOperationsRoutes(router gin.IRouter, authentication *platformauth.Service) {
    module.http.RegisterOperationsRoutes(router, authentication)
}
