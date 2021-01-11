package application

import (
	"context"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/boxes"
)

// CountBoxesRequest ...
type CountBoxesRequest struct {
}

// BindAndValidate ...
func (req *CountBoxesRequest) BindAndValidate(_ echo.Context) error {
	return nil
}

// CountBoxes ...
func (app *BoxApplication) CountBoxes(ctx context.Context, _ request.Request) (interface{}, error) {
	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}

	count, err := boxes.CountForSender(ctx, app.DB, app.RedConn, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("counting sender boxes")
	}

	return count, nil
}
