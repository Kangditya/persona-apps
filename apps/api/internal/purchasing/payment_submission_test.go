package purchasing

import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestAuthorizePurchaseTokenValidatesShapeAndConstantDigest(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer database.Close()
    token := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
    digest := sha256.Sum256([]byte(token))

    mock.ExpectQuery("SELECT access_token_hash FROM purchases").WithArgs("purchase").WillReturnRows(sqlmock.NewRows([]string{"access_token_hash"}).AddRow(digest[:]))
    if err := AuthorizePurchaseToken(context.Background(), database, "purchase", token); err != nil {
        t.Fatal(err)
    }
    if err := AuthorizePurchaseToken(context.Background(), database, "purchase", "bad-token"); !errors.Is(err, ErrUnauthenticated) {
        t.Fatalf("malformed token error = %v", err)
    }
    mock.ExpectQuery("SELECT access_token_hash FROM purchases").WithArgs("purchase").WillReturnRows(sqlmock.NewRows([]string{"access_token_hash"}).AddRow(make([]byte, 32)))
    if err := AuthorizePurchaseToken(context.Background(), database, "purchase", base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))); !errors.Is(err, ErrForbidden) {
        t.Fatalf("wrong token error = %v", err)
    }
    mock.ExpectQuery("SELECT access_token_hash FROM purchases").WithArgs("missing").WillReturnRows(sqlmock.NewRows([]string{"access_token_hash"}))
    if err := AuthorizePurchaseToken(context.Background(), database, "missing", token); !errors.Is(err, ErrNotFound) {
        t.Fatalf("missing purchase error = %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}
