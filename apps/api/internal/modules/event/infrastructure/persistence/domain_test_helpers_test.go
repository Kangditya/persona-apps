package persistence

import eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"

type Event = eventdomain.Event
type CreateInput = eventdomain.CreateInput
type UpdateInput = eventdomain.UpdateInput
type TransitionInput = eventdomain.TransitionInput
type ListInput = eventdomain.ListInput

var Create = eventdomain.Create
var Update = eventdomain.Update
var Publish = eventdomain.Publish
var Activate = eventdomain.Activate

const (
    StatusDraft    = eventdomain.StatusDraft
    MaxSafeInteger = eventdomain.MaxSafeInteger
)

var (
    ErrDuplicateYear  = eventdomain.ErrDuplicateYear
    ErrInvalidCursor  = eventdomain.ErrInvalidCursor
    ErrStaleVersion   = eventdomain.ErrStaleVersion
    ErrActiveConflict = eventdomain.ErrActiveConflict
    ErrNotFound       = eventdomain.ErrNotFound
)
