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

// ListBoxMembersRequest ...
type ListBoxMembersRequest struct {
	boxID string
}

// BindAndValidate ...
func (req *ListBoxMembersRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

// ListBoxMembers ...
func (app *BoxApplication) ListBoxMembers(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListBoxMembersRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	membersIDs, err := events.ListBoxMemberIDs(ctx, app.DB, app.RedConn, req.boxID)
	if err != nil {
		return nil, merr.From(err).Desc("listing box members")
	}
	// transparent identity for admins to list members (they wanna know the identifier)
	isAdmin, err := events.IsAdmin(ctx, app.DB, req.boxID, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("checking admin")
	}
	return identityMapper.List(ctx, membersIDs, isAdmin)
}
