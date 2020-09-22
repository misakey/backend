package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type CryptoActionsHTTP struct {
	service application.SSOService
}

func NewCryptoActionsHTTP(service application.SSOService) CryptoActionsHTTP {
	return CryptoActionsHTTP{
		service: service,
	}
}

func (entrypoint CryptoActionsHTTP) ListCryptoActions(ctx echo.Context) error {
	query := application.ListCryptoActionsQuery{}

	query.AccountID = ctx.Param("id")
	if err := query.Validate(); err != nil {
		return merror.Transform(err)
	}

	actions, err := entrypoint.service.ListCryptoActions(ctx.Request().Context(), query)
	if err != nil {
		return merror.Transform(err).Describe("listing crypto actions")
	}
	return ctx.JSON(http.StatusOK, actions)
}

func (entrypoint CryptoActionsHTTP) DeleteCryptoActions(ctx echo.Context) error {
	query := application.DeleteCryptoActionsQuery{}

	if err := ctx.Bind(&query); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}
	if err := query.Validate(); err != nil {
		return err
	}

	query.AccountID = ctx.Param("id")

	if err := entrypoint.service.DeleteCryptoActionsUntil(ctx.Request().Context(), query); err != nil {
		return merror.Transform(err).Describe("processing deletion request")
	}

	return ctx.JSON(http.StatusNoContent, nil)
}
