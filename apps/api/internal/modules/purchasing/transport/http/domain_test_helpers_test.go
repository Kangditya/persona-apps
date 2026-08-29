package http

import (
    purchasingapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/application"
    purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"
)

type Purchase = purchasingdomain.Purchase
type Participant = purchasingdomain.Participant
type PartySummary = purchasingdomain.PartySummary
type Detail = purchasingdomain.Detail
type ListInput = purchasingdomain.ListInput
type Reservation = purchasingdomain.Reservation
type CreateInput = purchasingdomain.CreateInput
type PurchaseParticipantInput = purchasingdomain.PurchaseParticipantInput
type Channel = purchasingdomain.Channel
type Status = purchasingdomain.Status

const (
    ChannelCommon                  = purchasingdomain.ChannelCommon
    StatusPendingPayment           = purchasingdomain.StatusPendingPayment
    publicPurchaseReferenceBytes   = 10
    publicPurchaseAccessTokenBytes = 32
)

var newPublicPurchaseReference = purchasingapplication.NewPurchaseReference
var newPublicPurchaseAccessToken = purchasingapplication.NewPurchaseAccessToken
var NewPurchase = purchasingdomain.NewPurchase

var (
    ErrInvalidInput         = purchasingdomain.ErrInvalidInput
    ErrInvalidRelationships = purchasingdomain.ErrInvalidRelationships
)
