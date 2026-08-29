package persistence

import identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"

type Party = identitydomain.Party
type CreateInput = identitydomain.CreateInput

var NewParty = identitydomain.NewParty

const (
    PartyTypePerson       = identitydomain.PartyTypePerson
    PartyTypeOrganization = identitydomain.PartyTypeOrganization
)

var ErrNotFound = identitydomain.ErrNotFound
