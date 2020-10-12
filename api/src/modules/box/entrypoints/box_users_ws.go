package entrypoints

import (
	"encoding/json"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

func (wh WebsocketHandler) BoxUsersWS(c echo.Context) error {

	// bind and validate
	identityID := c.Param("id")

	if err := v.Validate(identityID, v.Required, is.UUIDv4); err != nil {
		return merror.BadRequest().From(merror.OriPath).Detail("id", merror.DVMalformed)
	}

	// check accesses
	acc := oidc.GetAccesses(c.Request().Context())
	if acc == nil || acc.IdentityID != identityID {
		return merror.Forbidden()
	}

	return wh.RedisListener(
		c,
		fmt.Sprintf("user_%s", acc.IdentityID),
		fmt.Sprintf("user_%s:ws", identityID),
		boxUsersHandler,
	)
}

type WSMessage struct {
	Type   string          `json:"type"`
	Object json.RawMessage `json:"object"`
}

type AckObject struct {
	SenderID string `json:"sender_id"`
	BoxID    string `json:"box_id"`
}

func boxUsersHandler(c echo.Context, wh WebsocketHandler, receivedMsg []byte) error {
	message := WSMessage{}
	if err := json.Unmarshal(receivedMsg, &message); err != nil {
		return err
	}
	if message.Type == "ack" {
		obj := AckObject{}
		if err := json.Unmarshal(message.Object, &obj); err != nil {
			return err
		}
		if err := events.DelCounts(c.Request().Context(), wh.boxService.RedConn, obj.SenderID, obj.BoxID); err != nil {
			return err
		}
	}
	return nil
}
