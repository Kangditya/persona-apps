package application

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/base32"
    "encoding/base64"
    "fmt"
    "io"
    "time"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
    purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
)

type Repository interface {
    Create(context.Context, purchasingdomain.Purchase) (purchasingdomain.Purchase, error)
    Get(context.Context, string) (purchasingdomain.Detail, error)
    List(context.Context, purchasingdomain.ListInput) (purchasingdomain.ListResult, error)
}

type RepositoryFactory func(platformdb.DBTX) Repository

type PartyCreator interface {
    CreateInTransaction(context.Context, *sql.Tx, identitydomain.Party) (identitydomain.Party, error)
}

type CheckoutLocker func(context.Context, *sql.Tx, string, time.Time) (purchasingdomain.CheckoutSource, error)
type ReservationCreator func(context.Context, *sql.Tx, purchasingdomain.CheckoutSource, purchasingdomain.Purchase) (purchasingdomain.Reservation, error)

type Service struct {
    db            *sql.DB
    repositories  RepositoryFactory
    parties       PartyCreator
    lockCheckout  CheckoutLocker
    createReserve ReservationCreator
    now           func() time.Time
    random        io.Reader
}

func NewService(db *sql.DB, repositories RepositoryFactory, parties PartyCreator, lockCheckout CheckoutLocker, createReserve ReservationCreator) *Service {
    return &Service{
        db: db, repositories: repositories, parties: parties,
        lockCheckout: lockCheckout, createReserve: createReserve, now: time.Now, random: rand.Reader,
    }
}

func (service *Service) SetRandomSource(random io.Reader) {
    service.random = random
}

func (service *Service) List(ctx context.Context, input purchasingdomain.ListInput) (purchasingdomain.ListResult, error) {
    return service.repositories(service.db).List(ctx, input)
}

func (service *Service) Get(ctx context.Context, id string) (purchasingdomain.Detail, error) {
    return service.repositories(service.db).Get(ctx, id)
}

func (service *Service) CreateCommonPurchaseInTransaction(ctx context.Context, tx *sql.Tx, offeringID string, relationships purchasingdomain.Relationships) (purchasingdomain.Purchase, purchasingdomain.Reservation, string, error) {
    source, err := service.lockCheckout(ctx, tx, offeringID, service.clock())
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    purchaserID, payerID, participants, err := service.persistRelationships(ctx, tx, relationships)
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    reference, err := NewPurchaseReference(source.Event.EventYear, service.randomSource())
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    accessToken, accessTokenHash, err := NewPurchaseAccessToken(service.randomSource())
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    purchase, err := purchasingdomain.NewSnapshotPurchase(purchasingdomain.SnapshotPurchaseInput{
        Source: source, PurchaseRef: reference, PurchaserPartyID: purchaserID, PayerPartyID: payerID,
        Participants: participants, AccessTokenHash: accessTokenHash,
    })
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    created, err := service.repositories(tx).Create(ctx, purchase)
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    reservation, err := service.createReserve(ctx, tx, source, created)
    if err != nil {
        return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
    }
    for _, eventType := range []string{"PurchaseReserved", "PurchasePendingPayment"} {
        if err := outbox.Write(ctx, tx, outbox.Event{
            AggregateType: "Purchase", AggregateID: created.ID, EventType: eventType,
            Payload: map[string]string{"purchase_id": created.ID}, OccurredAt: source.OccurredAt,
        }); err != nil {
            return purchasingdomain.Purchase{}, purchasingdomain.Reservation{}, "", err
        }
    }
    return created, reservation, accessToken, nil
}

func (service *Service) persistRelationships(ctx context.Context, tx *sql.Tx, relationships purchasingdomain.Relationships) (string, string, []purchasingdomain.PurchaseParticipantInput, error) {
    partyIDs := make(map[string]string, len(relationships.DeclaredParties))
    for _, declared := range relationships.DeclaredParties {
        created, err := service.parties.CreateInTransaction(ctx, tx, declared.Party)
        if err != nil {
            return "", "", nil, err
        }
        partyIDs[declared.Ref] = created.ID
    }
    purchaserID, purchaserFound := partyIDs[relationships.PurchaserRef]
    payerID, payerFound := partyIDs[relationships.PayerRef]
    if !purchaserFound || !payerFound {
        return "", "", nil, purchasingdomain.ErrInvalidRelationships
    }
    participants := make([]purchasingdomain.PurchaseParticipantInput, 0, len(relationships.Participants))
    for _, participant := range relationships.Participants {
        input := purchasingdomain.PurchaseParticipantInput{DisplayName: participant.DisplayName}
        if participant.PartyRef != "" {
            partyID, found := partyIDs[participant.PartyRef]
            if !found {
                return "", "", nil, purchasingdomain.ErrInvalidRelationships
            }
            input.PartyID = &partyID
        }
        participants = append(participants, input)
    }
    return purchaserID, payerID, participants, nil
}

const (
    purchaseReferenceBytes   = 10
    purchaseAccessTokenBytes = 32
)

func NewPurchaseReference(eventYear int, random io.Reader) (string, error) {
    if eventYear < 1900 || eventYear > 9999 || random == nil {
        return "", purchasingdomain.ErrInvalidInput
    }
    var entropy [purchaseReferenceBytes]byte
    if _, err := io.ReadFull(random, entropy[:]); err != nil {
        return "", fmt.Errorf("generate purchase reference: %w", err)
    }
    return "QRB-" + fmt.Sprint(eventYear) + "-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entropy[:]), nil
}

func NewPurchaseAccessToken(random io.Reader) (string, []byte, error) {
    if random == nil {
        return "", nil, purchasingdomain.ErrInvalidInput
    }
    var token [purchaseAccessTokenBytes]byte
    if _, err := io.ReadFull(random, token[:]); err != nil {
        return "", nil, fmt.Errorf("generate purchase access token: %w", err)
    }
    encoded := base64.RawURLEncoding.EncodeToString(token[:])
    hash := sha256.Sum256([]byte(encoded))
    return encoded, hash[:], nil
}

func (service *Service) clock() time.Time {
    if service.now == nil {
        return time.Now().UTC()
    }
    return service.now().UTC()
}

func (service *Service) randomSource() io.Reader {
    if service.random == nil {
        return rand.Reader
    }
    return service.random
}
