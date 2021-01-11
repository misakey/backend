package events

import (
	"context"
	"sync"

	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// IdentityMapper ...
type IdentityMapper struct {
	sync.Mutex

	querier external.IdentityRepo

	mem map[string]SenderView
}

// NewIdentityMapper ...
func NewIdentityMapper(querier external.IdentityRepo) *IdentityMapper {
	return &IdentityMapper{
		querier: querier,
		mem:     make(map[string]SenderView),
	}
}

// Get the identity considering the ID
// transparent set to true will return all the information about the identity
// where a false value will remove some information from the identity considering their privacy
// use true for internal checks or exceptional business cases.
func (mapper *IdentityMapper) Get(ctx context.Context, identityID string, transparent bool) (SenderView, error) {
	var sender SenderView
	var ok bool
	sender, ok = mapper.mem[identityID]
	if !ok {
		// get unknown identity and save it
		existingIdentity, err := mapper.querier.Get(ctx, identityID)
		// NOTE: on not found, the system still fills the SenderView with anonymous information
		if err != nil && !merr.IsANotFound(err) {
			return sender, merr.From(err).Desc("getting identity")
		}

		if err == nil {
			sender = senderViewFrom(existingIdentity)
		} else { // it is a not found identity
			sender = anonymousSenderView()
		}

		mapper.Lock()
		mapper.mem[existingIdentity.ID] = sender
		mapper.Unlock()
	}

	if !transparent {
		return sender.copyOpaque(), nil
	}
	return sender, nil
}

// List the identities considering IDs
// transparent set to true will return all the information about the identity
// where a false value will remove some information from the identity considering their privacy
// use true for internal checks or exceptional business cases.
func (mapper *IdentityMapper) List(ctx context.Context, identityIDs []string, transparent bool) ([]SenderView, error) {
	// compute unknowns
	var unknownIDs []string
	for _, identityID := range identityIDs {
		_, ok := mapper.mem[identityID]
		if !ok {
			unknownIDs = append(unknownIDs, identityID)
		}
	}

	if len(unknownIDs) > 0 {
		// get all unknowns and save them
		identities, err := mapper.querier.List(ctx, identity.Filters{IDs: unknownIDs})
		if err != nil {
			return nil, merr.From(err).Desc("listing identities")
		}

		// put them in memory
		mapper.Lock()
		for _, identity := range identities {
			mapper.mem[identity.ID] = senderViewFrom(*identity)
		}
		// if any identities has not been found, set an anonymous view
		for _, unknownID := range unknownIDs {
			_, ok := mapper.mem[unknownID]
			if !ok {
				mapper.mem[unknownID] = anonymousSenderView()
			}
		}
		mapper.Unlock()
	}

	views := make([]SenderView, len(identityIDs))
	for idx, identityID := range identityIDs {
		// should always be in map there considering code above
		sender := mapper.mem[identityID]
		if transparent {
			views[idx] = sender
		} else {
			views[idx] = sender.copyOpaque()
		}
	}
	return views, nil
}

// CreateNotifs for identity
func (mapper *IdentityMapper) CreateNotifs(ctx context.Context, identityIDs []string, nType string, details null.JSON) {
	if err := mapper.querier.NotificationBulkCreate(ctx, identityIDs, nType, details); err != nil {
		logger.FromCtx(ctx).Err(err).Msgf("creating %v notifs", identityIDs)
	}
}

// MapToAccountID ...
func (mapper *IdentityMapper) MapToAccountID(ctx context.Context, identityIDs []string) (map[string]string, error) {
	identities, err := mapper.querier.List(ctx, identity.Filters{IDs: identityIDs})
	if err != nil {
		return nil, merr.From(err).Desc("listing identities")
	}

	result := make(map[string]string)
	for _, existing := range identities {
		accountID := existing.AccountID.String
		if existing.AccountID.Valid {
			result[existing.ID] = accountID
		}
	}
	return result, nil
}

func senderViewFrom(identity identity.Identity) SenderView {
	sender := SenderView{
		ID:           identity.ID,
		IdentifierID: identity.IdentifierID,
		DisplayName:  identity.DisplayName,
		AvatarURL:    identity.AvatarURL,
	}
	sender.Identifier.Value = identity.Identifier.Value
	sender.Identifier.Kind = string(identity.Identifier.Kind)
	return sender
}

func anonymousSenderView() SenderView {
	sender := SenderView{
		ID:           "anonymous-user",
		IdentifierID: "anonymous-user",
		DisplayName:  "Anonymous User",
	}
	return sender
}
