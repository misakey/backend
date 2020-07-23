package entrypoints

import (
	"net/http"
	"path/filepath"

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

// Handles GET /identities/:id - retrieve an existing identity
func (entrypoint IdentityHTTP) GetIdentity(ctx echo.Context) error {
	query := application.IdentityQuery{
		IdentityID: ctx.Param("id"),
	}

	if err := query.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	identity, err := entrypoint.service.GetIdentity(ctx.Request().Context(), query)
	if err != nil {
		return merror.Transform(err).Describe("get identity").From(merror.OriBody)
	}
	return ctx.JSON(http.StatusOK, identity)
}

// Handles PUT /identities/authable - retrieve authable identity information
func (entrypoint IdentityHTTP) RequireAuthableIdentity(ctx echo.Context) error {
	cmd := application.IdentityAuthableCmd{}
	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	identity, err := entrypoint.service.RequireAuthableIdentity(ctx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).Describe("could not require authable identity").From(merror.OriBody)
	}
	return ctx.JSON(http.StatusOK, identity)
}

// Handles PATCH /identities/:id - partially update an identity
func (entrypoint IdentityHTTP) PartiallyUpdateIdentity(ctx echo.Context) error {
	cmd := application.PartialUpdateIdentityCmd{}
	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.IdentityID = ctx.Param("id")
	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	err := entrypoint.service.PartialUpdateIdentity(ctx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).Describe("patch identity").From(merror.OriBody)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Handles PUT /identities/:id/avatar - upload a new avatar
func (entrypoint IdentityHTTP) UploadAvatar(ctx echo.Context) error {
	cmd := application.UploadAvatarCmd{}

	cmd.IdentityID = ctx.Param("id")

	file, err := ctx.FormFile("avatar")
	if err != nil {
		return merror.BadRequest().From(merror.OriBody).Detail("avatar", merror.DVRequired).Describe(err.Error())
	}
	if file.Size >= 3*1024*1024 {
		return merror.BadRequest().From(merror.OriBody).Detail("size", merror.DVInvalid).Describe("size must be < 3 mo")
	}

	data, err := file.Open()
	if err != nil {
		return err
	}

	cmd.Data = data
	cmd.Extension = filepath.Ext(file.Filename)

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	if err := entrypoint.service.UploadAvatar(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("upload avatar")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Handles DELETE /identities/:id/avatar - delete an existing avatar
func (entrypoint IdentityHTTP) DeleteAvatar(ctx echo.Context) error {
	cmd := application.DeleteAvatarCmd{}

	cmd.IdentityID = ctx.Param("id")

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}

	if err := entrypoint.service.DeleteAvatar(ctx.Request().Context(), cmd); err != nil {
		return merror.Transform(err).Describe("delete avatar")
	}

	return ctx.NoContent(http.StatusNoContent)
}
