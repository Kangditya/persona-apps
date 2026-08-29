package identity

import (
    identityapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/application"
    identitypersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/infrastructure/persistence"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

type Module struct {
    service *identityapplication.Service
}

func NewModule() *Module {
    service := identityapplication.NewService(func(database platformdb.DBTX) identityapplication.Repository {
        return identitypersistence.NewRepository(database)
    })
    return &Module{service: service}
}

func (module *Module) Service() *identityapplication.Service {
    return module.service
}
