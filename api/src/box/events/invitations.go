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
)

// createInvitationAction for the given identity
// using the identity Pubkey (if nonIdentified is false)
// or the identity NonIdentifiedPubkey (if nonIdentified is true)
func createInvitationActions(
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

	// check and assign pubkeys
	var pubkey string
	if !nonIdentified && guest.pubkey.Valid {
		pubkey = guest.pubkey.String
	} else if nonIdentified && guest.nonIdentifiedPubkey.Valid {
		pubkey = guest.nonIdentifiedPubkey.String
	} else {
		return merr.Conflict().Desc("guest does not have a public key")
	}

	// cryptoaction
	encryptedCryptoAction, present := actionsData[pubkey]
	if !present {
		return merr.BadRequest().Descf("missing encrypted crypto action for pubkey \"%s\"", pubkey)
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
