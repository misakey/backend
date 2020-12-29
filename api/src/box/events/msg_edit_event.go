package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// MsgEditContent is exported
// because application layer need to access the "EventID" field
type MsgEditContent struct {
	NewEncrypted string `json:"new_encrypted"`
	NewPublicKey string `json:"new_public_key"`
}

// Unmarshal a msg.edit content JSON into its typed structure
func (c *MsgEditContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

// Validate a msg.edit content structure
func (c MsgEditContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.NewEncrypted, v.Required, is.Base64),
		v.Field(&c.NewEncrypted, v.Required), // URL-safe base64
	)
}

func doEditMsg(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, _ *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender is a member of the box
	if err := MustBeMember(ctx, exec, redConn, e.BoxID, e.SenderID); err != nil {
		return nil, err
	}

	// check that the event contains a referrer_id
	if err := checkReferrer(*e); err != nil {
		return nil, err
	}

	msg, err := buildMessage(ctx, exec, e.ReferrerID.String)
	if err != nil {
		return nil, merror.Transform(err).Describe("building message")
	}

	if e.SenderID != msg.InitialSenderID {
		return nil, merror.Forbidden().Describe("can only edit own messages")
	}
	// if the message is already deleted, do not go further
	if !msg.DeletedAt.IsZero() {
		return nil, merror.Gone().Describe("cannot edit a deleted message")
	}
	if msg.Type == etype.Msgfile {
		return msg, merror.Forbidden().Describef("cannot edit event type %s", msg.Type)
	}

	if err := e.persist(ctx, exec); err != nil {
		return nil, err
	}

	// add the recently created event to the built message
	if err := msg.addEvent(*e); err != nil {
		return nil, err
	}
	return &msg, nil
}
