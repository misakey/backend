package events

import (
	"context"
	"encoding/json"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func BuildAggregate(ctx context.Context, exec boil.ContextExecutor, e *Event) error {
	// only msg.text and msg.file events can be aggregates
	if e.Type == "msg.text" || e.Type == "msg.file" {
		msg, err := buildMessage(ctx, exec, e.ID)
		if err != nil {
			return merror.Transform(err).Describef("building message %s", e.ID)
		}

		// if the message is deleted
		if !msg.DeletedAt.IsZero() {
			var content DeletedContent
			err := json.Unmarshal(e.JSONContent, &content)
			if err != nil {
				return merror.Transform(err).Describef("unmarshaling %s json", e.Type)
			}
			content.Deleted.ByIdentityID = msg.LastSenderID
			if err := e.JSONContent.Marshal(content); err != nil {
				return merror.Transform(err).Describef("marshalling %s content", e.Type)
			}
		} else if !msg.LastEditedAt.IsZero() {
			var content MsgTextContent
			err := json.Unmarshal(e.JSONContent, &content)
			if err != nil {
				return merror.Transform(err).Describef("unmarshaling %s json", e.Type)
			}
			content.Encrypted = msg.Encrypted
			content.PublicKey = msg.PublicKey
			content.LastEditedAt = null.TimeFrom(msg.LastEditedAt)
			if err := e.JSONContent.Marshal(content); err != nil {
				return merror.Transform(err).Describef("marshalling %s content", e.Type)
			}
		}
	}
	return nil
}
