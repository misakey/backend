package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

// ContactBox model
type ContactBox struct {
	// user info
	ContactedIdentityID string
	IdentityID          string

	// box info
	Title     string
	PublicKey string

	// contact info
	OtherShareHash              string
	Share                       string
	EncryptedInvitationKeyShare string
	InvitationDataJSON          types.JSON
}

// CreateContactBox creates the box, sets the access mode to public and invite the contacted user
func CreateContactBox(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identityMapper *IdentityMapper, _ files.FileStorageRepo, cryptoRepo external.CryptoRepo, contact ContactBox) (*Box, error) {

	// get contacted user identity
	contactedUser, err := identityMapper.querier.Get(ctx, contact.ContactedIdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting contacted user identity")
	}

	if contactedUser.AccountID.IsZero() {
		return nil, merr.Forbidden().Desc("contacted user must have an account_id").Add("contacted_account", merr.DVRequired)
	}

	// create contact box
	event, err := CreateCreateEvent(
		ctx,
		exec, redConn, identityMapper,
		contact.Title, contact.PublicKey, contact.IdentityID,
	)
	if err != nil {
		return nil, merr.From(err).Desc("creating create event")
	}

	if err := keyshares.Create(
		ctx, exec,
		contact.OtherShareHash,
		contact.Share,
		contact.EncryptedInvitationKeyShare,
		event.BoxID,
		contact.IdentityID,
	); err != nil {
		return nil, merr.From(err).Desc("creating key share")
	}

	// set the box in public mode
	accessModeEvent, err := newWithAnyContent(
		etype.Stateaccessmode,
		&AccessModeContent{
			Value: PublicMode,
		},
		event.BoxID,
		contact.IdentityID,
		nil,
	)
	if err != nil {
		return nil, merr.From(err).Desc("newing access mode event")
	}
	if _, err := doAddAccess(ctx, &accessModeEvent, null.JSON{}, exec, redConn, identityMapper, cryptoRepo, nil); err != nil {
		return nil, merr.From(err).Desc("creating access event")
	}

	// create crypto invitation actions
	decodedInvitationDataJSON, err := contact.InvitationDataJSON.MarshalJSON()
	if err != nil {
		return nil, merr.From(err).Desc("decoding json")
	}
	if err := cryptoRepo.CreateInvitationActionsForIdentity(ctx, contact.IdentityID, event.BoxID, contact.Title, contact.ContactedIdentityID, null.JSONFrom(decodedInvitationDataJSON)); err != nil {
		return nil, merr.From(err).Desc("creating invitation action")
	}

	// compute box to return it
	box, err := Compute(ctx, event.BoxID, exec, identityMapper, &event)
	if err != nil {
		return nil, merr.From(err).Desc("computing box")
	}

	return &box, nil

}
