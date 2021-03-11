package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/datatag"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/org"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

// CreateBoxRequest ...
type CreateBoxRequest struct {
	PublicKey    string  `json:"public_key"`
	Title        string  `json:"title"`
	OwnerOrgID   *string `json:"owner_org_id"`
	KeyShareData *struct {
		OtherShareHash              string      `json:"other_share_hash"`
		Share                       string      `json:"misakey_share"`
		EncryptedInvitationKeyShare null.String `json:"encrypted_invitation_key_share"`
	} `json:"key_share"`
	InvitationData null.JSON `json:"invitation_data"`
}

// BindAndValidate ...
func (req *CreateBoxRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	err := v.ValidateStruct(req,
		v.Field(&req.PublicKey, v.Required),
		v.Field(&req.Title, v.Required, v.Length(1, 50)),
		v.Field(&req.OwnerOrgID, is.UUIDv4),
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

// CreateBox ...
func (app *BoxApplication) CreateBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateBoxRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}

	if authz.IsAMachine(*acc) {
		return nil, merr.Forbidden().Desc("please use /organizations/:id/boxes creation endpoint")
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// set defaults
	if req.OwnerOrgID == nil {
		req.OwnerOrgID = &app.selfOrgID
	}

	if *req.OwnerOrgID != app.selfOrgID {
		if err := org.MustBeAdmin(ctx, app.SSODB, *req.OwnerOrgID, acc.IdentityID); err != nil {
			return nil, merr.Forbidden()
		}
	}

	event, err := events.CreateCreateEvent(
		ctx,
		app.DB, app.RedConn, identityMapper,
		req.Title, req.PublicKey, *req.OwnerOrgID,
		nil, nil,
		acc.IdentityID,
	)
	if err != nil {
		return nil, merr.From(err).Desc("creating create event")
	}

	box, err := events.GetBoxView(ctx, app.DB, identityMapper, app.RedConn, event.BoxID, events.SetLastEvent(&event))
	if err != nil {
		return nil, merr.From(err).Desc("building box")
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
			return nil, merr.From(err).Desc("creating key share")
		}
	}
	return box, nil
}

type CreateOrgBoxRequest struct {
	PublicKey    string `json:"public_key"`
	Title        string `json:"title"`
	OwnerOrgID   string `json:"owner_org_id"`
	DataSubject  string `json:"data_subject"`
	DatatagID    string `json:"datatag_id"`
	KeyShareData *struct {
		OtherShareHash              string      `json:"other_share_hash"`
		Share                       string      `json:"misakey_share"`
		EncryptedInvitationKeyShare null.String `json:"encrypted_invitation_key_share"`
	} `json:"key_share"`
	InvitationData null.JSON `json:"invitation_data"`
}

// BindAndValidate ...
func (req *CreateOrgBoxRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	err := v.ValidateStruct(req,
		v.Field(&req.PublicKey, v.Required),
		v.Field(&req.Title, v.Required, v.Length(1, 50)),
		v.Field(&req.OwnerOrgID, v.Required, is.UUIDv4),
		v.Field(&req.DatatagID, v.Required, is.UUIDv4),
		v.Field(&req.DataSubject, v.Required, is.EmailFormat),
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

// CreateBoxForOrg ...
func (app *BoxApplication) CreateOrgBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateOrgBoxRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	if authz.IsNotAMachine(*acc) {
		return nil, merr.Forbidden().Desc("machine")
	}
	if req.OwnerOrgID != acc.IdentityID {
		return nil, merr.Forbidden()
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// check that the datatag belongs to the organization
	datatag, err := datatag.Get(ctx, app.SSODB, req.DatatagID)
	if err != nil {
		return nil, merr.From(err).Desc("getting datatag")
	}
	if req.OwnerOrgID != datatag.OrganizationID {
		return nil, merr.Forbidden()
	}

	subject, err := identity.Require(ctx, app.SSODB, app.RedConn, req.DataSubject)
	if err != nil {
		return nil, merr.From(err).Desc("requiring identity")
	}
	// if data subject has an account, invitation data is required for auto-invitation
	if subject.AccountID.Valid && req.InvitationData.IsZero() {
		return nil, merr.BadRequest().Add("invitation_data", merr.DVRequired)
	}

	event, err := events.CreateCreateEvent(
		ctx,
		app.DB, app.RedConn, identityMapper,
		req.Title, req.PublicKey, req.OwnerOrgID,
		&req.DatatagID, &subject.ID,
		acc.IdentityID,
	)
	if err != nil {
		return nil, merr.From(err).Desc("creating create event")
	}

	box, err := events.GetSimpleBox(ctx, app.DB, identityMapper, event.BoxID)
	if err != nil {
		return nil, merr.From(err).Desc("building box")
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
			return nil, merr.From(err).Desc("creating key share")
		}
	}

	// auto invite subject if linked to an account
	if err := events.InviteIdentityIfPossible(ctx, app.cryptoRepo, identityMapper, box, subject, acc.IdentityID, req.InvitationData); err != nil {
		return nil, merr.From(err).Desc("creating crypto actions")
	}
	return box, nil
}
