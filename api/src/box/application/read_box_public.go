package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

type ReadBoxPublicRequest struct {
	boxID          string
	OtherShareHash string `query:"other_share_hash"`
}

func (req *ReadBoxPublicRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	)
}

type PublicBoxView struct {
	Title   string            `json:"title"`
	Creator events.SenderView `json:"creator"`
}

// Since ReadBoxPublic returns public data, there is no access check performed
func (app *BoxApplication) ReadBoxPublic(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ReadBoxPublicRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// get key share
	keyShare, err := keyshares.Get(ctx, app.DB, req.OtherShareHash)
	if err != nil {
		return nil, err
	}
	if keyShare.BoxID != req.boxID {
		return nil, merror.Forbidden().Describe("wrong other share hash").Detail("other_share_hash", merror.DVInvalid)
	}

	// get box title
	box, err := boxes.Get(ctx, app.DB, identityMapper, req.boxID)
	if err != nil {
		return nil, err
	}

	view := PublicBoxView{
		Title:   box.Title,
		Creator: box.Creator,
	}
	return view, nil
}
