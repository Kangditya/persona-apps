package persistence

import (
    "context"
    "database/sql"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryCreateWritesPurchaseParticipantsAndInitialHistory(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    purchase := testPurchase(t)
    now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
    mock.ExpectQuery("INSERT INTO purchases").
        WithArgs("event-id", "purchase-ref", ChannelCommon, "purchaser-id", "payer-id", "offering-id", 2, "Share", "SHARE", int64(100), int32(2), int64(200), "IDR", StatusPendingPayment, make([]byte, 32)).
        WillReturnRows(purchaseRows(now, "purchase-id"))
    mock.ExpectQuery("INSERT INTO purchase_participants").
        WithArgs("event-id", "purchase-id", "purchaser-id", 1, "Siti").
        WillReturnRows(sqlmock.NewRows([]string{"id", "party_id", "sequence_no", "display_name_snapshot"}).AddRow("participant-1", "purchaser-id", 1, "Siti"))
    mock.ExpectQuery("INSERT INTO purchase_participants").
        WithArgs("event-id", "purchase-id", nil, 2, "Ahmad").
        WillReturnRows(sqlmock.NewRows([]string{"id", "party_id", "sequence_no", "display_name_snapshot"}).AddRow("participant-2", nil, 2, "Ahmad"))
    mock.ExpectExec("INSERT INTO purchase_status_history").
        WithArgs("purchase-id", StatusDraft, StatusPendingPayment).
        WillReturnResult(sqlmock.NewResult(1, 1))

    created, err := NewRepository(database).Create(context.Background(), purchase)
    if err != nil {
        t.Fatal(err)
    }
    if created.ID != "purchase-id" || len(created.Participants) != 2 || created.Participants[0].ID != "participant-1" || created.Participants[1].PartyID != nil {
        t.Fatalf("created purchase = %#v", created)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestRepositoryGetAndListReturnSnapshotsAndPartySummaries(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
    purchaseID := "11111111-1111-1111-1111-111111111111"
    detailColumns := []string{"id", "event_id", "purchase_ref", "channel", "purchaser_party_id", "payer_party_id", "offering_id", "offering_name_snapshot", "offering_kind_snapshot", "offering_unit_price_minor", "participant_capacity_snapshot", "participant_count", "total_amount_minor", "currency_code", "status", "version", "created_at", "updated_at", "purchaser_id", "purchaser_display_name", "payer_id", "payer_display_name"}
    mock.ExpectQuery(regexp.QuoteMeta("SELECT " + purchaseDetailColumns)).WithArgs(purchaseID).WillReturnRows(
        sqlmock.NewRows(detailColumns).AddRow(purchaseID, "event-id", "purchase-ref", "COMMON", "purchaser-id", "payer-id", "offering-id", "Share", "SHARE", 100, 2, 2, 200, "IDR", "PENDING_PAYMENT", 1, now, now, "purchaser-id", "Siti", "payer-id", "Budi"),
    )
    mock.ExpectQuery("SELECT id, party_id, sequence_no, display_name_snapshot FROM purchase_participants").WithArgs(purchaseID).WillReturnRows(
        sqlmock.NewRows([]string{"id", "party_id", "sequence_no", "display_name_snapshot"}).
            AddRow("participant-1", "purchaser-id", 1, "Siti").
            AddRow("participant-2", nil, 2, "Ahmad"),
    )

    detail, err := NewRepository(database).Get(context.Background(), purchaseID)
    if err != nil {
        t.Fatal(err)
    }
    if detail.Purchaser.DisplayName != "Siti" || detail.Payer == nil || detail.Payer.DisplayName != "Budi" || detail.Purchase.OfferingNameSnapshot != "Share" || len(detail.Purchase.Participants) != 2 {
        t.Fatalf("detail = %#v", detail)
    }

    mock.ExpectQuery("SELECT purchases.id").WithArgs(2).WillReturnRows(
        sqlmock.NewRows(detailColumns).
            AddRow(purchaseID, "event-id", "purchase-ref", "COMMON", "purchaser-id", "payer-id", "offering-id", "Share", "SHARE", 100, 2, 2, 200, "IDR", "PENDING_PAYMENT", 1, now, now, "purchaser-id", "Siti", "payer-id", "Budi").
            AddRow("22222222-2222-2222-2222-222222222222", "event-id", "purchase-next", "COMMON", "purchaser-id", "payer-id", "offering-id", "Share", "SHARE", 100, 2, 2, 200, "IDR", "PENDING_PAYMENT", 1, now.Add(-time.Second), now.Add(-time.Second), "purchaser-id", "Siti", "payer-id", "Budi"),
    )
    listed, err := NewRepository(database).List(context.Background(), ListInput{Limit: 1})
    if err != nil {
        t.Fatal(err)
    }
    if len(listed.Purchases) != 1 || listed.NextCursor == "" {
        t.Fatalf("list result = %#v", listed)
    }
    if _, err := decodeCursor(listed.NextCursor); err != nil {
        t.Fatal(err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestNormalizeListInputRejectsUnsupportedStatusAndMalformedCursor(t *testing.T) {
    if _, _, err := normalizeListInput(ListInput{Status: StatusDraft}); err == nil {
        t.Fatal("draft status unexpectedly accepted")
    }
    if _, _, err := normalizeListInput(ListInput{Cursor: "not-a-cursor"}); err == nil {
        t.Fatal("malformed cursor unexpectedly accepted")
    }
}

func TestPurchaseAndReservationRowsMapNullableValuesAndUTC(t *testing.T) {
    location := time.FixedZone("test", 7*60*60)
    created := time.Date(2026, time.August, 21, 10, 0, 0, 0, location)
    purchase := (purchaseRow{
        ID:                          "purchase-id",
        EventID:                     "event-id",
        PurchaseRef:                 "purchase-ref",
        Channel:                     "COMMON",
        PurchaserPartyID:            "purchaser-id",
        PayerPartyID:                sql.NullString{String: "payer-id", Valid: true},
        OfferingID:                  "offering-id",
        OfferingNameSnapshot:        "Share",
        OfferingKindSnapshot:        "SHARE",
        OfferingUnitPriceMinor:      100,
        ParticipantCapacitySnapshot: 2,
        ParticipantCount:            2,
        TotalAmountMinor:            200,
        CurrencyCode:                "IDR",
        Status:                      "PENDING_PAYMENT",
        Version:                     1,
        CreatedAt:                   created,
        UpdatedAt:                   created.Add(time.Minute),
    }).purchase()
    if purchase.ID != "purchase-id" || purchase.PayerPartyID == nil || *purchase.PayerPartyID != "payer-id" || purchase.Channel != ChannelCommon || purchase.Status != StatusPendingPayment {
        t.Fatalf("purchase = %#v", purchase)
    }
    if purchase.CreatedAt.Location() != time.UTC || purchase.UpdatedAt.Location() != time.UTC {
        t.Fatalf("purchase timestamps = %v, %v", purchase.CreatedAt, purchase.UpdatedAt)
    }

    reservationRow := reservationRow{
        ID:               "reservation-id",
        EventID:          "event-id",
        PurchaseID:       "purchase-id",
        OfferingID:       "offering-id",
        AttemptNo:        1,
        ParticipantUnits: 2,
        Status:           "RESERVED",
        ExpiresAt:        sql.NullTime{Time: created.Add(24 * time.Hour), Valid: true},
        Version:          1,
        CreatedAt:        created,
        UpdatedAt:        created.Add(time.Minute),
    }
    reservation, err := reservationRow.reservation()
    if err != nil {
        t.Fatal(err)
    }
    if reservation.Status != ReservationStatusReserved || reservation.ExpiresAt.Location() != time.UTC || reservation.CreatedAt.Location() != time.UTC || reservation.UpdatedAt.Location() != time.UTC {
        t.Fatalf("reservation = %#v", reservation)
    }
    reservationRow.ExpiresAt = sql.NullTime{}
    if _, err := reservationRow.reservation(); err == nil {
        t.Fatal("reservation without expiry unexpectedly accepted")
    }
}

func testPurchase(t *testing.T) Purchase {
    t.Helper()
    purchaser := "purchaser-id"
    purchase, err := NewPurchase(CreateInput{
        EventID: "event-id", PurchaseRef: "purchase-ref", PurchaserPartyID: purchaser, PayerPartyID: "payer-id",
        OfferingID: "offering-id", OfferingNameSnapshot: "Share", OfferingKindSnapshot: "SHARE",
        OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: 2, TotalAmountMinor: 200, CurrencyCode: "IDR",
        Participants:    []PurchaseParticipantInput{{PartyID: &purchaser, DisplayName: "Siti"}, {DisplayName: "Ahmad"}},
        AccessTokenHash: make([]byte, 32),
    })
    if err != nil {
        t.Fatal(err)
    }
    return purchase
}

func purchaseRows(now time.Time, id string) *sqlmock.Rows {
    return sqlmock.NewRows([]string{"id", "event_id", "purchase_ref", "channel", "purchaser_party_id", "payer_party_id", "offering_id", "offering_name_snapshot", "offering_kind_snapshot", "offering_unit_price_minor", "participant_capacity_snapshot", "participant_count", "total_amount_minor", "currency_code", "status", "version", "created_at", "updated_at"}).
        AddRow(id, "event-id", "purchase-ref", "COMMON", "purchaser-id", "payer-id", "offering-id", "Share", "SHARE", 100, 2, 2, 200, "IDR", "PENDING_PAYMENT", 1, now, now)
}
