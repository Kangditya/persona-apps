package http

import (
    purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"
    "time"
)

type operationsPartySummary struct {
    ID          string `json:"id"`
    DisplayName string `json:"display_name"`
}

type operationsParticipantSnapshot struct {
    SequenceNo  int    `json:"sequence_no"`
    DisplayName string `json:"display_name"`
}

type operationsPurchase struct {
    ID                   string                          `json:"id"`
    EventID              string                          `json:"event_id"`
    PurchaseRef          string                          `json:"purchase_ref"`
    Channel              purchasingdomain.Channel        `json:"channel"`
    Purchaser            operationsPartySummary          `json:"purchaser"`
    Payer                *operationsPartySummary         `json:"payer,omitempty"`
    OfferingID           string                          `json:"offering_id"`
    OfferingNameSnapshot string                          `json:"offering_name_snapshot"`
    ParticipantCount     int                             `json:"participant_count"`
    Participants         []operationsParticipantSnapshot `json:"participants,omitempty"`
    TotalAmountMinor     int64                           `json:"total_amount_minor"`
    CurrencyCode         string                          `json:"currency_code"`
    Status               purchasingdomain.Status         `json:"status"`
    CreatedAt            time.Time                       `json:"created_at"`
    UpdatedAt            time.Time                       `json:"updated_at"`
}

type operationsPurchaseResponse struct {
    Data operationsPurchase `json:"data"`
}

type operationsPurchaseListResponse struct {
    Data []operationsPurchase `json:"data"`
    Page operationsPage       `json:"page"`
}

type operationsPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
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
    Channel              purchasingdomain.Channel    `json:"channel"`
    Offering             publicPurchaseOffering      `json:"offering"`
    ParticipantCount     int                         `json:"participant_count"`
    Participants         []publicPurchaseParticipant `json:"participants"`
    TotalAmountMinor     int64                       `json:"total_amount_minor"`
    CurrencyCode         string                      `json:"currency_code"`
    Status               purchasingdomain.Status     `json:"status"`
    ReservationExpiresAt time.Time                   `json:"reservation_expires_at"`
    CreatedAt            time.Time                   `json:"created_at"`
}

type publicCreatePurchaseResponse struct {
    Data struct {
        Purchase    publicPurchase `json:"purchase"`
        AccessToken string         `json:"access_token"`
    } `json:"data"`
}

func publicPurchaseResponse(purchase purchasingdomain.Purchase, reservation purchasingdomain.Reservation, accessToken string) publicCreatePurchaseResponse {
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

func toOperationsPurchase(detail purchasingdomain.Detail, includeParticipants bool) operationsPurchase {
    purchase := detail.Purchase
    result := operationsPurchase{
        ID: purchase.ID, EventID: purchase.EventID, PurchaseRef: purchase.PurchaseRef, Channel: purchase.Channel,
        Purchaser:  operationsPartySummary{ID: detail.Purchaser.ID, DisplayName: detail.Purchaser.DisplayName},
        OfferingID: purchase.OfferingID, OfferingNameSnapshot: purchase.OfferingNameSnapshot,
        ParticipantCount: purchase.ParticipantCount, TotalAmountMinor: purchase.TotalAmountMinor,
        CurrencyCode: purchase.CurrencyCode, Status: purchase.Status,
        CreatedAt: purchase.CreatedAt.UTC(), UpdatedAt: purchase.UpdatedAt.UTC(),
    }
    if detail.Payer != nil {
        result.Payer = &operationsPartySummary{ID: detail.Payer.ID, DisplayName: detail.Payer.DisplayName}
    }
    if includeParticipants {
        result.Participants = make([]operationsParticipantSnapshot, 0, len(purchase.Participants))
        for _, participant := range purchase.Participants {
            result.Participants = append(result.Participants, operationsParticipantSnapshot{SequenceNo: participant.SequenceNo, DisplayName: participant.DisplayName})
        }
    }
    return result
}
