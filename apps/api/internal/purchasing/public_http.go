package purchasing

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Kangditya/persona-apps/apps/api/internal/identity"
	"github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
	"github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
	"github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
	"github.com/gin-gonic/gin"
)

const (
	publicPurchaseBodyLimit        = 64 << 10
	publicPurchaseReferenceTries   = 5
	publicPurchaseReferenceBytes   = 10
	publicPurchaseAccessTokenBytes = 32
)

type PublicHandler struct {
	db     *sql.DB
	cipher idempotency.Cipher
	logger *slog.Logger
	now    func() time.Time
	random io.Reader
}

type publicStringField struct {
	Value string
	Set   bool
}

type publicPartyDeclarationRequest struct {
	PartyRef    publicStringField `json:"party_ref"`
	DisplayName publicStringField `json:"display_name"`
	Email       publicStringField `json:"email"`
	Phone       publicStringField `json:"phone"`
}

type publicParticipantRequest struct {
	PartyRef    publicStringField `json:"party_ref"`
	DisplayName publicStringField `json:"display_name"`
}

type publicCreatePurchaseRequest struct {
	OfferingID   string                        `json:"offering_id"`
	Purchaser    publicPartyDeclarationRequest `json:"purchaser"`
	Payer        publicPartyDeclarationRequest `json:"payer"`
	Participants []publicParticipantRequest    `json:"participants"`
}

type publicPurchaseOffering struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Kind                string `json:"kind"`
	UnitPriceMinor      int64  `json:"unit_price_minor"`
	ParticipantCapacity int32  `json:"participant_capacity"`
}

type publicPurchaseParticipant struct {
	SequenceNo  int    `json:"sequence_no"`
	DisplayName string `json:"display_name"`
}

type publicPurchase struct {
	ID                   string                      `json:"id"`
	PurchaseRef          string                      `json:"purchase_ref"`
	Channel              Channel                     `json:"channel"`
	Offering             publicPurchaseOffering      `json:"offering"`
	ParticipantCount     int                         `json:"participant_count"`
	Participants         []publicPurchaseParticipant `json:"participants"`
	TotalAmountMinor     int64                       `json:"total_amount_minor"`
	CurrencyCode         string                      `json:"currency_code"`
	Status               Status                      `json:"status"`
	ReservationExpiresAt time.Time                   `json:"reservation_expires_at"`
	CreatedAt            time.Time                   `json:"created_at"`
}

type publicCreatePurchaseResponse struct {
	Data struct {
		Purchase    publicPurchase `json:"purchase"`
		AccessToken string         `json:"access_token"`
	} `json:"data"`
}

func NewPublicHandler(db *sql.DB, cipher idempotency.Cipher, logger *slog.Logger) *PublicHandler {
	return &PublicHandler{db: db, cipher: cipher, logger: logger, now: time.Now, random: rand.Reader}
}

func (handler *PublicHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/purchases", handler.create)
}

func (field *publicStringField) UnmarshalJSON(data []byte) error {
	field.Set = true
	if string(data) == "null" {
		return errors.New("string fields cannot be null")
	}
	return json.Unmarshal(data, &field.Value)
}

func (handler *PublicHandler) create(c *gin.Context) {
	var request publicCreatePurchaseRequest
	if !decodePublicPurchase(c, &request) {
		return
	}
	key, err := httpx.IdempotencyKey(c.GetHeader("Idempotency-Key"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "Idempotency-Key is required and invalid", httpx.RequestID(c.Request.Context()), nil)
		return
	}
	offeringID, err := httpx.CanonicalUUID(request.OfferingID)
	if err != nil {
		writePublicPurchaseError(c, handler.logger, ErrInvalidInput)
		return
	}
	relationships, err := request.relationships()
	if err != nil {
		writePublicPurchaseError(c, handler.logger, err)
		return
	}
	hash, err := idempotency.RequestHash(struct {
		OfferingID    string        `json:"offering_id"`
		Relationships Relationships `json:"relationships"`
	}{OfferingID: offeringID, Relationships: relationships})
	if err != nil {
		writePublicPurchaseError(c, handler.logger, err)
		return
	}
	if handler.db == nil {
		writePublicPurchaseError(c, handler.logger, httpx.ErrDependencyUnavailable)
		return
	}

	command := idempotency.Command{
		Namespace:   "storefront.purchase.create.guest",
		Key:         key,
		RequestHash: hash,
		Retention:   0,
	}
	var response idempotency.Response
	for attempt := 0; attempt < publicPurchaseReferenceTries; attempt++ {
		response, _, err = idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, command, func(transaction *sql.Tx) (idempotency.Response, error) {
			return handler.createCheckout(c.Request.Context(), transaction, offeringID, relationships)
		})
		if !errors.Is(err, ErrDuplicateReference) {
			break
		}
	}
	if errors.Is(err, ErrDuplicateReference) {
		err = errors.New("purchase reference generation exhausted")
	}
	if err != nil {
		writePublicPurchaseError(c, handler.logger, err)
		return
	}
	c.Data(response.Status, "application/json", response.Body)
}

func (request publicCreatePurchaseRequest) relationships() (Relationships, error) {
	participants := make([]ParticipantInput, 0, len(request.Participants))
	for _, participant := range request.Participants {
		if participant.PartyRef.Set == participant.DisplayName.Set {
			return Relationships{}, ErrInvalidRelationships
		}
		participants = append(participants, ParticipantInput{PartyRef: participant.PartyRef.Value, DisplayName: participant.DisplayName.Value})
	}
	payer := PayerInput{Ref: request.Payer.PartyRef.Value}
	if request.Payer.DisplayName.Set || request.Payer.Email.Set || request.Payer.Phone.Set {
		declaration := request.Payer.declaration()
		payer = PayerInput{Declaration: &declaration}
	}
	return NewRelationships(RelationshipsInput{
		Purchaser:    request.Purchaser.declaration(),
		Payer:        payer,
		Participants: participants,
	})
}

func (request publicPartyDeclarationRequest) declaration() PartyDeclarationInput {
	return PartyDeclarationInput{
		Ref: request.PartyRef.Value,
		Party: identity.CreateInput{
			Type:        identity.PartyTypePerson,
			DisplayName: request.DisplayName.Value,
			Email:       publicOptionalString(request.Email),
			Phone:       publicOptionalString(request.Phone),
		},
	}
}

func publicOptionalString(field publicStringField) *string {
	if !field.Set {
		return nil
	}
	value := field.Value
	return &value
}

func (handler *PublicHandler) createCheckout(ctx context.Context, transaction *sql.Tx, offeringID string, relationships Relationships) (idempotency.Response, error) {
	source, err := LockCheckoutForOffering(ctx, transaction, offeringID, handler.clock())
	if err != nil {
		return idempotency.Response{}, err
	}
	purchaserID, payerID, participants, err := persistRelationships(ctx, transaction, relationships)
	if err != nil {
		return idempotency.Response{}, err
	}
	reference, err := newPublicPurchaseReference(source.Event.EventYear, handler.randomSource())
	if err != nil {
		return idempotency.Response{}, err
	}
	accessToken, accessTokenHash, err := newPublicPurchaseAccessToken(handler.randomSource())
	if err != nil {
		return idempotency.Response{}, err
	}
	purchase, err := NewSnapshotPurchase(SnapshotPurchaseInput{
		Source:           source,
		PurchaseRef:      reference,
		PurchaserPartyID: purchaserID,
		PayerPartyID:     payerID,
		Participants:     participants,
		AccessTokenHash:  accessTokenHash,
	})
	if err != nil {
		return idempotency.Response{}, err
	}
	created, err := NewRepository(transaction).Create(ctx, purchase)
	if err != nil {
		return idempotency.Response{}, err
	}
	reservation, err := Reserve(ctx, transaction, source, created)
	if err != nil {
		return idempotency.Response{}, err
	}
	for _, eventType := range []string{"PurchaseReserved", "PurchasePendingPayment"} {
		if err := outbox.Write(ctx, transaction, outbox.Event{
			AggregateType: "Purchase", AggregateID: created.ID, EventType: eventType,
			Payload: map[string]string{"purchase_id": created.ID}, OccurredAt: source.OccurredAt,
		}); err != nil {
			return idempotency.Response{}, err
		}
	}
	response, err := json.Marshal(publicPurchaseResponse(created, reservation, accessToken))
	if err != nil {
		return idempotency.Response{}, fmt.Errorf("marshal public purchase response: %w", err)
	}
	return idempotency.Response{Status: http.StatusCreated, Body: response}, nil
}

func persistRelationships(ctx context.Context, transaction *sql.Tx, relationships Relationships) (string, string, []PurchaseParticipantInput, error) {
	parties := identity.NewRepository(transaction)
	partyIDs := make(map[string]string, len(relationships.DeclaredParties))
	for _, declared := range relationships.DeclaredParties {
		created, err := parties.Create(ctx, declared.Party)
		if err != nil {
			return "", "", nil, err
		}
		partyIDs[declared.Ref] = created.ID
	}
	purchaserID, purchaserFound := partyIDs[relationships.PurchaserRef]
	payerID, payerFound := partyIDs[relationships.PayerRef]
	if !purchaserFound || !payerFound {
		return "", "", nil, ErrInvalidRelationships
	}
	participants := make([]PurchaseParticipantInput, 0, len(relationships.Participants))
	for _, participant := range relationships.Participants {
		input := PurchaseParticipantInput{DisplayName: participant.DisplayName}
		if participant.PartyRef != "" {
			partyID, found := partyIDs[participant.PartyRef]
			if !found {
				return "", "", nil, ErrInvalidRelationships
			}
			input.PartyID = &partyID
		}
		participants = append(participants, input)
	}
	return purchaserID, payerID, participants, nil
}

func newPublicPurchaseReference(eventYear int, random io.Reader) (string, error) {
	if eventYear < 1900 || eventYear > 9999 || random == nil {
		return "", ErrInvalidInput
	}
	var entropy [publicPurchaseReferenceBytes]byte
	if _, err := io.ReadFull(random, entropy[:]); err != nil {
		return "", fmt.Errorf("generate purchase reference: %w", err)
	}
	return "QRB-" + strconv.Itoa(eventYear) + "-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entropy[:]), nil
}

func newPublicPurchaseAccessToken(random io.Reader) (string, []byte, error) {
	if random == nil {
		return "", nil, ErrInvalidInput
	}
	var token [publicPurchaseAccessTokenBytes]byte
	if _, err := io.ReadFull(random, token[:]); err != nil {
		return "", nil, fmt.Errorf("generate purchase access token: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(token[:])
	hash := sha256.Sum256([]byte(encoded))
	return encoded, hash[:], nil
}

func publicPurchaseResponse(purchase Purchase, reservation Reservation, accessToken string) publicCreatePurchaseResponse {
	participants := make([]publicPurchaseParticipant, 0, len(purchase.Participants))
	for _, participant := range purchase.Participants {
		participants = append(participants, publicPurchaseParticipant{SequenceNo: participant.SequenceNo, DisplayName: participant.DisplayName})
	}
	response := publicCreatePurchaseResponse{}
	response.Data.Purchase = publicPurchase{
		ID: purchase.ID, PurchaseRef: purchase.PurchaseRef, Channel: purchase.Channel,
		Offering: publicPurchaseOffering{
			ID: purchase.OfferingID, Name: purchase.OfferingNameSnapshot, Kind: purchase.OfferingKindSnapshot,
			UnitPriceMinor: purchase.OfferingUnitPriceMinor, ParticipantCapacity: purchase.ParticipantCapacitySnapshot,
		},
		ParticipantCount: purchase.ParticipantCount, Participants: participants,
		TotalAmountMinor: purchase.TotalAmountMinor, CurrencyCode: purchase.CurrencyCode, Status: purchase.Status,
		ReservationExpiresAt: reservation.ExpiresAt.UTC(), CreatedAt: purchase.CreatedAt.UTC(),
	}
	response.Data.AccessToken = accessToken
	return response
}

func (handler *PublicHandler) clock() time.Time {
	if handler.now == nil {
		return time.Now().UTC()
	}
	return handler.now().UTC()
}

func (handler *PublicHandler) randomSource() io.Reader {
	if handler.random == nil {
		return rand.Reader
	}
	return handler.random
}

func decodePublicPurchase(c *gin.Context, destination any) bool {
	err := httpx.DecodeJSON(c.Writer, c.Request, destination, publicPurchaseBodyLimit)
	switch {
	case err == nil:
		return true
	case errors.Is(err, httpx.ErrRequestTooLarge):
		httpx.WriteError(c, http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", httpx.RequestID(c.Request.Context()), nil)
	case errors.Is(err, httpx.ErrUnsupportedMediaType):
		httpx.WriteError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json", httpx.RequestID(c.Request.Context()), nil)
	default:
		httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid JSON request", httpx.RequestID(c.Request.Context()), nil)
	}
	return false
}

func writePublicPurchaseError(c *gin.Context, logger *slog.Logger, err error) {
	requestID := httpx.RequestID(c.Request.Context())
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidRelationships), errors.Is(err, identity.ErrInvalidInput), errors.Is(err, ErrNotFound):
		httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "purchase validation failed", requestID, nil)
	case errors.Is(err, ErrStateConflict):
		httpx.WriteError(c, http.StatusConflict, "state_conflict", "purchase state conflicts with the request", requestID, nil)
	case errors.Is(err, ErrQuotaUnavailable):
		httpx.WriteError(c, http.StatusConflict, "quota_unavailable", "participant quota is unavailable", requestID, nil)
	case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
		httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
	case httpx.IsDependencyUnavailable(err):
		logPublicPurchaseError(logger, slog.LevelWarn, requestID, err)
		httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
	default:
		logPublicPurchaseError(logger, slog.LevelError, requestID, err)
		httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", requestID, nil)
	}
}

func logPublicPurchaseError(logger *slog.Logger, level slog.Level, requestID string, err error) {
	if logger == nil {
		return
	}
	logger.Log(context.Background(), level, "public purchase checkout failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
}
