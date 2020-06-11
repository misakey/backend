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
	query := application.AccountQuery{
		AccountID: ctx.Param("id"),
	}

	if err := query.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	view, err := entrypoint.service.GetBackup(ctx.Request().Context(), query)
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

	cmd.SetAccountID(ctx.Param("id"))
	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := entrypoint.service.UpdateBackup(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("service update backup").From(merror.OriBody)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// Handles GET /accounts/:id/pwd-params - get the account password public parameters
func (entrypoint AccountHTTP) GetPwdParams(ctx echo.Context) error {
	query := application.AccountQuery{
		AccountID: ctx.Param("id"),
	}

	if err := query.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	view, err := entrypoint.service.GetAccountPwdParams(ctx.Request().Context(), query)
	if err != nil {
		return merror.Transform(err).Describe("service get account pwd params").From(merror.OriBody)
	}
	return ctx.JSON(http.StatusOK, view)
}

// Handles PUT /accounts/:id/password - change the password of the account
func (entrypoint AccountHTTP) ChangePassword(ctx echo.Context) error {
	cmd := application.ChangePasswordCmd{}

	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.AccountID = ctx.Param("id")
	if err := cmd.Validate(); err != nil {
		return merror.Transform(err)
	}

	if err := entrypoint.service.ChangePassword(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("change password")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Handles PUT /accounts/:id/password/reset - reset the password of the account
func (entrypoint AccountHTTP) ResetPassword(ctx echo.Context) error {
	cmd := application.ResetPasswordCmd{}

	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.AccountID = ctx.Param("id")
	if err := cmd.Validate(); err != nil {
		return merror.Transform(err)
	}

	if err := entrypoint.service.ResetPassword(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("reset password").From(merror.OriBody)
	}

	return ctx.NoContent(http.StatusNoContent)
}
