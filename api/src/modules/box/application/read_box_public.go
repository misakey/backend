package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/keyshares"
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
	Title string `json:"title"`
}

func (bs *BoxApplication) ReadBoxPublic(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*ReadBoxPublicRequest)

	// get key share
	keyShare, err := keyshares.Get(ctx, bs.db, req.OtherShareHash)
	if err != nil {
		return nil, err
	}
	if keyShare.BoxID != req.boxID {
		return nil, merror.Forbidden().Describe("wrong other share hash").Detail("other_share_hash", merror.DVInvalid)
	}

	// get box title
	box, err := boxes.Get(ctx, bs.db, bs.identities, req.boxID)
	if err != nil {
		return nil, err
	}

	view := PublicBoxView{
		Title: box.Title,
	}

	return view, nil
}
