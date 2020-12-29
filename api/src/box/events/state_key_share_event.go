package events

import (
	"context"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"

	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

func doStateKeyShare(ctx context.Context, e *Event, extraJSON null.JSON, exec boil.ContextExecutor, redConn *redis.Client, identityMapper *IdentityMapper, cryptoActionService external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check accesses
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	// there is no "content" for this kind of event,
	// only "extra"

	extra := struct {
		MisakeyShare                string `json:"misakey_share"`
		OtherShareHash              string `json:"other_share_hash"`
		EncryptedInvitationKeyShare string `json:"encrypted_invitation_key_share"`
	}{}

	if err := extraJSON.Unmarshal(&extra); err != nil {
		return nil, merror.Transform(err).Describe("unmarshalling \"for server no store\"")
	}
	if err := v.ValidateStruct(&extra,
		v.Field(&extra.MisakeyShare, v.Required),
		v.Field(&extra.OtherShareHash, v.Required),
		v.Field(&extra.EncryptedInvitationKeyShare, v.Required),
	); err != nil {
		return nil, merror.Transform(err).Describe("validating \"for server no store\"")
	}

	err := keyshares.EmptyAll(ctx, exec, e.BoxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("deleting previous key shares")
	}

	if err = keyshares.Create(
		ctx, exec,
		extra.OtherShareHash, extra.MisakeyShare, extra.EncryptedInvitationKeyShare,
		e.BoxID, e.SenderID,
	); err != nil {
		return nil, merror.Transform(err).Describe("creating key share")
	}

	// Creation of crypto actions
	// Note: unlike with auto-invitations,
	// here *every ACR2 member* of the box can receive the cryptoaction.
	// this because the encrypted payload of the cryptoaction
	// is decrypted with the box secret key (which all members should have)
	// instead of an identity key like in auto-invitations
	// (users being invited to box don't have the box secret key yet).

	boxPublicKey, err := GetBoxPublicKey(ctx, exec, e.BoxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting box public key")
	}

	membersIdentityID, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing box members IDs")
	}

	// note that identities that don't have an account (ACR1 identities)
	// will not be present in the "accountIDs" mapping
	accountIDsByIdentityID, err := identityMapper.MapToAccountID(ctx, membersIdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting members identities")
	}

	// it would be tempting to set an initial size for the crypto action slice,
	// but if there is a single one that is left empty
	// it's going to mess everything up,
	// and we don't really know how many crypto actions will be needed
	var cryptoActions []crypto.Action

	// set for uniqueness
	processedAccounts := make(map[string]bool, len(accountIDsByIdentityID)-1)

	for identityID, accountID := range accountIDsByIdentityID {
		_, processed := processedAccounts[accountID]
		if !processed && identityID != e.SenderID {
			action := crypto.Action{
				AccountID:           accountID,
				Type:                "set_box_key_share",
				SenderIdentityID:    null.StringFrom(e.SenderID),
				BoxID:               null.StringFrom(e.BoxID),
				Encrypted:           extra.EncryptedInvitationKeyShare,
				EncryptionPublicKey: boxPublicKey,
				CreatedAt:           time.Now(),
			}
			cryptoActions = append(cryptoActions, action)
			processedAccounts[accountID] = true
		}
	}

	err = cryptoActionService.CreateActions(ctx, cryptoActions)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating crypto actions")
	}

	return nil, e.persist(ctx, exec)

}
