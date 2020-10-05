package events

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type Message struct {
	Encrypted       string
	PublicKey       string
	LastEditedAt    time.Time
	DeletedAt       time.Time
	FileID          null.String
	BoxID           string
	Type            string
	InitialSenderID string
	LastSenderID    string
	OldSize         int
	NewSize         int
}

func doMessage(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender has access to the box
	if err := MustMemberHaveAccess(ctx, exec, redConn, identities, e.BoxID, e.SenderID); err != nil {
		return nil, err
	}

	if e.ReferrerID.Valid {
		return nil, merror.BadRequest().Describe("referrer id cannot be set").Detail("referrer_id", merror.DVForbidden)
	}

	if err := e.persist(ctx, exec); err != nil {
		return nil, err
	}

	msg, err := BuildMessage(ctx, exec, e.ID)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func BuildMessage(ctx context.Context, exec boil.ContextExecutor, eventID string) (msg Message, err error) {
	// get message and referrers
	msgEvents, err := listEventAndReferrers(ctx, exec, eventID)
	if err != nil {
		return msg, err
	}

	// manage initial message
	e := msgEvents[0]
	msg.Type = e.Type
	msg.BoxID = e.BoxID
	msg.InitialSenderID = e.SenderID
	switch e.Type {
	case etype.Msgtext:
		var content MsgTextContent
		if err := e.JSONContent.Unmarshal(&content); err != nil {
			return msg, err
		}
		msg.Encrypted = content.Encrypted
		msg.PublicKey = content.PublicKey
		msg.NewSize = len(content.Encrypted)
	case etype.Msgfile:
		var content MsgFileContent
		if err := e.JSONContent.Unmarshal(&content); err != nil {
			return msg, err
		}
		file, err := files.Get(ctx, exec, content.EncryptedFileID)
		if err != nil {
			return msg, merror.Transform(err).Describe("getting file")
		}
		msg.NewSize = int(file.Size)
		msg.PublicKey = content.PublicKey
		msg.FileID = null.StringFrom(content.EncryptedFileID)
	default:
		return msg, merror.Forbidden().Describef("wrong event type %s", e.Type)
	}

	// if there are no modifiers, return
	if len(msgEvents) <= 1 {
		return msg, nil
	}

	// consider modifiers
Modifiers:
	for _, e := range msgEvents[1:] {
		msg.LastSenderID = e.SenderID
		switch e.Type {
		case etype.Msgdelete:
			msg.Encrypted = ""
			msg.OldSize = msg.NewSize
			msg.NewSize = 0
			msg.DeletedAt = e.CreatedAt
			// we don’t need to go further
			// as delete should be the last event
			break Modifiers
		case etype.Msgedit:
			var content MsgEditContent
			if err := e.JSONContent.Unmarshal(&content); err != nil {
				return msg, err
			}
			msg.Encrypted = content.NewEncrypted
			msg.PublicKey = content.NewPublicKey
			msg.LastEditedAt = e.CreatedAt
			msg.OldSize = msg.NewSize
			msg.NewSize = len(msg.Encrypted)
		default:
			return msg, merror.Forbidden().Describef("wrong event type %s", e.Type)
		}
	}

	return msg, nil
}
