package seeder

import "context"

type sampleEvent struct{}

const (
	samplePartyID    = "00000000-0000-0000-0000-000000000001"
	sampleEventID    = "00000000-0000-0000-0000-000000000002"
	sampleLocationID = "00000000-0000-0000-0000-000000000003"
	sampleOfferingID = "00000000-0000-0000-0000-000000000004"
)

func (sampleEvent) Name() string { return "development.sample-event" }
func (sampleEvent) Group() Group { return Development }
func (sampleEvent) Order() int   { return 100 }
func (sampleEvent) Run(ctx context.Context, db DBTX) error {
	statements := []struct {
		query string
		args  []any
	}{
		{
			query: `
INSERT INTO parties (id, party_type, display_name, email)
VALUES ($1, 'PERSON', 'Demo Purchaser', 'demo@example.invalid')
ON CONFLICT (id) DO NOTHING`,
			args: []any{samplePartyID},
		},
		{
			query: `
INSERT INTO qurban_events (id, event_year, name, status, participant_quota)
VALUES ($1, 2099, 'Demo Qurban 2099', 'DRAFT', 7)
ON CONFLICT (id) DO NOTHING`,
			args: []any{sampleEventID},
		},
		{
			query: `
INSERT INTO event_locations (id, event_id, code, name, location_type)
VALUES ($1, $2, 'DEMO-PEN', 'Demo Pen', 'PEN')
ON CONFLICT (id) DO NOTHING`,
			args: []any{sampleLocationID, sampleEventID},
		},
		{
			query: `
INSERT INTO offerings (id, event_id, code, name, offering_kind, description, price_minor, currency_code, participant_capacity, status)
VALUES ($1, $2, 'DEMO-COW-7', 'Demo Cow Share', 'COW_SHARE', 'Development-only sample offering', 1000000, 'IDR', 7, 'PUBLISHED')
ON CONFLICT (id) DO NOTHING`,
			args: []any{sampleOfferingID, sampleEventID},
		},
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			return err
		}
	}
	return nil
}
