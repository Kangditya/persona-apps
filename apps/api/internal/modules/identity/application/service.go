package application

import (
    "context"
    "database/sql"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

type Repository interface {
    Create(context.Context, identitydomain.Party) (identitydomain.Party, error)
    Get(context.Context, string) (identitydomain.Party, error)
}

type RepositoryFactory func(platformdb.DBTX) Repository

type Service struct {
    repositories RepositoryFactory
}

func NewService(repositories RepositoryFactory) *Service {
    return &Service{repositories: repositories}
}

func (service *Service) CreateInTransaction(ctx context.Context, tx *sql.Tx, party identitydomain.Party) (identitydomain.Party, error) {
    return service.repositories(tx).Create(ctx, party)
}

func (service *Service) Get(ctx context.Context, db platformdb.DBTX, id string) (identitydomain.Party, error) {
    return service.repositories(db).Get(ctx, id)
}
