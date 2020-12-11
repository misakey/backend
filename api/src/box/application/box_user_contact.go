package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
)

// BoxUserContactRequest ...
type BoxUserContactRequest struct {
	identityID          string
	ContactedIdentityID string `json:"identity_id"`
	Box                 struct {
		Title     string `json:"title"`
		PublicKey string `json:"public_key"`
	} `json:"box"`
	KeyShareData struct {
		OtherShareHash              string `json:"other_share_hash"`
		Share                       string `json:"misakey_share"`
		EncryptedInvitationKeyShare string `json:"encrypted_invitation_key_share"`
	} `json:"key_share"`
	InvitationData types.JSON `json:"invitation_data"`
}

// BindAndValidate ...
func (req *BoxUserContactRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.identityID = eCtx.Param("id")
	if err := v.ValidateStruct(req,
		v.Field(&req.identityID, v.Required, is.UUIDv4),
		v.Field(&req.ContactedIdentityID, v.Required, is.UUIDv4),
		v.Field(&req.InvitationData, v.Required),
	); err != nil {
		return err
	}

	if err := v.ValidateStruct(&req.Box,
		v.Field(&req.Box.Title, v.Required, v.Length(1, 50)),
		v.Field(&req.Box.PublicKey, v.Required),
	); err != nil {
		return err
	}

	if err := v.ValidateStruct(&req.KeyShareData,
		v.Field(&req.KeyShareData.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&req.KeyShareData.Share, v.Required, is.Base64),
		v.Field(&req.KeyShareData.EncryptedInvitationKeyShare, v.Required, is.Base64),
	); err != nil {
		return err
	}

	return nil
}

// BoxUserContact creates a new box and invite the contacted user
func (app *BoxApplication) BoxUserContact(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*BoxUserContactRequest)
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if acc.IdentityID != req.identityID {
		return nil, merror.Forbidden()
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	contact := events.ContactBox{
		ContactedIdentityID: req.ContactedIdentityID,
		IdentityID:          req.identityID,

		Title:     req.Box.Title,
		PublicKey: req.Box.PublicKey,

		OtherShareHash:              req.KeyShareData.OtherShareHash,
		Share:                       req.KeyShareData.Share,
		EncryptedInvitationKeyShare: req.KeyShareData.EncryptedInvitationKeyShare,

		InvitationDataJSON: req.InvitationData,
	}

	tr, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("initiating transaction")
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	box, err := events.CreateContactBox(ctx, tr, app.RedConn, identityMapper, app.filesRepo, app.cryptoRepo, contact)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating contact box")
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merror.Transform(cErr).Describe("committing transaction")
	}

	return box, err
}
