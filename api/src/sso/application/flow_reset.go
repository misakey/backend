package application

import (
	"context"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/login"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

type FlowResetCmd struct {
	Challenge string `query:"login_challenge"`
}

func (cmd *FlowResetCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriQuery)
	}
	return nil
}

// Reset the authorization code flow by redirecting to the initial request url
// if the request url is not found, redirect the main page of the app
func (sso *SSOService) ResetFlow(ctx context.Context, gen request.Request) (interface{}, error) {
	req := gen.(*FlowResetCmd)

	// get info about the flow to know the original request URL
	var oriAuthURL string
	var loginCtx login.Context
	var contextErr error
	if req.Challenge != "" {
		loginCtx, contextErr = sso.authFlowService.GetLoginContext(ctx, req.Challenge)
		oriAuthURL = loginCtx.RequestURL
	}

	// if the auth url has no login_hint, try to find it in current login context
	// NOTE: the system consider impossible its retrieval if not any loginCtx has been retrieved
	if authflow.HasNoLoginHint(oriAuthURL) && contextErr == nil {
		// Try 1: get it from the retrieved login context
		if lh := loginCtx.OIDCContext.LoginHint(); lh != "" {
			oriAuthURL, _ = format.AddQueryParam(oriAuthURL, "login_hint", lh)
			return sso.authFlowService.BuildResetURL(oriAuthURL), nil
		}

		// Try 2: use the authentication session
		session, err := sso.AuthenticationService.GetSession(ctx, loginCtx.SessionID)
		if err == nil && session.IdentityID != "" {
			// retrieve the identity/identifier couple using the
			identity, err := identity.Get(ctx, sso.sqlDB, session.IdentityID)
			if err == nil {
				oriAuthURL, _ = format.AddQueryParam(oriAuthURL, "login_hint", identity.Identifier.Value)
				return sso.authFlowService.BuildResetURL(oriAuthURL), nil
			}
		}

		// Try 3: use the authentication process
		process, err := sso.AuthenticationService.GetProcess(ctx, loginCtx.Challenge)
		if err == nil && process.IdentityID != "" {
			// retrieve the identity/identifier couple using the
			curIdentity, err := identity.Get(ctx, sso.sqlDB, process.IdentityID)
			if err == nil {
				oriAuthURL, _ = format.AddQueryParam(oriAuthURL, "login_hint", curIdentity.Identifier.Value)
				return sso.authFlowService.BuildResetURL(oriAuthURL), nil
			}
		}
	}
	return sso.authFlowService.BuildResetURL(oriAuthURL), nil
}
