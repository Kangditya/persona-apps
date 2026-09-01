package purchasing

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "strings"
    "time"

    "github.com/jackc/pgx/v5/pgconn"
)

const (
    DefaultListLimit = 50
    MaxListLimit     = 100
    maxCursorLength  = 512
    maxCursorPayload = 256
)

const purchaseColumns = `id, event_id, purchase_ref, channel, purchaser_party_id, payer_party_id, offering_id, offering_name_snapshot, offering_kind_snapshot, offering_unit_price_minor, participant_capacity_snapshot, participant_count, total_amount_minor, currency_code, status, version, created_at, updated_at`

const purchaseDetailColumns = `purchases.id, purchases.event_id, purchases.purchase_ref, purchases.channel, purchases.purchaser_party_id, purchases.payer_party_id, purchases.offering_id, purchases.offering_name_snapshot, purchases.offering_kind_snapshot, purchases.offering_unit_price_minor, purchases.participant_capacity_snapshot, purchases.participant_count, purchases.total_amount_minor, purchases.currency_code, purchases.status, purchases.version, purchases.created_at, purchases.updated_at, purchaser.id, purchaser.display_name, payer.id, payer.display_name`

type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Repository struct {
    db DBTX
}

type PartySummary struct {
    ID          string
    DisplayName string
}

type Detail struct {
    Purchase  Purchase
    Purchaser PartySummary
    Payer     *PartySummary
}

type ListInput struct {
    EventID string
    Status  Status
    Cursor  string
    Limit   int
}

type ListResult struct {
    Purchases  []Detail
    NextCursor string
    Limit      int
}

type purchaseCursor struct {
    Version   int       `json:"v"`
    CreatedAt time.Time `json:"t"`
    ID        string    `json:"i"`
}

type scanner interface {
    Scan(...any) error
}

func NewRepository(db DBTX) *Repository {
    return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, purchase Purchase) (Purchase, error) {
    created, err := scanPurchase(repository.db.QueryRowContext(ctx, `
INSERT INTO purchases (event_id, purchase_ref, channel, purchaser_party_id, payer_party_id, offering_id, participant_count, offering_name_snapshot, offering_kind_snapshot, offering_unit_price_minor, participant_capacity_snapshot, total_amount_minor, currency_code, status, access_token_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING `+purchaseColumns,
        purchase.EventID, purchase.PurchaseRef, purchase.Channel, purchase.PurchaserPartyID, purchase.PayerPartyID,
        purchase.OfferingID, purchase.ParticipantCount, purchase.OfferingNameSnapshot, purchase.OfferingKindSnapshot,
        purchase.OfferingUnitPriceMinor, purchase.ParticipantCapacitySnapshot, purchase.TotalAmountMinor, purchase.CurrencyCode,
        purchase.Status, purchase.accessTokenHash))
    if err != nil {
        return Purchase{}, fmt.Errorf("insert purchase: %w", translateError(err))
    }
    created.Participants = make([]Participant, 0, len(purchase.Participants))
    for _, participant := range purchase.Participants {
        inserted, insertErr := scanParticipant(repository.db.QueryRowContext(ctx, `
INSERT INTO purchase_participants (event_id, purchase_id, party_id, sequence_no, display_name_snapshot)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, party_id, sequence_no, display_name_snapshot`, created.EventID, created.ID, participant.PartyID, participant.SequenceNo, participant.DisplayName))
        if insertErr != nil {
            return Purchase{}, fmt.Errorf("insert purchase participant: %w", translateError(insertErr))
        }
        created.Participants = append(created.Participants, inserted)
    }
    if _, err := repository.db.ExecContext(ctx, `INSERT INTO purchase_status_history (purchase_id, from_status, to_status) VALUES ($1, $2, $3)`, created.ID, StatusDraft, StatusPendingPayment); err != nil {
        return Purchase{}, fmt.Errorf("insert purchase status history: %w", translateError(err))
    }
    return created, nil
}

func (repository *Repository) Get(ctx context.Context, id string) (Detail, error) {
    detail, err := scanDetail(repository.db.QueryRowContext(ctx, `SELECT `+purchaseDetailColumns+`
FROM purchases
JOIN parties AS purchaser ON purchaser.id = purchases.purchaser_party_id
LEFT JOIN parties AS payer ON payer.id = purchases.payer_party_id
WHERE purchases.id = $1`, id))
    if errors.Is(err, sql.ErrNoRows) {
        return Detail{}, ErrNotFound
    }
    if err != nil {
        return Detail{}, fmt.Errorf("get purchase: %w", translateError(err))
    }
    participants, err := repository.participants(ctx, id)
    if err != nil {
        return Detail{}, err
    }
    detail.Purchase.Participants = participants
    return detail, nil
}

func (repository *Repository) List(ctx context.Context, input ListInput) (ListResult, error) {
    limit, cursor, err := normalizeListInput(input)
    if err != nil {
        return ListResult{}, err
    }
    query, arguments := listQuery(input, cursor, limit+1)
    rows, err := repository.db.QueryContext(ctx, query, arguments...)
    if err != nil {
        if cursor != nil && isInvalidCursorQuery(err) {
            return ListResult{}, ErrInvalidCursor
        }
        return ListResult{}, fmt.Errorf("list purchases: %w", translateError(err))
    }
    defer rows.Close()

    result := ListResult{Purchases: make([]Detail, 0, limit), Limit: limit}
    for rows.Next() {
        detail, scanErr := scanDetail(rows)
        if scanErr != nil {
            return ListResult{}, fmt.Errorf("scan purchase list row: %w", scanErr)
        }
        result.Purchases = append(result.Purchases, detail)
    }
    if err := rows.Err(); err != nil {
        return ListResult{}, fmt.Errorf("iterate purchase list rows: %w", err)
    }
    if len(result.Purchases) > limit {
        last := result.Purchases[limit-1].Purchase
        next, encodeErr := encodeCursor(purchaseCursor{Version: 1, CreatedAt: last.CreatedAt, ID: last.ID})
        if encodeErr != nil {
            return ListResult{}, encodeErr
        }
        result.Purchases = result.Purchases[:limit]
        result.NextCursor = next
    }
    return result, nil
}

func (repository *Repository) participants(ctx context.Context, purchaseID string) ([]Participant, error) {
    rows, err := repository.db.QueryContext(ctx, `SELECT id, party_id, sequence_no, display_name_snapshot FROM purchase_participants WHERE purchase_id = $1 ORDER BY sequence_no ASC`, purchaseID)
    if err != nil {
        return nil, fmt.Errorf("list purchase participants: %w", translateError(err))
    }
    defer rows.Close()
    participants := make([]Participant, 0)
    for rows.Next() {
        participant, scanErr := scanParticipant(rows)
        if scanErr != nil {
            return nil, fmt.Errorf("scan purchase participant: %w", scanErr)
        }
        participants = append(participants, participant)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate purchase participants: %w", err)
    }
    return participants, nil
}

func scanPurchase(value scanner) (Purchase, error) {
    var purchase Purchase
    var channel, status string
    var payerID sql.NullString
    if err := value.Scan(
        &purchase.ID, &purchase.EventID, &purchase.PurchaseRef, &channel, &purchase.PurchaserPartyID, &payerID,
        &purchase.OfferingID, &purchase.OfferingNameSnapshot, &purchase.OfferingKindSnapshot,
        &purchase.OfferingUnitPriceMinor, &purchase.ParticipantCapacitySnapshot, &purchase.ParticipantCount,
        &purchase.TotalAmountMinor, &purchase.CurrencyCode, &status, &purchase.Version, &purchase.CreatedAt, &purchase.UpdatedAt,
    ); err != nil {
        return Purchase{}, err
    }
    purchase.Channel = Channel(channel)
    purchase.Status = Status(status)
    if payerID.Valid {
        value := payerID.String
        purchase.PayerPartyID = &value
    }
    purchase.CreatedAt = purchase.CreatedAt.UTC()
    purchase.UpdatedAt = purchase.UpdatedAt.UTC()
    return purchase, nil
}

func scanDetail(value scanner) (Detail, error) {
    var detail Detail
    var channel, status string
    var purchasePayerID, payerID, payerDisplayName sql.NullString
    if err := value.Scan(
        &detail.Purchase.ID, &detail.Purchase.EventID, &detail.Purchase.PurchaseRef, &channel,
        &detail.Purchase.PurchaserPartyID, &purchasePayerID, &detail.Purchase.OfferingID,
        &detail.Purchase.OfferingNameSnapshot, &detail.Purchase.OfferingKindSnapshot,
        &detail.Purchase.OfferingUnitPriceMinor, &detail.Purchase.ParticipantCapacitySnapshot,
        &detail.Purchase.ParticipantCount, &detail.Purchase.TotalAmountMinor, &detail.Purchase.CurrencyCode,
        &status, &detail.Purchase.Version, &detail.Purchase.CreatedAt, &detail.Purchase.UpdatedAt,
        &detail.Purchaser.ID, &detail.Purchaser.DisplayName, &payerID, &payerDisplayName,
    ); err != nil {
        return Detail{}, err
    }
    detail.Purchase.Channel = Channel(channel)
    detail.Purchase.Status = Status(status)
    if purchasePayerID.Valid {
        payerIDValue := purchasePayerID.String
        detail.Purchase.PayerPartyID = &payerIDValue
    }
    if payerID.Valid {
        payerIDValue := payerID.String
        detail.Payer = &PartySummary{ID: payerIDValue, DisplayName: payerDisplayName.String}
    }
    detail.Purchase.CreatedAt = detail.Purchase.CreatedAt.UTC()
    detail.Purchase.UpdatedAt = detail.Purchase.UpdatedAt.UTC()
    return detail, nil
}

func scanParticipant(value scanner) (Participant, error) {
    var participant Participant
    var partyID sql.NullString
    if err := value.Scan(&participant.ID, &partyID, &participant.SequenceNo, &participant.DisplayName); err != nil {
        return Participant{}, err
    }
    if partyID.Valid {
        partyIDValue := partyID.String
        participant.PartyID = &partyIDValue
    }
    return participant, nil
}

func normalizeListInput(input ListInput) (int, *purchaseCursor, error) {
    limit := input.Limit
    if limit == 0 {
        limit = DefaultListLimit
    }
    if limit < 1 || limit > MaxListLimit {
        return 0, nil, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidInput, MaxListLimit)
    }
    if input.Status != "" && !validListStatus(input.Status) {
        return 0, nil, fmt.Errorf("%w: unsupported purchase status filter", ErrInvalidInput)
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

func validListStatus(status Status) bool {
    switch status {
    case StatusPendingPayment, StatusPaid, StatusEligible, StatusCancelled:
        return true
    default:
        return false
    }
}

func listQuery(input ListInput, cursor *purchaseCursor, limit int) (string, []any) {
    conditions := make([]string, 0, 3)
    arguments := make([]any, 0, 4)
    if input.EventID != "" {
        arguments = append(arguments, input.EventID)
        conditions = append(conditions, fmt.Sprintf("purchases.event_id = $%d", len(arguments)))
    }
    if input.Status != "" {
        arguments = append(arguments, input.Status)
        conditions = append(conditions, fmt.Sprintf("purchases.status = $%d", len(arguments)))
    }
    if cursor != nil {
        arguments = append(arguments, cursor.CreatedAt, cursor.ID)
        conditions = append(conditions, fmt.Sprintf("(purchases.created_at < $%d OR (purchases.created_at = $%d AND purchases.id > $%d))", len(arguments)-1, len(arguments)-1, len(arguments)))
    }
    query := `SELECT ` + purchaseDetailColumns + `
FROM purchases
JOIN parties AS purchaser ON purchaser.id = purchases.purchaser_party_id
LEFT JOIN parties AS payer ON payer.id = purchases.payer_party_id`
    if len(conditions) > 0 {
        query += "\nWHERE " + strings.Join(conditions, " AND ")
    }
    arguments = append(arguments, limit)
    query += fmt.Sprintf("\nORDER BY purchases.created_at DESC, purchases.id ASC LIMIT $%d", len(arguments))
    return query, arguments
}

func encodeCursor(cursor purchaseCursor) (string, error) {
    payload, err := json.Marshal(cursor)
    if err != nil {
        return "", fmt.Errorf("encode purchase cursor: %w", err)
    }
    return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeCursor(value string) (purchaseCursor, error) {
    if len(value) == 0 || len(value) > maxCursorLength {
        return purchaseCursor{}, ErrInvalidCursor
    }
    payload, err := base64.RawURLEncoding.DecodeString(value)
    if err != nil || len(payload) == 0 || len(payload) > maxCursorPayload {
        return purchaseCursor{}, ErrInvalidCursor
    }
    decoder := json.NewDecoder(bytes.NewReader(payload))
    decoder.DisallowUnknownFields()
    var cursor purchaseCursor
    if err := decoder.Decode(&cursor); err != nil {
        return purchaseCursor{}, ErrInvalidCursor
    }
    if err := decoder.Decode(&struct{}{}); err != io.EOF {
        return purchaseCursor{}, ErrInvalidCursor
    }
    if cursor.Version != 1 || cursor.CreatedAt.IsZero() || len(cursor.ID) != 36 {
        return purchaseCursor{}, ErrInvalidCursor
    }
    cursor.CreatedAt = cursor.CreatedAt.UTC()
    return cursor, nil
}

func translateError(err error) error {
    var postgresError *pgconn.PgError
    if !errors.As(err, &postgresError) {
        return err
    }
    switch postgresError.Code {
    case "23505":
        if postgresError.ConstraintName == "purchases_purchase_ref_key" {
            return ErrDuplicateReference
        }
        return ErrConflict
    case "23502", "23503", "23514", "22P02":
        return ErrInvalidInput
    }
    return err
}

func isInvalidCursorQuery(err error) bool {
    var postgresError *pgconn.PgError
    return errors.As(err, &postgresError) && postgresError.Code == "22P02"
}
