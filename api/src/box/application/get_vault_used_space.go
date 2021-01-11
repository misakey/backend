package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/quota"
)

// GetVaultUsedSpaceRequest ...
type GetVaultUsedSpaceRequest struct {
	IdentityID string
}

// BindAndValidate ...
func (req *GetVaultUsedSpaceRequest) BindAndValidate(eCtx echo.Context) error {
	req.IdentityID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

// GetVaultUsedSpace ...
func (app *BoxApplication) GetVaultUsedSpace(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetVaultUsedSpaceRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merr.Unauthorized()
	}
	if req.IdentityID != access.IdentityID {
		return nil, merr.Forbidden().Add("id", merr.DVForbidden)
	}

	vaultSpace, err := quota.GetVault(ctx, app.DB, req.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("get vault space")
	}

	return vaultSpace, nil
}
