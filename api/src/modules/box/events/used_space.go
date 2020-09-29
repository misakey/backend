package events

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/quota"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type MetadataForUsedSpaceHandler struct {
	OldEventSize int64
	NewEventSize int64
}

func computeUsedSpace(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	var incrementValue int64
	var decrementValue int64 = 0
	var metadata = e.MetadataForHandlers

	if e.Type == "msg.text" {
		var content MsgTextContent
		err := json.Unmarshal(e.JSONContent, &content)
		if err != nil {
			return merror.Transform(err).Describe("unmarshaling content of event file")
		}
		incrementValue = int64(len(content.Encrypted))
	} else {
		incrementValue = metadata.NewEventSize
		decrementValue = metadata.OldEventSize
	}

	return quota.UpdateBoxUsedSpace(ctx, exec, e.BoxID, incrementValue, decrementValue)
}
