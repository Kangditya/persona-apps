package persistence

import (
    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
)

type Offering = offeringdomain.Offering
type CreateInput = offeringdomain.CreateInput
type UpdateInput = offeringdomain.UpdateInput
type PublishInput = offeringdomain.PublishInput
type TransitionInput = offeringdomain.TransitionInput
type Mutation = offeringdomain.Mutation
type ListInput = offeringdomain.ListInput

var Create = offeringdomain.Create
var Update = offeringdomain.Update
var Publish = offeringdomain.Publish
var MarkUnavailable = offeringdomain.MarkUnavailable

const (
    StatusDraft            = offeringdomain.StatusDraft
    MaxSafeInteger         = offeringdomain.MaxSafeInteger
    MaxParticipantCapacity = offeringdomain.MaxParticipantCapacity
)

var (
    ErrDuplicateCode = offeringdomain.ErrDuplicateCode
    ErrInvalidCursor = offeringdomain.ErrInvalidCursor
    ErrStaleVersion  = offeringdomain.ErrStaleVersion
    ErrNotFound      = offeringdomain.ErrNotFound
)

type Event = eventdomain.Event

func int64Pointer(value int64) *int64 {
    return &value
}
