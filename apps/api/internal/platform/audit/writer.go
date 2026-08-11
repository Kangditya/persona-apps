package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Entry struct {
	ActorOperatorID string
	ActorReference  string
	Action          string
	TargetType      string
	TargetID        string
	RequestID       string
	Reason          string
	Source          string
	Permission      string
	Before          any
	After           any
}

func Write(ctx context.Context, tx *sql.Tx, entry Entry) error {
	before, err := json.Marshal(entry.Before)
	if err != nil {
		return fmt.Errorf("marshal audit before data: %w", err)
	}
	after, err := json.Marshal(entry.After)
	if err != nil {
		return fmt.Errorf("marshal audit after data: %w", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO audit_log (actor_operator_id, actor_reference, action, target_type, target_id, request_id, reason, before_data, after_data, source_application, source, permission) VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, ''), NULLIF($7, ''), $8::jsonb, $9::jsonb, $10, $10, NULLIF($11, ''))", entry.ActorOperatorID, entry.ActorReference, entry.Action, entry.TargetType, entry.TargetID, entry.RequestID, entry.Reason, string(before), string(after), entry.Source, entry.Permission)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
