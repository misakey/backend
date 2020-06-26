package box

import (
	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/keyshare"
)

type createKeyShareCmd struct {
	keyshare.KeyShare
}

func (cmd createKeyShareCmd) validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.InvitationHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&cmd.Share, v.Required, is.Base64),
	)
}

func (h *handler) createKeyShare(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// user must be connected
	if ajwt.GetAccesses(ctx) == nil {
		return merror.Forbidden()
	}

	cmd := createKeyShareCmd{}
	if err := eCtx.Bind(&cmd); err != nil {
		return err
	}

	if err := cmd.validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := keyshare.Create(ctx, h.repo.DB(), cmd.InvitationHash, cmd.Share); err != nil {
		return merror.Transform(err).Describe("creating key share")
	}

	return eCtx.JSON(http.StatusCreated, cmd)
}

type keyShareQuery struct {
	invitationHash string
}

func (query keyShareQuery) validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.invitationHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	)
}

func (h *handler) getKeyShare(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// user must be connected
	if ajwt.GetAccesses(ctx) == nil {
		return merror.Forbidden()
	}

	query := keyShareQuery{
		invitationHash: eCtx.Param("invitation-hash"),
	}
	if err := query.validate(); err != nil {
		return err
	}

	ks, err := keyshare.Get(ctx, h.repo.DB(), query.invitationHash)
	if err != nil {
		return merror.Transform(err).Describe("getting key share")
	}

	return eCtx.JSON(http.StatusOK, ks)
}
