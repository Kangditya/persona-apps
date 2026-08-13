package offering

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "unicode/utf8"

    "github.com/jackc/pgx/v5/pgconn"
)

const (
    DefaultListLimit = 50
    MaxListLimit     = 100
    maxCursorLength  = 512
    maxCursorPayload = 256
)

const offeringColumns = `id, event_id, code, name, offering_kind, description, price_minor, currency_code, participant_capacity, participant_quota, status, published_at, version, created_at, updated_at`

const catalogueOfferingColumns = `offerings.id, offerings.event_id, offerings.code, offerings.name, offerings.offering_kind, offerings.description, offerings.price_minor, offerings.currency_code, offerings.participant_capacity, offerings.participant_quota, offerings.status, offerings.published_at, offerings.version, offerings.created_at, offerings.updated_at`

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
    Offerings  []Offering
    NextCursor string
    Limit      int
}

type CatalogueOffering struct {
    Offering                  Offering
    AvailableParticipantUnits *int64
}

type CatalogueResult struct {
    Offerings  []CatalogueOffering
    NextCursor string
    Limit      int
}

type offeringCursor struct {
    Version int    `json:"v"`
    Code    string `json:"c"`
    ID      string `json:"i"`
}

type offeringScanner interface {
    Scan(...any) error
}

type offeringRow struct {
    ID                  string
    EventID             string
    Code                string
    Name                string
    Kind                string
    Description         sql.NullString
    PriceMinor          int64
    CurrencyCode        string
    ParticipantCapacity int32
    ParticipantQuota    sql.NullInt64
    Status              string
    PublishedAt         sql.NullTime
    Version             int64
    CreatedAt           sql.NullTime
    UpdatedAt           sql.NullTime
}

func NewRepository(db DBTX) *Repository {
    return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, value Offering) (Offering, error) {
    created, err := scanOffering(repository.db.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, description, price_minor, currency_code, participant_capacity, participant_quota, status, published_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING `+offeringColumns, value.EventID, value.Code, value.Name, value.Kind, value.Description, value.PriceMinor, value.CurrencyCode, value.ParticipantCapacity, value.ParticipantQuota, value.Status, value.PublishedAt))
    if err != nil {
        return Offering{}, fmt.Errorf("insert offering: %w", translateError(err))
    }
    return created, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (Offering, error) {
    return repository.get(ctx, `SELECT `+offeringColumns+` FROM offerings WHERE id = $1`, id)
}

func (repository *Repository) GetForUpdate(ctx context.Context, id string) (Offering, error) {
    return repository.get(ctx, `SELECT `+offeringColumns+` FROM offerings WHERE id = $1 FOR UPDATE`, id)
}

func (repository *Repository) List(ctx context.Context, eventID string, input ListInput) (ListResult, error) {
    limit, cursor, err := normalizeListInput(input)
    if err != nil {
        return ListResult{}, err
    }
    query := `SELECT ` + offeringColumns + ` FROM offerings WHERE event_id = $1 ORDER BY code ASC, id ASC LIMIT $2`
    arguments := []any{eventID, limit + 1}
    if cursor != nil {
        query = `SELECT ` + offeringColumns + ` FROM offerings WHERE event_id = $1 AND (code > $2 OR (code = $2 AND id > $3)) ORDER BY code ASC, id ASC LIMIT $4`
        arguments = []any{eventID, cursor.Code, cursor.ID, limit + 1}
    }
    rows, err := repository.db.QueryContext(ctx, query, arguments...)
    if err != nil {
        if cursor != nil && isInvalidCursorQuery(err) {
            return ListResult{}, ErrInvalidCursor
        }
        return ListResult{}, fmt.Errorf("list offerings: %w", err)
    }
    defer rows.Close()

    result := ListResult{Offerings: make([]Offering, 0, limit), Limit: limit}
    for rows.Next() {
        value, scanErr := scanOffering(rows)
        if scanErr != nil {
            return ListResult{}, fmt.Errorf("scan offering list row: %w", scanErr)
        }
        result.Offerings = append(result.Offerings, value)
    }
    if err := rows.Err(); err != nil {
        return ListResult{}, fmt.Errorf("iterate offering list: %w", err)
    }
    if len(result.Offerings) > limit {
        last := result.Offerings[limit-1]
        next, encodeErr := encodeCursor(offeringCursor{Version: 1, Code: last.Code, ID: last.ID})
        if encodeErr != nil {
            return ListResult{}, encodeErr
        }
        result.Offerings = result.Offerings[:limit]
        result.NextCursor = next
    }
    return result, nil
}

func (repository *Repository) Save(ctx context.Context, value Offering, expectedVersion int64) (Offering, error) {
    saved, err := scanOffering(repository.db.QueryRowContext(ctx, `UPDATE offerings SET name = $1, description = $2, price_minor = $3, participant_quota = $4, status = $5, published_at = $6, version = version + 1, updated_at = now() WHERE id = $7 AND version = $8 RETURNING `+offeringColumns, value.Name, value.Description, value.PriceMinor, value.ParticipantQuota, value.Status, value.PublishedAt, value.ID, expectedVersion))
    if errors.Is(err, sql.ErrNoRows) {
        return Offering{}, repository.classifyMissingOrStale(ctx, value.ID)
    }
    if err != nil {
        return Offering{}, fmt.Errorf("save offering: %w", translateError(err))
    }
    return saved, nil
}

func (repository *Repository) ListPublic(ctx context.Context, eventID string, input ListInput) (CatalogueResult, error) {
    limit, cursor, err := normalizeListInput(input)
    if err != nil {
        return CatalogueResult{}, err
    }
    if err := repository.requireActiveEvent(ctx, eventID); err != nil {
        return CatalogueResult{}, err
    }
    query := catalogueListQuery
    arguments := []any{eventID, limit + 1}
    if cursor != nil {
        query = catalogueListAfterCursorQuery
        arguments = []any{eventID, cursor.Code, cursor.ID, limit + 1}
    }
    rows, err := repository.db.QueryContext(ctx, query, arguments...)
    if err != nil {
        if cursor != nil && isInvalidCursorQuery(err) {
            return CatalogueResult{}, ErrInvalidCursor
        }
        return CatalogueResult{}, fmt.Errorf("list public offerings: %w", err)
    }
    defer rows.Close()

    result := CatalogueResult{Offerings: make([]CatalogueOffering, 0, limit), Limit: limit}
    for rows.Next() {
        value, scanErr := scanCatalogueOffering(rows)
        if scanErr != nil {
            return CatalogueResult{}, fmt.Errorf("scan public offering list row: %w", scanErr)
        }
        result.Offerings = append(result.Offerings, value)
    }
    if err := rows.Err(); err != nil {
        return CatalogueResult{}, fmt.Errorf("iterate public offering list: %w", err)
    }
    if len(result.Offerings) > limit {
        last := result.Offerings[limit-1].Offering
        next, encodeErr := encodeCursor(offeringCursor{Version: 1, Code: last.Code, ID: last.ID})
        if encodeErr != nil {
            return CatalogueResult{}, encodeErr
        }
        result.Offerings = result.Offerings[:limit]
        result.NextCursor = next
    }
    return result, nil
}

func (repository *Repository) GetPublic(ctx context.Context, id string) (CatalogueOffering, error) {
    value, err := scanCatalogueOffering(repository.db.QueryRowContext(ctx, catalogueDetailQuery, id))
    if errors.Is(err, sql.ErrNoRows) {
        return CatalogueOffering{}, ErrNotFound
    }
    if err != nil {
        return CatalogueOffering{}, fmt.Errorf("get public offering: %w", err)
    }
    return value, nil
}

func (repository *Repository) get(ctx context.Context, query, id string) (Offering, error) {
    value, err := scanOffering(repository.db.QueryRowContext(ctx, query, id))
    if errors.Is(err, sql.ErrNoRows) {
        return Offering{}, ErrNotFound
    }
    if err != nil {
        return Offering{}, fmt.Errorf("get offering: %w", err)
    }
    return value, nil
}

func (repository *Repository) requireActiveEvent(ctx context.Context, eventID string) error {
    var active bool
    if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM qurban_events WHERE id = $1 AND status = 'ACTIVE')`, eventID).Scan(&active); err != nil {
        return fmt.Errorf("check public event: %w", err)
    }
    if !active {
        return ErrNotFound
    }
    return nil
}

func (repository *Repository) classifyMissingOrStale(ctx context.Context, id string) error {
    var found bool
    if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM offerings WHERE id = $1)`, id).Scan(&found); err != nil {
        return fmt.Errorf("classify offering update: %w", err)
    }
    if !found {
        return ErrNotFound
    }
    return ErrStaleVersion
}

func scanOffering(scanner offeringScanner) (Offering, error) {
    var row offeringRow
    if err := scanner.Scan(row.destinations()...); err != nil {
        return Offering{}, err
    }
    return row.offering(), nil
}

func scanCatalogueOffering(scanner offeringScanner) (CatalogueOffering, error) {
    var row offeringRow
    var eventQuota sql.NullInt64
    var eventReserved int64
    var eventConsumed int64
    var offeringReserved int64
    var offeringConsumed int64
    destinations := row.destinations()
    destinations = append(destinations, &eventQuota, &eventReserved, &eventConsumed, &offeringReserved, &offeringConsumed)
    if err := scanner.Scan(destinations...); err != nil {
        return CatalogueOffering{}, err
    }
    value := row.offering()
    availability, err := AvailableParticipantUnits(AvailabilityInput{
        EventQuota:    nullableInt64(eventQuota),
        OfferingQuota: value.ParticipantQuota,
        EventUsage:    ReservationUsage{Reserved: eventReserved, Consumed: eventConsumed},
        OfferingUsage: ReservationUsage{Reserved: offeringReserved, Consumed: offeringConsumed},
    })
    if err != nil {
        return CatalogueOffering{}, err
    }
    return CatalogueOffering{Offering: value, AvailableParticipantUnits: availability}, nil
}

func (row *offeringRow) destinations() []any {
    return []any{
        &row.ID,
        &row.EventID,
        &row.Code,
        &row.Name,
        &row.Kind,
        &row.Description,
        &row.PriceMinor,
        &row.CurrencyCode,
        &row.ParticipantCapacity,
        &row.ParticipantQuota,
        &row.Status,
        &row.PublishedAt,
        &row.Version,
        &row.CreatedAt,
        &row.UpdatedAt,
    }
}

func (row offeringRow) offering() Offering {
    value := Offering{
        ID:                  row.ID,
        EventID:             row.EventID,
        Code:                row.Code,
        Name:                row.Name,
        Kind:                row.Kind,
        PriceMinor:          row.PriceMinor,
        CurrencyCode:        row.CurrencyCode,
        ParticipantCapacity: row.ParticipantCapacity,
        Status:              Status(row.Status),
        Version:             row.Version,
    }
    if row.Description.Valid {
        description := row.Description.String
        value.Description = &description
    }
    value.ParticipantQuota = nullableInt64(row.ParticipantQuota)
    if row.PublishedAt.Valid {
        publishedAt := row.PublishedAt.Time.UTC()
        value.PublishedAt = &publishedAt
    }
    if row.CreatedAt.Valid {
        value.CreatedAt = row.CreatedAt.Time.UTC()
    }
    if row.UpdatedAt.Valid {
        value.UpdatedAt = row.UpdatedAt.Time.UTC()
    }
    return value
}

func nullableInt64(value sql.NullInt64) *int64 {
    if !value.Valid {
        return nil
    }
    result := value.Int64
    return &result
}

func normalizeListInput(input ListInput) (int, *offeringCursor, error) {
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

func encodeCursor(cursor offeringCursor) (string, error) {
    payload, err := json.Marshal(cursor)
    if err != nil {
        return "", fmt.Errorf("encode offering cursor: %w", err)
    }
    return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeCursor(value string) (offeringCursor, error) {
    if len(value) == 0 || len(value) > maxCursorLength {
        return offeringCursor{}, ErrInvalidCursor
    }
    payload, err := base64.RawURLEncoding.DecodeString(value)
    if err != nil || len(payload) == 0 || len(payload) > maxCursorPayload {
        return offeringCursor{}, ErrInvalidCursor
    }
    decoder := json.NewDecoder(bytes.NewReader(payload))
    decoder.DisallowUnknownFields()
    var cursor offeringCursor
    if err := decoder.Decode(&cursor); err != nil {
        return offeringCursor{}, ErrInvalidCursor
    }
    if err := decoder.Decode(&struct{}{}); err != io.EOF {
        return offeringCursor{}, ErrInvalidCursor
    }
    if cursor.Version != 1 || cursor.Code == "" || utf8.RuneCountInString(cursor.Code) > MaxCodeLength || len(cursor.ID) != 36 {
        return offeringCursor{}, ErrInvalidCursor
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
        if postgresError.ConstraintName == "offerings_event_id_code_key" {
            return ErrDuplicateCode
        }
    case "23503":
        if postgresError.ConstraintName == "offerings_event_id_fkey" {
            return ErrInvalidParentState
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

const catalogueListQuery = `WITH event_usage AS (
    SELECT
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'RESERVED'), 0)::bigint AS reserved_units,
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'CONSUMED'), 0)::bigint AS consumed_units
    FROM quota_reservations
    WHERE event_id = $1 AND status IN ('RESERVED', 'CONSUMED')
), offering_usage AS (
    SELECT
        offering_id,
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'RESERVED'), 0)::bigint AS reserved_units,
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'CONSUMED'), 0)::bigint AS consumed_units
    FROM quota_reservations
    WHERE event_id = $1 AND status IN ('RESERVED', 'CONSUMED')
    GROUP BY offering_id
)
SELECT ` + catalogueOfferingColumns + `,
    event.participant_quota,
    event_usage.reserved_units,
    event_usage.consumed_units,
    COALESCE(offering_usage.reserved_units, 0)::bigint,
    COALESCE(offering_usage.consumed_units, 0)::bigint
FROM offerings
JOIN qurban_events AS event ON event.id = offerings.event_id AND event.status = 'ACTIVE'
CROSS JOIN event_usage
LEFT JOIN offering_usage ON offering_usage.offering_id = offerings.id
WHERE offerings.event_id = $1 AND offerings.status = 'PUBLISHED'
ORDER BY offerings.code ASC, offerings.id ASC
LIMIT $2`

const catalogueListAfterCursorQuery = `WITH event_usage AS (
    SELECT
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'RESERVED'), 0)::bigint AS reserved_units,
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'CONSUMED'), 0)::bigint AS consumed_units
    FROM quota_reservations
    WHERE event_id = $1 AND status IN ('RESERVED', 'CONSUMED')
), offering_usage AS (
    SELECT
        offering_id,
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'RESERVED'), 0)::bigint AS reserved_units,
        COALESCE(SUM(participant_units) FILTER (WHERE status = 'CONSUMED'), 0)::bigint AS consumed_units
    FROM quota_reservations
    WHERE event_id = $1 AND status IN ('RESERVED', 'CONSUMED')
    GROUP BY offering_id
)
SELECT ` + catalogueOfferingColumns + `,
    event.participant_quota,
    event_usage.reserved_units,
    event_usage.consumed_units,
    COALESCE(offering_usage.reserved_units, 0)::bigint,
    COALESCE(offering_usage.consumed_units, 0)::bigint
FROM offerings
JOIN qurban_events AS event ON event.id = offerings.event_id AND event.status = 'ACTIVE'
CROSS JOIN event_usage
LEFT JOIN offering_usage ON offering_usage.offering_id = offerings.id
WHERE offerings.event_id = $1
    AND offerings.status = 'PUBLISHED'
    AND (offerings.code > $2 OR (offerings.code = $2 AND offerings.id > $3))
ORDER BY offerings.code ASC, offerings.id ASC
LIMIT $4`

const catalogueDetailQuery = `SELECT ` + catalogueOfferingColumns + `,
    event.participant_quota,
    COALESCE((
        SELECT SUM(participant_units) FILTER (WHERE status = 'RESERVED')
        FROM quota_reservations
        WHERE event_id = offerings.event_id AND status IN ('RESERVED', 'CONSUMED')
    ), 0)::bigint,
    COALESCE((
        SELECT SUM(participant_units) FILTER (WHERE status = 'CONSUMED')
        FROM quota_reservations
        WHERE event_id = offerings.event_id AND status IN ('RESERVED', 'CONSUMED')
    ), 0)::bigint,
    COALESCE((
        SELECT SUM(participant_units) FILTER (WHERE status = 'RESERVED')
        FROM quota_reservations
        WHERE offering_id = offerings.id AND status IN ('RESERVED', 'CONSUMED')
    ), 0)::bigint,
    COALESCE((
        SELECT SUM(participant_units) FILTER (WHERE status = 'CONSUMED')
        FROM quota_reservations
        WHERE offering_id = offerings.id AND status IN ('RESERVED', 'CONSUMED')
    ), 0)::bigint
FROM offerings
JOIN qurban_events AS event ON event.id = offerings.event_id AND event.status = 'ACTIVE'
WHERE offerings.id = $1 AND offerings.status = 'PUBLISHED'`
