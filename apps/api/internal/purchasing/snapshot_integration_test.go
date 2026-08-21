package purchasing

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
	"time"

	platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestSnapshotPurchasePostgreSQLStaysHistoricalAfterSourceChanges(t *testing.T) {
	database := integrationDatabase(t)
	ctx := context.Background()
	acquireActiveEventIntegrationLock(t, ctx, database)
	prefix := fmt.Sprintf("snapshot-%d", time.Now().UnixNano())
	eventID, offeringID, partyID := createReservationFixture(t, ctx, database, prefix, nil, nil)
	t.Cleanup(func() { cleanupReservationFixture(t, ctx, database, eventID, partyID) })
	now := time.Now().UTC().Truncate(time.Microsecond)
	hash := sha256.Sum256([]byte(prefix))

	var created Purchase
	if err := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
		source, err := LockCheckout(ctx, transaction, eventID, offeringID, now)
		if err != nil {
			return err
		}
		purchase, err := NewSnapshotPurchase(SnapshotPurchaseInput{
			Source: source, PurchaseRef: prefix, PurchaserPartyID: partyID, PayerPartyID: partyID,
			Participants: []PurchaseParticipantInput{{PartyID: &partyID, DisplayName: prefix + " purchaser"}}, AccessTokenHash: hash[:],
		})
		if err != nil {
			return err
		}
		created, err = NewRepository(transaction).Create(ctx, purchase)
		return err
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := database.ExecContext(ctx, `UPDATE qurban_events SET name = $1 WHERE id = $2`, prefix+" changed event", eventID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE offerings SET name = $1, offering_kind = 'CHANGED', price_minor = 999, participant_capacity = 2, currency_code = 'USD' WHERE id = $2`, prefix+" changed offering", offeringID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE parties SET display_name = $1 WHERE id = $2`, prefix+" changed party", partyID); err != nil {
		t.Fatal(err)
	}

	found, err := NewRepository(database).Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Purchase.EventID != eventID || found.Purchase.OfferingID != offeringID || found.Purchase.OfferingNameSnapshot != "One share" || found.Purchase.OfferingKindSnapshot != "SHARE" || found.Purchase.OfferingUnitPriceMinor != 100 || found.Purchase.ParticipantCapacitySnapshot != 1 || found.Purchase.TotalAmountMinor != 100 || found.Purchase.CurrencyCode != "IDR" {
		t.Fatalf("stored commercial snapshot = %#v", found.Purchase)
	}
	if len(found.Purchase.Participants) != 1 || found.Purchase.Participants[0].DisplayName != prefix+" purchaser" || found.Purchaser.DisplayName != prefix+" changed party" {
		t.Fatalf("stored participant/current purchaser = %#v / %#v", found.Purchase.Participants, found.Purchaser)
	}
}
