package outbox

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "time"
)

type Event struct {
    AggregateType string
    AggregateID   string
    EventType     string
    Payload       any
    OccurredAt    time.Time
}

func Write(ctx context.Context, tx *sql.Tx, event Event) error {
    if event.AggregateType == "" || event.AggregateID == "" || event.EventType == "" || event.OccurredAt.IsZero() {
        return errors.New("outbox event is incomplete")
    }
    payload, err := json.Marshal(event.Payload)
    if err != nil {
        return fmt.Errorf("marshal outbox payload: %w", err)
    }
    if _, err := tx.ExecContext(ctx, "INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, occurred_at) VALUES ($1, $2::uuid, $3, $4::jsonb, $5)", event.AggregateType, event.AggregateID, event.EventType, string(payload), event.OccurredAt.UTC()); err != nil {
        return fmt.Errorf("insert outbox event: %w", err)
    }
    return nil
}
