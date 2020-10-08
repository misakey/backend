package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type ListBoxMembersRequest struct {
	boxID string
}

func (req *ListBoxMembersRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) ListBoxMembers(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListBoxMembersRequest)

	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustMemberHaveAccess(ctx, bs.DB, bs.RedConn, bs.Identities, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	membersIDs, err := events.ListBoxMemberIDs(ctx, bs.DB, bs.RedConn, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing box members")
	}

	members := make([]events.SenderView, len(membersIDs))
	i := 0
	for _, id := range membersIDs {
		identity, err := bs.Identities.Get(ctx, id)
		if err != nil {
			return nil, merror.Transform(err).Describe("getting identity")
		}

		members[i] = events.NewSenderView(identity)
		i += 1
	}

	return members, nil
}
