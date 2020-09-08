package application

import (
	"context"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
)

type CountBoxesRequest struct {
}

func (req *CountBoxesRequest) BindAndValidate(_ echo.Context) error {
	return nil
}

func (bs *BoxApplication) CountBoxes(ctx context.Context, _ entrypoints.Request) (interface{}, error) {
	// retrieve accesses to filters boxes to return
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	count, err := boxes.CountForSender(ctx, bs.db, acc.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("counting sender boxes")
	}

	return count, nil
}
