package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type AccountHTTP struct {
	service application.SSOService
}

func NewAccountHTTP(service application.SSOService) AccountHTTP {
	return AccountHTTP{service: service}
}

// Handles GET /accounts/:id/backup - get the account backup information
func (entrypoint AccountHTTP) GetBackup(ctx echo.Context) error {
	cmd := application.GetBackupCmd{
		AccountID: ctx.Param("id"),
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	view, err := entrypoint.service.GetBackup(ctx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).Describe("service get backup").From(merror.OriBody)
	}
	return ctx.JSON(http.StatusOK, view)
}

// Handles PUT /accounts/:id/backup - update the account backup information
func (entrypoint AccountHTTP) UpdateBackup(ctx echo.Context) error {
	cmd := application.UpdateBackupCmd{}

	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.AccountID = ctx.Param("id")
	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := entrypoint.service.UpdateBackup(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("service update backup").From(merror.OriBody)
	}
	return ctx.NoContent(http.StatusNoContent)
}
