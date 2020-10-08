package entrypoints

import (
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

func (wh WebsocketHandler) ListEventsWS(c echo.Context) error {

	// bind and validate
	boxID := c.Param("id")

	if err := v.Validate(boxID, v.Required, is.UUIDv4); err != nil {
		return merror.BadRequest().From(merror.OriPath).Detail("id", merror.DVMalformed)
	}

	// check accesses
	acc := oidc.GetAccesses(c.Request().Context())
	if acc == nil {
		return merror.Forbidden()
	}
	if err := events.MustMemberHaveAccess(
		c.Request().Context(),
		wh.boxService.DB,
		wh.boxService.RedConn,
		wh.boxService.Identities,
		boxID,
		acc.IdentityID,
	); err != nil {
		return err
	}

	return wh.RedisListener(
		c,
		fmt.Sprintf("%s_%s", acc.IdentityID, boxID),
		fmt.Sprintf("%s:events", boxID),
		fmt.Sprintf("interrupt:%s:%s", boxID, acc.IdentityID),
		func(_ echo.Context, _ WebsocketHandler, _ []byte) error { return nil },
	)
}
