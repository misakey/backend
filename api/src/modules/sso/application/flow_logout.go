package application

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// Try to logout the user by invalidating the authentication session
func (sso *SSOService) Logout(ctx context.Context, _ request.Request) (interface{}, error) {
	// verify accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}
	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, err)

	err = sso.authFlowService.Logout(ctx, acc.Subject, acc.Token)
	if err != nil {
		return nil, merror.Transform(err).Describe("logging out on auth")
	}

	// expire all current authentication steps for the logged out subject
	err = sso.AuthenticationService.ExpireAll(ctx, tr, acc.IdentityID)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

// CleanCookie for authentication
func (sso *SSOService) CleanCookie(eCtx echo.Context, _ interface{}) error {
	// access token
	eCtx.SetCookie(&http.Cookie{
		Name:     "accesstoken",
		Value:    "",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})

	// token type
	eCtx.SetCookie(&http.Cookie{
		Name:     "tokentype",
		Value:    "",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})

	return nil
}
