package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

type CreateBoxRequest struct {
	PublicKey    string `json:"public_key"`
	Title        string `json:"title"`
	KeyShareData *struct {
		OtherShareHash              string      `json:"other_share_hash"`
		Share                       string      `json:"misakey_share"`
		EncryptedInvitationKeyShare null.String `json:"encrypted_invitation_key_share"`
	} `json:"key_share"`
}

func (req *CreateBoxRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	err := v.ValidateStruct(req,
		v.Field(&req.PublicKey, v.Required),
		v.Field(&req.Title, v.Required, v.Length(1, 50)),
	)
	if err != nil {
		return err
	}

	if req.KeyShareData != nil {
		err := v.ValidateStruct(req.KeyShareData,
			v.Field(&req.KeyShareData.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
			v.Field(&req.KeyShareData.Share, v.Required, is.Base64),
			v.Field(&req.KeyShareData.EncryptedInvitationKeyShare, v.Required, is.Base64),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *BoxApplication) CreateBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateBoxRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	event, err := events.CreateCreateEvent(
		ctx,
		req.Title,
		req.PublicKey,
		acc.IdentityID,
		app.DB,
		app.RedConn,
		identityMapper,
		app.filesRepo,
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating create event")
	}

	// build the box view and return it
	box, err := events.Compute(ctx, event.BoxID, app.DB, identityMapper, &event)
	if err != nil {
		return nil, merror.Transform(err).Describe("building box")
	}

	if req.KeyShareData != nil {
		err := keyshares.Create(
			ctx, app.DB,
			req.KeyShareData.OtherShareHash,
			req.KeyShareData.Share,
			req.KeyShareData.EncryptedInvitationKeyShare.String,
			box.ID,
			acc.IdentityID,
		)
		if err != nil {
			return nil, merror.Transform(err).Describe("creating key share")
		}
	}

	return box, nil
}
