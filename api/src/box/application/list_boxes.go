package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/boxes"
)

// ListBoxesRequest ...
type ListBoxesRequest struct {
	OwnerOrgID *string `query:"owner_org_id" json:"-"`
	Offset     int     `query:"offset" json:"-"`
	Limit      int     `query:"limit" json:"-"`
}

// BindAndValidate ...
func (req *ListBoxesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriQuery)
	}
	return v.ValidateStruct(req,
		v.Field(&req.OwnerOrgID, is.UUIDv4),
		v.Field(&req.Offset, v.Min(0)),
		v.Field(&req.Limit, v.Min(0)),
	)
}

// ListBoxes ...
func (app *BoxApplication) ListBoxes(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListBoxesRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}

	// default limit is 10
	if req.Limit == 0 {
		req.Limit = 10
	}

	// default owner org id is the self org
	if req.OwnerOrgID == nil {
		req.OwnerOrgID = &app.selfOrgID
	}

	boxes, err := boxes.ListBoxesForIdentity(
		ctx,
		app.DB, app.RedConn, identityMapper,
		acc.IdentityID, *req.OwnerOrgID,
		req.Limit, req.Offset,
	)
	if err != nil {
		return nil, merr.From(err).Desc("getting sender boxes")
	}

	return boxes, nil
}
