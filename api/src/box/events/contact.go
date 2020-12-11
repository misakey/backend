package events

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

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

// CreateContactBox creates the box, sets the `invitation_link` access and invite the contacted user
func CreateContactBox(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identityMapper *IdentityMapper, filesRepo files.FileStorageRepo, cryptoRepo external.CryptoRepo, contact ContactBox) (*Box, error) {

	// get contacted user identity
	contactedUser, err := identityMapper.querier.Get(ctx, contact.ContactedIdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting contacted user identity")
	}

	if contactedUser.AccountID.IsZero() {
		return nil, merror.Forbidden().Describe("contacted user must have an account_id").Detail("contacted_account", merror.DVRequired)
	}

	// create contact box
	event, err := CreateCreateEvent(
		ctx,
		contact.Title,
		contact.PublicKey,
		contact.IdentityID,
		exec,
		redConn,
		identityMapper,
		filesRepo,
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating create event")
	}

	if err := keyshares.Create(
		ctx, exec,
		contact.OtherShareHash,
		contact.Share,
		contact.EncryptedInvitationKeyShare,
		event.BoxID,
		contact.IdentityID,
	); err != nil {
		return nil, merror.Transform(err).Describe("creating key share")
	}

	// create access event
	eventID, err := uuid.NewString()
	if err != nil {
		return nil, merror.Transform(err).Describe("generating access event id")
	}

	accessContent := accessContent{
		RestrictionType: "invitation_link",
		Value:           contact.OtherShareHash,
		AutoInvite:      false,
	}

	serializedContent, err := json.Marshal(accessContent)
	if err != nil {
		return nil, merror.Transform(err).Describe("serializing access content")
	}

	accessEvent := Event{
		ID:          eventID,
		BoxID:       event.BoxID,
		SenderID:    contact.IdentityID,
		Type:        etype.Accessadd,
		JSONContent: serializedContent,
	}

	decodedInvitationDataJSON, err := contact.InvitationDataJSON.MarshalJSON()
	if err != nil {
		return nil, merror.Transform(err).Describe("decoding json")
	}
	if _, err := doAddAccess(ctx, &accessEvent, null.JSON{}, exec, redConn, identityMapper, cryptoRepo, nil); err != nil {
		return nil, merror.Transform(err).Describe("creating access event")
	}

	if err := cryptoRepo.CreateInvitationActionsForIdentity(ctx, contact.IdentityID, event.BoxID, contact.Title, contact.ContactedIdentityID, null.JSONFrom(decodedInvitationDataJSON)); err != nil {
		return nil, merror.Transform(err).Describe("creating invitation action")
	}

	// compute box
	box, err := Compute(ctx, event.BoxID, exec, identityMapper, &event)
	if err != nil {
		return nil, merror.Transform(err).Describe("computing box")
	}

	return &box, nil

}
