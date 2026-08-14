package seeder

import "context"

type localOperator struct{}

func (localOperator) Name() string { return "development.local-operator" }
func (localOperator) Group() Group { return Development }
func (localOperator) Order() int   { return 10 }
func (localOperator) Run(ctx context.Context, db DBTX) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO operator_users (external_subject, display_name, status)
VALUES ('local-operator', 'Local Operator', 'ACTIVE')
ON CONFLICT (external_subject) DO UPDATE
SET display_name = EXCLUDED.display_name, status = EXCLUDED.status, updated_at = now()`)
	return err
}
