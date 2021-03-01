package application

import (
	"context"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// Logout the user by invalidating the authentication session
func (sso *SSOService) Logout(ctx context.Context, _ request.Request) (interface{}, error) {
	// verify accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	err = sso.authFlowService.Logout(ctx, acc.Subject, acc.Token)
	if err != nil {
		return nil, merr.From(err).Desc("logging out on auth")
	}

	// expire all current authentication steps for the logged out subject
	err = sso.AuthenticationService.ExpireAll(ctx, tr, acc.IdentityID)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

// CleanOIDCCookie
func (sso *SSOService) CleanOIDCCookie(eCtx echo.Context, _ interface{}) error {
	authz.DelCookies(eCtx, "accesstoken", "tokentype")
	return nil
}

// CleanAuthnCookie
func (sso *SSOService) CleanAuthnCookie(eCtx echo.Context, _ interface{}) error {
	authz.DelCookies(eCtx, "authnaccesstoken", "authntokentype")
	return nil
}
