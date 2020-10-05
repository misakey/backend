package events

import (
	"context"
	"encoding/json"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

// MsgEditContent is exported
// because application layer need to access the "EventID" field
type MsgEditContent struct {
	NewEncrypted string `json:"new_encrypted"`
	NewPublicKey string `json:"new_public_key"`
}

func (c *MsgEditContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c MsgEditContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.NewEncrypted, v.Required, is.Base64),
		v.Field(&c.NewEncrypted, v.Required), // URL-safe base64
	)
}

func doEditMsg(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender has access to the box
	if err := MustMemberHaveAccess(ctx, exec, redConn, identities, e.BoxID, e.SenderID); err != nil {
		return nil, err
	}

	// check that the event contains a referrer_id
	if err := checkReferrer(ctx, *e); err != nil {
		return nil, err
	}

	var content MsgEditContent
	if err := json.Unmarshal(e.JSONContent, &content); err != nil {
		return nil, merror.Internal().Describe("unmarshaling content json")
	}

	msg, err := BuildMessage(ctx, exec, e.ReferrerID.String)
	if err != nil {
		return nil, merror.Transform(err).Describe("building message")
	}

	if e.SenderID != msg.InitialSenderID {
		return nil, merror.Forbidden().Describe("can only edit own messages")
	}

	if !msg.DeletedAt.IsZero() {
		return nil, merror.Gone().Describe("cannot edit a deleted event")
	}

	// update the message
	msg.Encrypted = content.NewEncrypted
	msg.PublicKey = content.NewPublicKey
	msg.LastEditedAt = e.CreatedAt
	msg.OldSize = msg.NewSize
	msg.NewSize = len(content.NewEncrypted)

	if err := e.persist(ctx, exec); err != nil {
		return nil, err
	}

	return &msg, nil
}
