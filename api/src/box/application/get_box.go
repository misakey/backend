package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
)

// GetBoxRequest ...
type GetBoxRequest struct {
	boxID string
}

// BindAndValidate ...
func (req *GetBoxRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")

	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

// GetBox ...
func (app *BoxApplication) GetBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetBoxRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// check the box exists
	if err := events.MustBoxExists(ctx, app.DB, req.boxID); err != nil {
		return nil, merr.Forbidden()
	}

	// check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}

	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, acc.IdentityID); err != nil {
		// if the err is 403, the user is not a member, check if it has a least has access or not
		if merr.IsAForbidden(err) {
			// if user cannot join, return the no access error
			if err := events.MustBeAbleToJoin(ctx, app.DB, identityMapper, req.boxID, acc.IdentityID); err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	// get the box view
	return events.GetBoxView(ctx, app.DB, identityMapper, app.RedConn, req.boxID)
}

type GetOrgBoxRequest struct {
	orgID string
	boxID string
}

// BindAndValidate ...
func (req *GetOrgBoxRequest) BindAndValidate(eCtx echo.Context) error {
	req.orgID = eCtx.Param("oid")
	req.boxID = eCtx.Param("id")

	return v.ValidateStruct(req,
		v.Field(&req.orgID, v.Required, is.UUIDv4),
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

func (app *BoxApplication) GetOrgBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetOrgBoxRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// check the box exists
	if err := events.MustBoxExists(ctx, app.DB, req.boxID); err != nil {
		return nil, merr.Forbidden()
	}

	// check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	// authencated identity must be the organization
	if acc.IdentityID != req.orgID {
		return nil, merr.Forbidden()
	}

	return events.GetSimpleBox(ctx, app.DB, identityMapper, req.boxID)
}
