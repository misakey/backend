package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application"
)

// AuthFlowHTTP provides function to bind to routes interacting with login flow
type AuthFlowHTTP struct {
	service *application.SSOService
}

// NewAuthFlowHTTP is AuthFlowHTTP constructor
func NewAuthFlowHTTP(service *application.SSOService) *AuthFlowHTTP {
	return &AuthFlowHTTP{
		service: service,
	}
}

// Handles GET /login - init login flow request
func (af AuthFlowHTTP) LoginInit(ctx echo.Context) error {
	// parse parameters
	loginChallenge := ctx.QueryParam("login_challenge")
	if loginChallenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("login_challenge", merror.DVRequired)
	}
	// init login then redirect
	redirectURL := af.service.LoginInit(ctx.Request().Context(), loginChallenge)
	return ctx.Redirect(http.StatusFound, redirectURL)
}

// Handles GET /login/info - get information about current login
func (af AuthFlowHTTP) LoginInfo(ctx echo.Context) error {
	// parse parameters
	loginChallenge := ctx.QueryParam("login_challenge")
	if loginChallenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("login_challenge", merror.DVRequired)
	}
	// init login then redirect
	info, err := af.service.LoginInfo(ctx.Request().Context(), loginChallenge)
	if err != nil {
		return merror.Transform(err).Describe("cannot retrieve login info")
	}
	return ctx.JSON(http.StatusOK, info)

}

// Handles POST /login/step - perform authentication request for a login flow
func (af AuthFlowHTTP) LoginAuthnStep(ctx echo.Context) error {
	cmd := application.LoginAuthnStepCmd{}

	if err := ctx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	redirect, err := af.service.LoginAuthnStep(ctx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).From(merror.OriBody).Describe("login flow step")
	}
	return ctx.JSON(http.StatusOK, redirect)
}

// Handles GET /consent - init consent flow request
func (af AuthFlowHTTP) ConsentInit(ctx echo.Context) error {
	consentChallenge := ctx.QueryParam("consent_challenge")
	if consentChallenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("consent_challenge", merror.DVRequired)
	}
	// init consent then redirect
	redirectURL := af.service.ConsentInit(ctx.Request().Context(), consentChallenge)
	return ctx.Redirect(http.StatusFound, redirectURL)
}

// Handles GET /consent/info - get information about current consent
func (af AuthFlowHTTP) ConsentInfo(ctx echo.Context) error {
	// parse parameters
	consentChallenge := ctx.QueryParam("consent_challenge")
	if consentChallenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("consent_challenge", merror.DVRequired)
	}
	// init consent then redirect
	info, err := af.service.ConsentInfo(ctx.Request().Context(), consentChallenge)
	if err != nil {
		return merror.Transform(err).Describe("cannot retrieve consent info")
	}
	return ctx.JSON(http.StatusOK, info)

}

func (af AuthFlowHTTP) Logout(ctx echo.Context) error {
	if err := af.service.Logout(ctx.Request().Context()); err != nil {
		return merror.Transform(err).From(merror.OriBody).Describe("logout")
	}
	return ctx.NoContent(http.StatusNoContent)
}

// Handles POST /consent - accept a consent request
func (af AuthFlowHTTP) ConsentAccept(eCtx echo.Context) error {
	cmd := application.ConsentAcceptCmd{}

	if err := eCtx.Bind(&cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	if err := cmd.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	redirect, err := af.service.ConsentAccept(eCtx.Request().Context(), cmd)
	if err != nil {
		return merror.Transform(err).From(merror.OriBody).Describe("consent flow")
	}
	return eCtx.JSON(http.StatusOK, redirect)
}

// Handles GET /backup - get backup during auth flow
func (af AuthFlowHTTP) GetBackup(eCtx echo.Context) error {
	query := application.GetBackupQuery{}
	// parse parameters
	query.LoginChallenge = eCtx.QueryParam("login_challenge")
	query.IdentityID = eCtx.QueryParam("identity_id")

	if err := query.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriQuery)
	}

	backup, err := af.service.GetBackupDuringAuth(eCtx.Request().Context(), query)
	if err != nil {
		return merror.Transform(err).Describe("get backup")
	}

	return eCtx.JSON(http.StatusOK, backup)
}
