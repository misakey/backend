package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// GetEncryptedKeyShareRequest ...
type GetEncryptedKeyShareRequest struct {
	BoxID string `query:"box_id"`
}

// BindAndValidate ...
func (req *GetEncryptedKeyShareRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriPath)
	}
	return v.ValidateStruct(req,
		v.Field(&req.BoxID, v.Required, is.UUIDv4),
	)
}

// GetEncryptedKeyShare ...
func (app *BoxApplication) GetEncryptedKeyShare(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetEncryptedKeyShareRequest)

	ks, err := keyshares.GetLastForBoxID(ctx, app.DB, req.BoxID)
	if err != nil {
		return nil, merr.From(err).Desc("getting key share")
	}

	return ks.EncryptedInvitationKeyShare, nil
}
