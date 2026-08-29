package persistence

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/jackc/pgx/v5/pgconn"
)

const partyColumns = `id, party_type, display_name, email, phone, created_at, updated_at`

type DBTX = platformdb.DBTX

type Repository struct {
    db DBTX
}

type partyScanner interface {
    Scan(...any) error
}

type partyRow struct {
    ID          string
    Type        string
    DisplayName string
    Email       sql.NullString
    Phone       sql.NullString
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func NewRepository(db DBTX) *Repository {
    return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, party identitydomain.Party) (identitydomain.Party, error) {
    created, err := scanParty(repository.db.QueryRowContext(ctx, `
INSERT INTO parties (party_type, display_name, email, phone)
VALUES ($1, $2, $3, $4)
RETURNING `+partyColumns, party.Type, party.DisplayName, party.Email, party.Phone))
    if err != nil {
        return identitydomain.Party{}, fmt.Errorf("insert party: %w", translateError(err))
    }
    return created, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (identitydomain.Party, error) {
    party, err := scanParty(repository.db.QueryRowContext(ctx, `SELECT `+partyColumns+` FROM parties WHERE id = $1`, id))
    if errors.Is(err, sql.ErrNoRows) {
        return identitydomain.Party{}, identitydomain.ErrNotFound
    }
    if err != nil {
        return identitydomain.Party{}, fmt.Errorf("get party: %w", err)
    }
    return party, nil
}

func scanParty(scanner partyScanner) (identitydomain.Party, error) {
    var row partyRow
    if err := scanner.Scan(&row.ID, &row.Type, &row.DisplayName, &row.Email, &row.Phone, &row.CreatedAt, &row.UpdatedAt); err != nil {
        return identitydomain.Party{}, err
    }
    return row.party(), nil
}

func (row partyRow) party() identitydomain.Party {
    party := identitydomain.Party{
        ID:          row.ID,
        Type:        identitydomain.PartyType(row.Type),
        DisplayName: row.DisplayName,
        CreatedAt:   row.CreatedAt.UTC(),
        UpdatedAt:   row.UpdatedAt.UTC(),
    }
    if row.Email.Valid {
        value := row.Email.String
        party.Email = &value
    }
    if row.Phone.Valid {
        value := row.Phone.String
        party.Phone = &value
    }
    return party
}

func translateError(err error) error {
    var postgresError *pgconn.PgError
    if errors.As(err, &postgresError) && postgresError.Code == "23514" {
        return identitydomain.ErrInvalidInput
    }
    return err
}
