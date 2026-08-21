package identity

import (
    "context"
    "database/sql"
    "errors"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryCreateAndGet(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })

    email := "purchaser@example.test"
    phone := "+628100000000"
    createdAt := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
    updatedAt := createdAt.Add(time.Minute)
    row := sqlmock.NewRows([]string{"id", "party_type", "display_name", "email", "phone", "created_at", "updated_at"}).
        AddRow("11111111-1111-1111-1111-111111111111", "PERSON", "Purchaser", email, phone, createdAt, updatedAt)
    mock.ExpectQuery("INSERT INTO parties").WithArgs(PartyTypePerson, "Purchaser", email, phone).WillReturnRows(row)

    repository := NewRepository(database)
    created, err := repository.Create(context.Background(), Party{Type: PartyTypePerson, DisplayName: "Purchaser", Email: &email, Phone: &phone})
    if err != nil {
        t.Fatal(err)
    }
    if created.ID == "" || created.Email == nil || *created.Email != email || created.Phone == nil || *created.Phone != phone || !created.CreatedAt.Equal(createdAt) || !created.UpdatedAt.Equal(updatedAt) {
        t.Fatalf("created Party = %#v", created)
    }

    mock.ExpectQuery("SELECT id, party_type").WithArgs(created.ID).WillReturnRows(
        sqlmock.NewRows([]string{"id", "party_type", "display_name", "email", "phone", "created_at", "updated_at"}).
            AddRow(created.ID, "PERSON", "Purchaser", nil, nil, createdAt, updatedAt),
    )
    found, err := repository.Get(context.Background(), created.ID)
    if err != nil {
        t.Fatal(err)
    }
    if found.ID != created.ID || found.Email != nil || found.Phone != nil || !found.CreatedAt.Equal(createdAt) || !found.UpdatedAt.Equal(updatedAt) {
        t.Fatalf("found Party = %#v", found)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestRepositoryGetReturnsNotFound(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    mock.ExpectQuery("SELECT id, party_type").WithArgs("missing").WillReturnError(sql.ErrNoRows)

    _, err = NewRepository(database).Get(context.Background(), "missing")
    if !errors.Is(err, ErrNotFound) {
        t.Fatalf("Get() error = %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}
