package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// CountBoxesRequest ...
type CountBoxesRequest struct {
	OwnerOrgID *string `query:"owner_org_id" json:"-"`
}

// BindAndValidate ...
func (req *CountBoxesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriQuery)
	}
	return v.ValidateStruct(req,
		v.Field(&req.OwnerOrgID, is.UUIDv4),
	)
}

// CountBoxes ...
func (app *BoxApplication) CountBoxes(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CountBoxesRequest)

	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}

	// default owner org id is the self org
	if req.OwnerOrgID == nil {
		req.OwnerOrgID = &app.selfOrgID
	}

	count, err := events.CountBoxesForIdentity(ctx,
		app.DB, app.RedConn,
		acc.IdentityID, *req.OwnerOrgID,
	)
	if err != nil {
		return nil, merr.From(err).Desc("counting sender boxes")
	}

	return count, nil
}
