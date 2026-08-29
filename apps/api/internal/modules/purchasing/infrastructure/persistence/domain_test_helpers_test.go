package persistence

import purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"

type CreateInput = purchasingdomain.CreateInput
type PurchaseParticipantInput = purchasingdomain.PurchaseParticipantInput
type SnapshotPurchaseInput = purchasingdomain.SnapshotPurchaseInput
type Reservation = purchasingdomain.Reservation
type ReservationStatus = purchasingdomain.ReservationStatus

var NewPurchase = purchasingdomain.NewPurchase
var NewSnapshotPurchase = purchasingdomain.NewSnapshotPurchase

const (
    ChannelCommon             = purchasingdomain.ChannelCommon
    ReservationStatusReserved = purchasingdomain.ReservationStatusReserved
    ReservationStatusConsumed = purchasingdomain.ReservationStatusConsumed
    ReservationStatusReleased = purchasingdomain.ReservationStatusReleased
    ReservationStatusExpired  = purchasingdomain.ReservationStatusExpired
    ReservationHold           = purchasingdomain.ReservationHold
)

var ErrQuotaUnavailable = purchasingdomain.ErrQuotaUnavailable
