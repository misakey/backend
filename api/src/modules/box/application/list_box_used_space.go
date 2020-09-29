package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/quota"
)

type ListBoxUsedSpaceRequest struct {
	// json tag is needed as without it BindAndValidate does not return the right Detail
	IdentityID string `query:"identity_id" json:"identity_id"`
}

func (req *ListBoxUsedSpaceRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriQuery)
	}
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) ListBoxUsedSpace(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListBoxUsedSpaceRequest)

	access := ajwt.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}

	creates, err := events.ListCreatorIDEvents(ctx, bs.DB, req.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing creator box ids")
	}
	boxIDs := make([]string, len(creates))
	for idx, event := range creates {
		boxIDs[idx] = event.BoxID
	}

	return quota.ListBoxUsedSpaces(ctx, bs.DB, boxIDs)
}