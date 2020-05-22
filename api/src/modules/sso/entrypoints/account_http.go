package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
)

type AccountHTTP struct {
	service application.SSOService
}

func NewAccountHTTP(service application.SSOService) *AccountHTTP {
	return &AccountHTTP{service: service}
}

func (a *AccountHTTP) Create(ctx echo.Context) error {
	cmd := application.CreateAccountCommand{}
	if err := ctx.Bind(&cmd); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	result, err := a.service.CreateAccount(ctx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).Describef("could not create account").From(merror.OriBody)
	}

	return ctx.JSON(http.StatusCreated, result)
}
