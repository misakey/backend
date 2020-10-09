package events

import (
	"context"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

func buildMessage(ctx context.Context, exec boil.ContextExecutor, eventID string) (msg Message, err error) {
	// get message and referrers
	msgEvents, err := listEventAndReferrers(ctx, exec, eventID)
	if err != nil {
		return msg, err
	}

	// manage initial message
	initialEvent := msgEvents[0]
	msg.Type = initialEvent.Type
	msg.BoxID = initialEvent.BoxID
	msg.InitialSenderID = initialEvent.SenderID
	switch initialEvent.Type {
	case etype.Msgtext:
		var content MsgTextContent
		if err := initialEvent.JSONContent.Unmarshal(&content); err != nil {
			return msg, err
		}
		msg.Encrypted = content.Encrypted
		msg.PublicKey = content.PublicKey
		msg.NewSize = len(content.Encrypted)
	case etype.Msgfile:
		var content MsgFileContent
		if err := initialEvent.JSONContent.Unmarshal(&content); err != nil {
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
		return msg, merror.Forbidden().Describef("wrong initial event type %s", initialEvent.Type)
	}

	// if there are no modifiers, return
	if len(msgEvents) <= 1 {
		return msg, nil
	}

	// consider modifiers
	for _, e := range msgEvents[1:] {
		if err := msg.addEvent(ctx, e); err != nil {
			return msg, err
		}
	}
	return msg, nil
}

func (msg *Message) addEvent(ctx context.Context, e Event) error {
	msg.LastSenderID = e.SenderID
	switch e.Type {
	case etype.Msgdelete:
		msg.Encrypted = ""
		msg.OldSize = msg.NewSize
		msg.NewSize = 0
		msg.DeletedAt = e.CreatedAt
	case etype.Msgedit:
		var content MsgEditContent
		if err := e.JSONContent.Unmarshal(&content); err != nil {
			return err
		}
		msg.Encrypted = content.NewEncrypted
		msg.PublicKey = content.NewPublicKey
		msg.LastEditedAt = e.CreatedAt
		msg.OldSize = msg.NewSize
		msg.NewSize = len(msg.Encrypted)
	default:
		return merror.Forbidden().Describef("wrong referrer event type %s", e.Type)
	}
	return nil
}
