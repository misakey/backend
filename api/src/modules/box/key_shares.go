package box

import (
	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/keyshare"
)

type createKeyShareCmd struct {
	keyshare.BoxKeyShare
}

func (cmd createKeyShareCmd) validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&cmd.Share, v.Required, is.Base64),
		v.Field(&cmd.BoxID, v.Required, is.UUIDv4),
	)
}

func (h *handler) createKeyShare(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// user must be connected
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	cmd := createKeyShareCmd{}
	if err := eCtx.Bind(&cmd); err != nil {
		return err
	}

	if err := cmd.validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := keyshare.Create(
		ctx, h.repo.DB(),
		cmd.OtherShareHash, cmd.Share, cmd.BoxID, acc.Subject); err != nil {
		return merror.Transform(err).Describe("creating key share")
	}

	return eCtx.JSON(http.StatusCreated, cmd)
}

type keyShareQuery struct {
	otherShareHash string
}

func (query keyShareQuery) validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.otherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	)
}

func (h *handler) getKeyShare(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// user must be connected
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	query := keyShareQuery{
		otherShareHash: eCtx.Param("other-share-hash"),
	}
	if err := query.validate(); err != nil {
		return err
	}

	ks, err := keyshare.Get(ctx, h.repo.DB(), query.otherShareHash)
	if err != nil {
		return merror.Transform(err).Describe("getting key share")
	}

	if err := events.StoreJoin(ctx, h.repo.DB(), ks.BoxID, acc.Subject); err != nil {
		return err
	}

	return eCtx.JSON(http.StatusOK, ks)
}
