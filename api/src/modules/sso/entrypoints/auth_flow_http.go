package entrypoints

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
)

// AuthFlowHTTP provides function to bind to routes interacting with login flow
type AuthFlowHTTP struct {
	service authflow.Handler
}

// NewAuthFlowHTTP is AuthFlowHTTP constructor
func NewAuthFlowHTTP(service authflow.Handler) *AuthFlowHTTP {
	return &AuthFlowHTTP{
		service: service,
	}
}

// Handles init login flow request
func (c AuthFlowHTTP) InitLogin(ctx echo.Context) error {
	// parse parameters
	challenge := ctx.QueryParam("login_challenge")
	if challenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("login_challenge", merror.DVRequired)
	}
	// init login then redirect
	redirectURL := c.service.InitLogin(ctx.Request().Context(), challenge)
	return ctx.Redirect(http.StatusFound, redirectURL)
}

// // Handles get login info request
// func (c AuthFlowHTTP) GetLoginInfo(ctx echo.Context) error {
// 	// parse parameters
// 	challenge := ctx.QueryParam("login_challenge")
// 	if challenge == "" {
// 		return merror.BadRequest().From(merror.OriQuery).Detail("login_challenge", merror.DVRequired)
// 	}
// 	logInfo, err := a.service.GetLoginInfo(ctx.Request().Context(), challenge)
// 	if err != nil {
// 		return merror.Transform(err).Describe("could not get login info")
// 	}
// 	return ctx.JSON(http.StatusOK, logInfo)
// }

// // Handles authentication preparation request
// //
// // XXX This method does not use Secret.Value and the client does not provide one
// // but the code still requires an "Authentication" object
// // so we provide a "fake" one with an empty Secret.Value
// // TODO make the service layer require a structure that does not have Secret.Value,
// // but then it is not clear how the "check" method can be called by both "run" and "init" methods
// // (see src/service/authn/confirmation_code.go)
// func (c *AuthFlowHTTP) PrepareAuthNMethod(ctx echo.Context) error {
// 	// check body
// 	req := model.AuthenticationRequest{}
// 	err := ctx.Bind(&req)
// 	if err != nil {
// 		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
// 	}
//
// 	if err := ctx.Validate(&req); err != nil {
// 		return merror.Transform(err).From(merror.OriBody)
// 	}
//
// 	authentication := req.Authentication
//
// 	switch req.Secret.Kind {
// 	case model.SecretConfirmationCode:
// 		authentication.Secret = secret.ConfirmationCode("")
//
// 	case model.SecretPassword:
// 		authentication.Secret = unhashed.Password("")
//
// 	case model.SecretPasswordHash:
// 		authentication.Secret = prehashed.Password{}
// 	}
//
// 	if err := a.service.PrepareAuthNMethod(ctx.Request().Context(), &authentication); err != nil {
// 		return merror.Transform(err).Describe("could not prepare authn")
// 	}
// 	return ctx.NoContent(http.StatusNoContent)
// }
//
// // Handles user authentication request
// func (c *AuthFlowHTTP) Authenticate(ctx echo.Context) error {
// 	// check body parameters
// 	req := model.AuthenticationRequest{}
// 	err := ctx.Bind(&req)
// 	if err != nil {
// 		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
// 	}
//
// 	if err := ctx.Validate(&req); err != nil {
// 		return merror.Transform(err).From(merror.OriBody)
// 	}
//
// 	authentication := req.Authentication
//
// 	switch req.Secret.Kind {
// 	case model.SecretConfirmationCode:
// 		s, ok := req.Secret.Value.(string)
// 		if !ok {
// 			return merror.BadRequest().From(merror.OriBody).Describe("could not cast confirmation code to string")
// 		}
// 		authentication.Secret = secret.ConfirmationCode(s)
//
// 	case model.SecretPassword:
// 		s, ok := req.Secret.Value.(string)
// 		if !ok {
// 			return merror.BadRequest().From(merror.OriBody).Describe("could not cast password to string")
// 		}
// 		authentication.Secret = unhashed.Password(s)
//
// 	case model.SecretPasswordHash:
// 		pwdHash := prehashed.Password{}
// 		err := mapstructure.Decode(req.Secret.Value, &pwdHash)
// 		if err != nil {
// 			return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
// 		}
// 		authentication.Secret = pwdHash
// 	}
//
// 	// challenge login then redirect
// 	acceptInfo, err := a.service.Authenticate(ctx.Request().Context(), &authentication)
// 	if err != nil {
// 		return merror.Transform(err).Describef("could not authenticate")
// 	}
// 	// do not redirect - frontend browser does not get this redirection
// 	return ctx.JSON(http.StatusOK, acceptInfo)
// }
//
// // Logout : handler for user logout
// func (c *AuthFlowHTTP) Logout(ctx echo.Context) error {
// 	logoutReq := model.LogoutRequest{}
// 	err := ctx.Bind(&logoutReq)
// 	if err != nil {
// 		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
// 	}
//
// 	err = ctx.Validate(&logoutReq)
// 	if err != nil {
// 		return merror.Transform(err).From(merror.OriBody)
// 	}
//
// 	_, err = uuid.Parse(logoutReq.UserID)
// 	if err != nil {
// 		return merror.BadRequest().From(merror.OriBody).Detail("user_id", merror.DVMalformed).Describe(err.Error())
// 	}
//
// 	err = a.service.Logout(ctx.Request().Context(), logoutReq.UserID)
// 	if err != nil {
// 		return merror.Transform(err).Describef("could not logout the user")
// 	}
// 	return ctx.NoContent(http.StatusNoContent)
// }
