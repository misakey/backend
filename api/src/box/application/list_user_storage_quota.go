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
		return nil, merr.Unauthorized()
	}
	if req.IdentityID != access.IdentityID {
		return nil, merr.Forbidden().Add("id", merr.DVForbidden)
	}

	userStorageQuota, err := quota.List(ctx, app.DB, req.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("listing user quota")
	}

	// on no base quota found, it means the identity is new and no base quota has been created yet
	// create then the base quota
	if noBaseFound(userStorageQuota) {
		baseQuotum := quota.Quotum{
			Origin:     "base",
			IdentityID: req.IdentityID,
			Value:      104857600, // default value for newcomers
		}
		if err := quota.Create(ctx, app.DB, &baseQuotum); err != nil {
			return nil, merr.From(err).Desc("creating base quota")
		}
	}

	return userStorageQuota, nil
}

// noBaseFound returns true if no quotum with origin base is found in the received slice
func noBaseFound(userQuota []quota.Quotum) bool {
	for _, quotum := range userQuota {
		if quotum.Origin == "base" {
			return false
		}
	}
	return true
}
