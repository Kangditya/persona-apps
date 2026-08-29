package persistence

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/jackc/pgx/v5/pgconn"
)

const (
    DefaultListLimit = 50
    MaxListLimit     = 100
    maxCursorLength  = 512
    maxCursorPayload = 256
)

const eventColumns = `id, event_year, name, status, registration_opens_at, registration_closes_at, participant_quota, version, created_at, updated_at`

type DBTX = platformdb.DBTX

type Repository struct {
    db DBTX
}

type eventCursor struct {
    Version   int    `json:"v"`
    EventYear int    `json:"y"`
    ID        string `json:"i"`
}

type eventScanner interface {
    Scan(...any) error
}

type eventRow struct {
    ID                   string
    EventYear            int
    Name                 string
    Status               string
    RegistrationOpensAt  sql.NullTime
    RegistrationClosesAt sql.NullTime
    ParticipantQuota     sql.NullInt64
    Version              int64
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

func NewRepository(db DBTX) *Repository {
    return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, value eventdomain.Event) (eventdomain.Event, error) {
    created, err := scanEvent(repository.db.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status, registration_opens_at, registration_closes_at, participant_quota) VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+eventColumns, value.EventYear, value.Name, value.Status, value.RegistrationOpensAt, value.RegistrationClosesAt, value.ParticipantQuota))
    if err != nil {
        return eventdomain.Event{}, fmt.Errorf("insert event: %w", translateError(err))
    }
    return created, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (eventdomain.Event, error) {
    return repository.get(ctx, `SELECT `+eventColumns+` FROM qurban_events WHERE id = $1`, id)
}

func (repository *Repository) GetForUpdate(ctx context.Context, id string) (eventdomain.Event, error) {
    return repository.get(ctx, `SELECT `+eventColumns+` FROM qurban_events WHERE id = $1 FOR UPDATE`, id)
}

func (repository *Repository) Active(ctx context.Context) (*eventdomain.Event, error) {
    value, err := scanEvent(repository.db.QueryRowContext(ctx, `SELECT `+eventColumns+` FROM qurban_events WHERE status = $1 ORDER BY event_year DESC, id ASC LIMIT 1`, eventdomain.StatusActive))
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("get active event: %w", err)
    }
    return &value, nil
}

func (repository *Repository) List(ctx context.Context, input eventdomain.ListInput) (eventdomain.ListResult, error) {
    limit, cursor, err := normalizeListInput(input)
    if err != nil {
        return eventdomain.ListResult{}, err
    }
    query := `SELECT ` + eventColumns + ` FROM qurban_events ORDER BY event_year DESC, id ASC LIMIT $1`
    arguments := []any{limit + 1}
    if cursor != nil {
        query = `SELECT ` + eventColumns + ` FROM qurban_events WHERE event_year < $1 OR (event_year = $1 AND id > $2) ORDER BY event_year DESC, id ASC LIMIT $3`
        arguments = []any{cursor.EventYear, cursor.ID, limit + 1}
    }
    rows, err := repository.db.QueryContext(ctx, query, arguments...)
    if err != nil {
        if cursor != nil && isInvalidCursorQuery(err) {
            return eventdomain.ListResult{}, eventdomain.ErrInvalidCursor
        }
        return eventdomain.ListResult{}, fmt.Errorf("list events: %w", err)
    }
    defer rows.Close()

    result := eventdomain.ListResult{Events: make([]eventdomain.Event, 0, limit), Limit: limit}
    for rows.Next() {
        value, scanErr := scanEvent(rows)
        if scanErr != nil {
            return eventdomain.ListResult{}, fmt.Errorf("scan event list row: %w", scanErr)
        }
        result.Events = append(result.Events, value)
    }
    if err := rows.Err(); err != nil {
        return eventdomain.ListResult{}, fmt.Errorf("iterate event list: %w", err)
    }
    if len(result.Events) > limit {
        last := result.Events[limit-1]
        next, encodeErr := encodeCursor(eventCursor{Version: 1, EventYear: last.EventYear, ID: last.ID})
        if encodeErr != nil {
            return eventdomain.ListResult{}, encodeErr
        }
        result.Events = result.Events[:limit]
        result.NextCursor = next
    }
    return result, nil
}

func (repository *Repository) Save(ctx context.Context, value eventdomain.Event, expectedVersion int64) (eventdomain.Event, error) {
    saved, err := scanEvent(repository.db.QueryRowContext(ctx, `UPDATE qurban_events SET name = $1, status = $2, registration_opens_at = $3, registration_closes_at = $4, participant_quota = $5, version = version + 1, updated_at = now() WHERE id = $6 AND version = $7 RETURNING `+eventColumns, value.Name, value.Status, value.RegistrationOpensAt, value.RegistrationClosesAt, value.ParticipantQuota, value.ID, expectedVersion))
    if errors.Is(err, sql.ErrNoRows) {
        return eventdomain.Event{}, repository.classifyMissingOrStale(ctx, value.ID)
    }
    if err != nil {
        return eventdomain.Event{}, fmt.Errorf("save event: %w", translateError(err))
    }
    return saved, nil
}

func (repository *Repository) get(ctx context.Context, query, id string) (eventdomain.Event, error) {
    value, err := scanEvent(repository.db.QueryRowContext(ctx, query, id))
    if errors.Is(err, sql.ErrNoRows) {
        return eventdomain.Event{}, eventdomain.ErrNotFound
    }
    if err != nil {
        return eventdomain.Event{}, fmt.Errorf("get event: %w", err)
    }
    return value, nil
}

func (repository *Repository) classifyMissingOrStale(ctx context.Context, id string) error {
    var found bool
    if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM qurban_events WHERE id = $1)`, id).Scan(&found); err != nil {
        return fmt.Errorf("classify event update: %w", err)
    }
    if !found {
        return eventdomain.ErrNotFound
    }
    return eventdomain.ErrStaleVersion
}

func scanEvent(scanner eventScanner) (eventdomain.Event, error) {
    var row eventRow
    if err := scanner.Scan(
        &row.ID, &row.EventYear, &row.Name, &row.Status, &row.RegistrationOpensAt,
        &row.RegistrationClosesAt, &row.ParticipantQuota, &row.Version, &row.CreatedAt, &row.UpdatedAt,
    ); err != nil {
        return eventdomain.Event{}, err
    }
    return row.event(), nil
}

func (row eventRow) event() eventdomain.Event {
    value := eventdomain.Event{
        ID:        row.ID,
        EventYear: row.EventYear,
        Name:      row.Name,
        Status:    eventdomain.Status(row.Status),
        Version:   row.Version,
        CreatedAt: row.CreatedAt.UTC(),
        UpdatedAt: row.UpdatedAt.UTC(),
    }
    if row.RegistrationOpensAt.Valid {
        opened := row.RegistrationOpensAt.Time.UTC()
        value.RegistrationOpensAt = &opened
    }
    if row.RegistrationClosesAt.Valid {
        closed := row.RegistrationClosesAt.Time.UTC()
        value.RegistrationClosesAt = &closed
    }
    if row.ParticipantQuota.Valid {
        quota := row.ParticipantQuota.Int64
        value.ParticipantQuota = &quota
    }
    return value
}

func normalizeListInput(input eventdomain.ListInput) (int, *eventCursor, error) {
    limit := input.Limit
    if limit == 0 {
        limit = DefaultListLimit
    }
    if limit < 1 || limit > MaxListLimit {
        return 0, nil, fmt.Errorf("%w: limit must be between 1 and %d", eventdomain.ErrInvalidInput, MaxListLimit)
    }
    if input.Cursor == "" {
        return limit, nil, nil
    }
    cursor, err := decodeCursor(input.Cursor)
    if err != nil {
        return 0, nil, err
    }
    return limit, &cursor, nil
}

func encodeCursor(cursor eventCursor) (string, error) {
    payload, err := json.Marshal(cursor)
    if err != nil {
        return "", fmt.Errorf("encode event cursor: %w", err)
    }
    return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeCursor(value string) (eventCursor, error) {
    if len(value) == 0 || len(value) > maxCursorLength {
        return eventCursor{}, eventdomain.ErrInvalidCursor
    }
    payload, err := base64.RawURLEncoding.DecodeString(value)
    if err != nil || len(payload) == 0 || len(payload) > maxCursorPayload {
        return eventCursor{}, eventdomain.ErrInvalidCursor
    }
    decoder := json.NewDecoder(bytes.NewReader(payload))
    decoder.DisallowUnknownFields()
    var cursor eventCursor
    if err := decoder.Decode(&cursor); err != nil {
        return eventCursor{}, eventdomain.ErrInvalidCursor
    }
    if err := decoder.Decode(&struct{}{}); err != io.EOF {
        return eventCursor{}, eventdomain.ErrInvalidCursor
    }
    if cursor.Version != 1 || cursor.EventYear < eventdomain.MinEventYear || cursor.EventYear > eventdomain.MaxEventYear || len(cursor.ID) != 36 {
        return eventCursor{}, eventdomain.ErrInvalidCursor
    }
    return cursor, nil
}

func translateError(err error) error {
    var postgresError *pgconn.PgError
    if !errors.As(err, &postgresError) {
        return err
    }
    switch postgresError.Code {
    case "23505":
        switch postgresError.ConstraintName {
        case "qurban_events_event_year_key":
            return eventdomain.ErrDuplicateYear
        case "uq_qurban_events_one_active":
            return eventdomain.ErrActiveConflict
        }
    case "23514":
        return eventdomain.ErrInvalidInput
    }
    return err
}

func isInvalidCursorQuery(err error) bool {
    var postgresError *pgconn.PgError
    return errors.As(err, &postgresError) && postgresError.Code == "22P02"
}
