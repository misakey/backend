package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
)

type AuthnHTTP struct {
	service application.SSOService
}

func NewAuthnHTTP(service application.SSOService) AuthnHTTP {
	return AuthnHTTP{service: service}
}

// Handles POST /authn-steps - init a new authentication step
func (entrypoint AuthnHTTP) InitAuthnStep(ctx echo.Context) error {
	cmd := application.AuthenticationStepCmd{}
	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := entrypoint.service.InitAuthnStep(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("initing authn step").From(merror.OriBody)
	}

	return ctx.NoContent(http.StatusNoContent)
}
