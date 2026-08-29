package http

import (
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
)

type Offering = offeringdomain.Offering
type ListInput = offeringdomain.ListInput
type CatalogueResult = offeringdomain.CatalogueResult
type CatalogueOffering = offeringdomain.CatalogueOffering
type offeringCursor = offeringdomain.Cursor

var encodeCursor = offeringdomain.EncodeCursor

const (
    DefaultListLimit = offeringdomain.DefaultListLimit
    StatusPublished  = offeringdomain.StatusPublished
)

var (
    ErrNoChanges = offeringdomain.ErrNoChanges
    ErrNotFound  = offeringdomain.ErrNotFound
)
