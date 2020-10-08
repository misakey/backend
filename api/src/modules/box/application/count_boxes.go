package application

import (
	"context"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
)

type CountBoxesRequest struct {
}

func (req *CountBoxesRequest) BindAndValidate(_ echo.Context) error {
	return nil
}

func (bs *BoxApplication) CountBoxes(ctx context.Context, _ request.Request) (interface{}, error) {
	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	count, err := boxes.CountForSender(ctx, bs.DB, bs.RedConn, acc.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("counting sender boxes")
	}

	return count, nil
}
