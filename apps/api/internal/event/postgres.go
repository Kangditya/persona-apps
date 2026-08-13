package event

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"

    "github.com/jackc/pgx/v5/pgconn"
)

const (
    DefaultListLimit = 50
    MaxListLimit     = 100
    maxCursorLength  = 512
    maxCursorPayload = 256
)

const eventColumns = `id, event_year, name, status, registration_opens_at, registration_closes_at, participant_quota, version, created_at, updated_at`

type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Repository struct {
    db DBTX
}

type ListInput struct {
    Cursor string
    Limit  int
}

type ListResult struct {
    Events     []Event
    NextCursor string
    Limit      int
}

type eventCursor struct {
    Version   int    `json:"v"`
    EventYear int    `json:"y"`
    ID        string `json:"i"`
}

type eventScanner interface {
    Scan(...any) error
}

func NewRepository(db DBTX) *Repository {
    return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, value Event) (Event, error) {
    created, err := scanEvent(repository.db.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status, registration_opens_at, registration_closes_at, participant_quota) VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+eventColumns, value.EventYear, value.Name, value.Status, value.RegistrationOpensAt, value.RegistrationClosesAt, value.ParticipantQuota))
    if err != nil {
        return Event{}, fmt.Errorf("insert event: %w", translateError(err))
    }
    return created, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (Event, error) {
    return repository.get(ctx, `SELECT `+eventColumns+` FROM qurban_events WHERE id = $1`, id)
}

func (repository *Repository) GetForUpdate(ctx context.Context, id string) (Event, error) {
    return repository.get(ctx, `SELECT `+eventColumns+` FROM qurban_events WHERE id = $1 FOR UPDATE`, id)
}

func (repository *Repository) Active(ctx context.Context) (*Event, error) {
    value, err := scanEvent(repository.db.QueryRowContext(ctx, `SELECT `+eventColumns+` FROM qurban_events WHERE status = $1 ORDER BY event_year DESC, id ASC LIMIT 1`, StatusActive))
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("get active event: %w", err)
    }
    return &value, nil
}

func (repository *Repository) List(ctx context.Context, input ListInput) (ListResult, error) {
    limit, cursor, err := normalizeListInput(input)
    if err != nil {
        return ListResult{}, err
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
            return ListResult{}, ErrInvalidCursor
        }
        return ListResult{}, fmt.Errorf("list events: %w", err)
    }
    defer rows.Close()

    result := ListResult{Events: make([]Event, 0, limit), Limit: limit}
    for rows.Next() {
        value, scanErr := scanEvent(rows)
        if scanErr != nil {
            return ListResult{}, fmt.Errorf("scan event list row: %w", scanErr)
        }
        result.Events = append(result.Events, value)
    }
    if err := rows.Err(); err != nil {
        return ListResult{}, fmt.Errorf("iterate event list: %w", err)
    }
    if len(result.Events) > limit {
        last := result.Events[limit-1]
        next, encodeErr := encodeCursor(eventCursor{Version: 1, EventYear: last.EventYear, ID: last.ID})
        if encodeErr != nil {
            return ListResult{}, encodeErr
        }
        result.Events = result.Events[:limit]
        result.NextCursor = next
    }
    return result, nil
}

func (repository *Repository) Save(ctx context.Context, value Event, expectedVersion int64) (Event, error) {
    saved, err := scanEvent(repository.db.QueryRowContext(ctx, `UPDATE qurban_events SET name = $1, status = $2, registration_opens_at = $3, registration_closes_at = $4, participant_quota = $5, version = version + 1, updated_at = now() WHERE id = $6 AND version = $7 RETURNING `+eventColumns, value.Name, value.Status, value.RegistrationOpensAt, value.RegistrationClosesAt, value.ParticipantQuota, value.ID, expectedVersion))
    if errors.Is(err, sql.ErrNoRows) {
        return Event{}, repository.classifyMissingOrStale(ctx, value.ID)
    }
    if err != nil {
        return Event{}, fmt.Errorf("save event: %w", translateError(err))
    }
    return saved, nil
}

func (repository *Repository) get(ctx context.Context, query, id string) (Event, error) {
    value, err := scanEvent(repository.db.QueryRowContext(ctx, query, id))
    if errors.Is(err, sql.ErrNoRows) {
        return Event{}, ErrNotFound
    }
    if err != nil {
        return Event{}, fmt.Errorf("get event: %w", err)
    }
    return value, nil
}

func (repository *Repository) classifyMissingOrStale(ctx context.Context, id string) error {
    var found bool
    if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM qurban_events WHERE id = $1)`, id).Scan(&found); err != nil {
        return fmt.Errorf("classify event update: %w", err)
    }
    if !found {
        return ErrNotFound
    }
    return ErrStaleVersion
}

func scanEvent(scanner eventScanner) (Event, error) {
    var value Event
    var status string
    var opensAt sql.NullTime
    var closesAt sql.NullTime
    var quota sql.NullInt64
    if err := scanner.Scan(&value.ID, &value.EventYear, &value.Name, &status, &opensAt, &closesAt, &quota, &value.Version, &value.CreatedAt, &value.UpdatedAt); err != nil {
        return Event{}, err
    }
    value.Status = Status(status)
    if opensAt.Valid {
        opened := opensAt.Time.UTC()
        value.RegistrationOpensAt = &opened
    }
    if closesAt.Valid {
        closed := closesAt.Time.UTC()
        value.RegistrationClosesAt = &closed
    }
    if quota.Valid {
        participantQuota := quota.Int64
        value.ParticipantQuota = &participantQuota
    }
    value.CreatedAt = value.CreatedAt.UTC()
    value.UpdatedAt = value.UpdatedAt.UTC()
    return value, nil
}

func normalizeListInput(input ListInput) (int, *eventCursor, error) {
    limit := input.Limit
    if limit == 0 {
        limit = DefaultListLimit
    }
    if limit < 1 || limit > MaxListLimit {
        return 0, nil, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidInput, MaxListLimit)
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
        return eventCursor{}, ErrInvalidCursor
    }
    payload, err := base64.RawURLEncoding.DecodeString(value)
    if err != nil || len(payload) == 0 || len(payload) > maxCursorPayload {
        return eventCursor{}, ErrInvalidCursor
    }
    decoder := json.NewDecoder(bytes.NewReader(payload))
    decoder.DisallowUnknownFields()
    var cursor eventCursor
    if err := decoder.Decode(&cursor); err != nil {
        return eventCursor{}, ErrInvalidCursor
    }
    if err := decoder.Decode(&struct{}{}); err != io.EOF {
        return eventCursor{}, ErrInvalidCursor
    }
    if cursor.Version != 1 || cursor.EventYear < MinEventYear || cursor.EventYear > MaxEventYear || len(cursor.ID) != 36 {
        return eventCursor{}, ErrInvalidCursor
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
            return ErrDuplicateYear
        case "uq_qurban_events_one_active":
            return ErrActiveConflict
        }
    case "23514":
        return ErrInvalidInput
    }
    return err
}

func isInvalidCursorQuery(err error) bool {
    var postgresError *pgconn.PgError
    return errors.As(err, &postgresError) && postgresError.Code == "22P02"
}
