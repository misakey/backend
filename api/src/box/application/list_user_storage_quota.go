package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/quota"
)

// ListUserStorageQuotaRequest ...
type ListUserStorageQuotaRequest struct {
	IdentityID string
}

// BindAndValidate ...
func (req *ListUserStorageQuotaRequest) BindAndValidate(eCtx echo.Context) error {
	req.IdentityID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

// ListUserStorageQuota ...
func (app *BoxApplication) ListUserStorageQuota(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListUserStorageQuotaRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("id", merror.DVForbidden)
	}

	quota, err := quota.List(ctx, app.DB, req.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing user quota")
	}

	return quota, nil
}
