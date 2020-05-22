package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type IdentityHTTP struct {
	service application.SSOService
}

func NewIdentityHTTP(service application.SSOService) IdentityHTTP {
	return IdentityHTTP{service: service}
}

func (entrypoint *IdentityHTTP) AssertIdentity(ctx echo.Context) error {
	cmd := application.IdentityProofAssertion{}
	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.IdentityID = ctx.Param("id")
	if err := entrypoint.service.ConfirmIdentity(ctx.Request().Context(), cmd.IdentityProof); err != nil {
		return merror.Transform(err).Describe("could not validate identity proof")
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (entrypoint *IdentityHTTP) CreateIdentity(ctx echo.Context) error {
	cmd := application.CreateIdentityCommand{}
	if err := ctx.Bind(&cmd); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	result, err := entrypoint.service.CreateIdentity(ctx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).Describef("could not create identity").From(merror.OriBody)
	}

	return ctx.JSON(http.StatusCreated, result)
}
