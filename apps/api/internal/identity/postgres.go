package identity

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5/pgconn"
)

const partyColumns = `id, party_type, display_name, email, phone, created_at, updated_at`

type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Repository struct {
    db DBTX
}

type partyScanner interface {
    Scan(...any) error
}

func NewRepository(db DBTX) *Repository {
    return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, party Party) (Party, error) {
    created, err := scanParty(repository.db.QueryRowContext(ctx, `
INSERT INTO parties (party_type, display_name, email, phone)
VALUES ($1, $2, $3, $4)
RETURNING `+partyColumns, party.Type, party.DisplayName, party.Email, party.Phone))
    if err != nil {
        return Party{}, fmt.Errorf("insert party: %w", translateError(err))
    }
    return created, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (Party, error) {
    party, err := scanParty(repository.db.QueryRowContext(ctx, `SELECT `+partyColumns+` FROM parties WHERE id = $1`, id))
    if errors.Is(err, sql.ErrNoRows) {
        return Party{}, ErrNotFound
    }
    if err != nil {
        return Party{}, fmt.Errorf("get party: %w", err)
    }
    return party, nil
}

func scanParty(scanner partyScanner) (Party, error) {
    var party Party
    var partyType string
    var email sql.NullString
    var phone sql.NullString
    if err := scanner.Scan(&party.ID, &partyType, &party.DisplayName, &email, &phone, &party.CreatedAt, &party.UpdatedAt); err != nil {
        return Party{}, err
    }
    party.Type = PartyType(partyType)
    if email.Valid {
        value := email.String
        party.Email = &value
    }
    if phone.Valid {
        value := phone.String
        party.Phone = &value
    }
    party.CreatedAt = party.CreatedAt.UTC()
    party.UpdatedAt = party.UpdatedAt.UTC()
    return party, nil
}

func translateError(err error) error {
    var postgresError *pgconn.PgError
    if errors.As(err, &postgresError) && postgresError.Code == "23514" {
        return ErrInvalidInput
    }
    return err
}
