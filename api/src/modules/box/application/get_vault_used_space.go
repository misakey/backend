package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/quota"
)

type GetVaultUsedSpaceRequest struct {
	IdentityID string
}

func (req *GetVaultUsedSpaceRequest) BindAndValidate(eCtx echo.Context) error {
	req.IdentityID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) GetVaultUsedSpace(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetVaultUsedSpaceRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("id", merror.DVForbidden)
	}

	vaultSpace, err := quota.GetVault(ctx, bs.DB, req.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("get vault space")
	}

	return vaultSpace, nil
}
