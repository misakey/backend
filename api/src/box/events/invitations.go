package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// InviteIdentityIfPossible and do not return any error if not
func InviteIdentityIfPossible(
	ctx context.Context, cryptoRepo external.CryptoRepo, identityMapper *IdentityMapper,
	box Box, guest identity.Identity, senderID string, actionsDataJSON null.JSON,
) error {
	if guest.AccountID.IsZero() {
		return nil
	}

	guestView := senderViewFrom(guest)

	return CreateInvitationActions(ctx, cryptoRepo, identityMapper, box, guestView, senderID, actionsDataJSON, false)
}

// CreateInvitationAction for the given identity
func selectCryptoActionData(actionsData map[string]string, guestPubkeys identity.IdentityPublicKeys, nonIdentified bool) (action string, pubkey string, err error) {
	var encryptedCryptoAction string
	var present bool // just so that we don't have to use ":="
	var triedPubkeys = make([]string, 0)

	tryPubkey := func(actionsData map[string]string, pubkey string) *string {
		if pubkey == "" {
			return nil
		}

		triedPubkeys = append(triedPubkeys, pubkey)
		encryptedCryptoAction, present = actionsData[pubkey]
		if !present {
			return nil
		}

		return &encryptedCryptoAction
	}

	if nonIdentified {
		pubkey := guestPubkeys.NonIdentifiedPubkey.String
		if a := tryPubkey(actionsData, pubkey); a != nil {
			return *a, pubkey, nil
		}
		pubkey = guestPubkeys.NonIdentifiedPubkeyAesRsa.String
		if a := tryPubkey(actionsData, pubkey); a != nil {
			return *a, pubkey, nil
		}
	} else {
		pubkey = guestPubkeys.Pubkey.String
		if a := tryPubkey(actionsData, pubkey); a != nil {
			return *a, pubkey, nil
		}
		pubkey = guestPubkeys.PubkeyAesRsa.String
		if a := tryPubkey(actionsData, pubkey); a != nil {
			return *a, pubkey, nil
		}
	}

	if len(triedPubkeys) == 0 {
		return "", "", merr.Conflict().Desc("guest does not have a suitable public key")
	} else {
		return "", "", merr.BadRequest().
			Descf("missing encrypted crypto action for pubkey (tried %v)", triedPubkeys)
	}
}

// createInvitationAction for the given identity
// using the identity Pubkey (if nonIdentified is false)
// or the identity NonIdentifiedPubkey (if nonIdentified is true)
func CreateInvitationActions(
	ctx context.Context, cryptoRepo external.CryptoRepo, identityMapper *IdentityMapper,
	box Box, guest SenderView, senderID string, actionsDataJSON null.JSON,
	nonIdentified bool,
) error {
	// verify actions metadata
	var actionsData map[string]string
	err := actionsDataJSON.Unmarshal(&actionsData)
	if err != nil {
		return merr.From(err).Desc("unmarshalling actions data")
	}
	if len(actionsData) != 1 {
		return merr.BadRequest().Desc("required one entry per identity public key in extra")
	}

	// cryptoaction
	encryptedCryptoAction, pubkey, err := selectCryptoActionData(actionsData, guest.identityPubkeys, nonIdentified)
	if err != nil {
		return err
	}

	// prepare action
	actionID, err := uuid.NewString()
	if err != nil {
		return merr.From(err).Desc("generating action UUID")
	}
	action := crypto.Action{
		ID:                  actionID,
		AccountID:           guest.accountID.String,
		Type:                "invitation",
		SenderIdentityID:    null.StringFrom(senderID),
		BoxID:               null.StringFrom(box.ID),
		Encrypted:           encryptedCryptoAction,
		EncryptionPublicKey: pubkey,
		CreatedAt:           time.Now(),
	}

	// prepare notification
	notifDetailsBytes, err := json.Marshal(struct {
		BoxID          string `json:"box_id"`
		BoxTitle       string `json:"box_title"`
		OwnerOrgID     string `json:"owner_org_id"`
		CryptoActionID string `json:"cryptoaction_id"`
	}{
		BoxID:      box.ID,
		BoxTitle:   box.Title,
		OwnerOrgID: box.OwnerOrgID,

		CryptoActionID: action.ID,
	})
	if err != nil {
		return merr.From(err).Descf("marshalling notif details for pubkey \"%s\"", pubkey)
	}

	// create the prepared action
	if err := cryptoRepo.CreateActions(ctx, []crypto.Action{action}); err != nil {
		return err
	}

	// create the prepared notification
	identityMapper.CreateNotifs(ctx, []string{guest.ID}, "box.auto_invite", null.JSONFrom(notifDetailsBytes))
	return nil
}
