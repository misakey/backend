package application

import (
	"context"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// LoginInitCmd ...
type LoginInitCmd struct {
	Challenge string `query:"login_challenge"`
}

// BindAndValidate ...
func (cmd *LoginInitCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriQuery)
	}
	if cmd.Challenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("login_challenge", merror.DVRequired)
	}
	return nil
}

// LoginInit a user authentication stage (a.k.a. login flow)
// It interacts with hydra and login sessions to know either user is already authenticated or not
// It returns a URL user's agent should be redirected to
func (sso *SSOService) LoginInit(ctx context.Context, gen request.Request) (interface{}, error) {
	req := gen.(*LoginInitCmd)

	// get info about current login session
	loginCtx, err := sso.authFlowService.GetLoginContext(ctx, req.Challenge)
	if err != nil {
		return sso.authFlowService.LoginRedirectErr(err), nil
	}

	sessionACR := oidc.ACR0
	expectedACR := loginCtx.OIDCContext.ACRValues().Get()

	// skip indicates if an active session has been detected
	// we check if login session ACR are high enough to accept authentication
	if loginCtx.Skip {
		session, err := sso.AuthenticationService.GetSession(ctx, loginCtx.SessionID)
		if err == nil {
			sessionACR = session.ACR
			// if the session ACR is higher or equivalent to the expected ACR, we accept the login
			if session.ACR >= expectedACR {
				// set browser cookie as authentication method
				loginCtx.OIDCContext.AddAMR(oidc.AMRBrowserCookie)
				loginCtx.OIDCContext.SetACRValue(session.ACR)
				loginCtx.OIDCContext.SetMID(session.IdentityID)
				loginCtx.OIDCContext.SetAID(session.AccountID)
				redirectTo, err := sso.authFlowService.BuildAndAcceptLogin(ctx, loginCtx)
				if err != nil {
					return sso.authFlowService.LoginRedirectErr(err), nil
				}
				return redirectTo, nil
			}
		}
		if authflow.HasNonePrompt(loginCtx.RequestURL) {
			return sso.authFlowService.LoginRequiredErr(), nil
		}
	}

	// store information about the incomming authentication process
	if err := sso.AuthenticationService.InitProcess(ctx, req.Challenge, sessionACR, expectedACR); err != nil {
		return sso.authFlowService.LoginRedirectErr(merror.Transform(err).Describe("initing authn process")), nil
	}

	// return the login page url
	return sso.authFlowService.BuildLoginURL(req.Challenge), nil
}

// IdentityAuthableCmd orders:
// - the assurance of an identifier matching the received value
// - a new account if not authable identity linked to such identifier is found
// - a new identity authable linking both previous entities
// - a init of confirmation code authencation method for the identity
type IdentityAuthableCmd struct {
	LoginChallenge string `json:"login_challenge"`
	Identifier     struct {
		Value string `json:"value"`
	} `json:"identifier"`
}

// BindAndValidate the IdentityAuthableCmd
func (cmd *IdentityAuthableCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	// validate nested structure separately
	if err := v.ValidateStruct(&cmd.Identifier,
		v.Field(&cmd.Identifier.Value, v.Required, is.EmailFormat),
	); err != nil {
		return err
	}

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
	); err != nil {
		return err
	}
	return nil
}

// IdentityAuthableView ...
type IdentityAuthableView struct {
	Identity struct {
		DisplayName string      `json:"display_name"`
		AvatarURL   null.String `json:"avatar_url"`
	} `json:"identity"`
	AuthnStep nextStepView `json:"authn_step"`
}

type nextStepView struct {
	IdentityID string         `json:"identity_id"`
	MethodName oidc.MethodRef `json:"method_name"`
	Metadata   *types.JSON    `json:"metadata"`
}

// RequireAuthableIdentity for an auth flow.
// This method is used to retrieve information about the authable identity attached to an identifier value.
// The identifier value is set by the end-user on the interface and we receive it here.
// The function returns information about the Account & Identity that corresponds to the identifier.
// It creates if required the trio identifier/account/identity.
// If an identity is created during this process, an confirmation code auth method is started
// This method will exceptionnaly both proof the identity and confirm the login flow within the auth flow.
func (sso *SSOService) RequireAuthableIdentity(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*IdentityAuthableCmd)

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// 0. check the login challenge exists
	logCtx, err := sso.authFlowService.GetLoginContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return nil, err
	}

	// 1. ensure create the Identifier does exist
	identifier := identity.Identifier{
		Kind:  identity.EmailIdentifier,
		Value: cmd.Identifier.Value,
	}
	err = identity.RequireIdentifier(ctx, tr, &identifier)
	if err != nil {
		return nil, err
	}

	// 2. check if an identity exist for the identifier
	identityNotFound := func(err error) bool { return err != nil && merror.HasCode(err, merror.NotFoundCode) }
	authable, err := identity.GetAuthableByIdentifierID(ctx, tr, identifier.ID)
	if err != nil && !identityNotFound(err) {
		return nil, err
	}

	// 3. create an identity if nothing was found
	if identityNotFound(err) {
		// a. create the Identity without account
		authable = identity.Identity{
			IdentifierID: identifier.ID,
			DisplayName:  strings.Title(strings.Replace(strings.Split(cmd.Identifier.Value, "@")[0], ".", " ", -1)),
			IsAuthable:   true,
			// fill the identifier manually for later use
			Identifier: identifier,
		}
		err = identity.Create(ctx, tr, sso.redConn, &authable)
		if err != nil {
			return nil, err
		}
	}
	// get the appropriate authn step
	// NOTE: not handled - authnsession ACR
	step, err := sso.AuthenticationService.NextStep(ctx, tr, authable, oidc.ACR0, logCtx.OIDCContext.ACRValues())
	if err != nil {
		return nil, merror.Transform(err).Describe("getting next authn step")
	}
	if cErr := tr.Commit(); cErr != nil {
		return nil, merror.Transform(cErr).Describe("committing transaction")
	}

	// bind identity information on view
	view := IdentityAuthableView{}
	view.Identity.DisplayName = authable.DisplayName
	view.Identity.AvatarURL = authable.AvatarURL
	view.AuthnStep.IdentityID = authable.ID
	view.AuthnStep.MethodName = step.MethodName
	if step.RawJSONMetadata != nil {
		view.AuthnStep.Metadata = &step.RawJSONMetadata
	}
	return view, nil
}

// LoginInfoQuery ...
type LoginInfoQuery struct {
	Challenge string `query:"login_challenge"`
}

// BindAndValidate ...
func (cmd *LoginInfoQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriQuery)
	}
	if cmd.Challenge == "" {
		return merror.BadRequest().From(merror.OriQuery).Detail("login_challenge", merror.DVRequired)
	}
	return nil
}

// LoginInfoView bears data about current user authentication status
type LoginInfoView struct {
	Client struct { // concerned relying party
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		LogoURL   null.String `json:"logo_uri"`
		TosURL    null.String `json:"tos_uri"`
		PolicyURL null.String `json:"policy_uri"`
	} `json:"client"`
	RequestedScope []string       `json:"scope"`
	ACRValues      oidc.ClassRefs `json:"acr_values"`
	LoginHint      string         `json:"login_hint"`
}

// LoginInfo ...
func (sso *SSOService) LoginInfo(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*LoginInfoQuery)
	view := LoginInfoView{}

	logCtx, err := sso.authFlowService.GetLoginContext(ctx, query.Challenge)
	if err != nil {
		return view, merror.Transform(err).Describe("could not get context")
	}

	// fill view with domain model
	view.Client.ID = logCtx.Client.ID
	view.Client.Name = logCtx.Client.Name
	view.Client.LogoURL = logCtx.Client.LogoURL
	view.Client.TosURL = logCtx.Client.TosURL
	view.Client.PolicyURL = logCtx.Client.PolicyURL
	view.RequestedScope = logCtx.RequestedScope
	view.ACRValues = logCtx.OIDCContext.ACRValues()
	view.LoginHint = logCtx.OIDCContext.LoginHint()
	return view, nil
}

// LoginAuthnStepCmd ...
type LoginAuthnStepCmd struct {
	LoginChallenge   string            `json:"login_challenge"`
	Step             authn.Step        `json:"authn_step"`
	PasswordResetExt *PasswordResetCmd `json:"password_reset"`
}

// BindAndValidate ...
func (cmd *LoginAuthnStepCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	// validate nested structure separately
	if err := v.ValidateStruct(&cmd.Step,
		v.Field(&cmd.Step.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Step.MethodName, v.Required),
		v.Field(&cmd.Step.RawJSONMetadata, v.Required),
	); err != nil {
		return err
	}

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
	); err != nil {
		return err
	}

	if cmd.PasswordResetExt != nil {
		return cmd.PasswordResetExt.Validate()
	}
	return nil
}

// LoginAuthnStepView ...
type LoginAuthnStepView struct {
	Next        string        `json:"next"`
	AccessToken string        `json:"access_token"`
	NextStep    *nextStepView `json:"authn_step,omitempty"`
	RedirectTo  *string       `json:"redirect_to,omitempty"`
}

// AssertAuthnStep in a multi-factor authentication process
// the authentication process is stored and considering the final expected ACR:
// - a new authn-step is returned to the client
// - the login flow is accepted and a redirect url is returned
func (sso *SSOService) AssertAuthnStep(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*LoginAuthnStepCmd)
	view := LoginAuthnStepView{}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// ensure the login challenge is correct and the identity is authable
	logCtx, err := sso.authFlowService.GetLoginContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return view, err
	}
	curIdentity, err := identity.Get(ctx, tr, cmd.Step.IdentityID)
	if err != nil {
		return view, err
	}
	if !curIdentity.IsAuthable {
		err = merror.Forbidden().Describe("identity not authable")
		return view, err
	}

	// try to assert the authentication step
	err = sso.AuthenticationService.AssertStep(ctx, tr, sso.redConn, logCtx.Challenge, &curIdentity, cmd.Step)
	if err != nil {
		return view, err
	}

	// emailed_code has potentially a reset password extension
	if cmd.Step.MethodName == oidc.AMREmailedCode && cmd.PasswordResetExt != nil {
		err = sso.resetPassword(ctx, tr, sso.redConn, *cmd.PasswordResetExt, cmd.Step.IdentityID)
		if err != nil {
			return view, err
		}
		cmd.Step.MethodName = oidc.AMRResetPassword
	}

	// upgrade the authentication process
	process, err := sso.AuthenticationService.UpgradeProcess(ctx, tr, logCtx.Challenge, curIdentity, cmd.Step.MethodName)
	if err != nil {
		return view, merror.Transform(err).Describe("upgrading authn process")
	}
	if cErr := tr.Commit(); cErr != nil {
		return nil, merror.Transform(cErr).Describe("committing transaction")
	}

	view.AccessToken = process.AccessToken

	// if an new authn step was returned - the login flow requires more authn steps
	if process.NextStep != nil {
		view.Next = "authn_step"
		view.NextStep = &nextStepView{
			IdentityID: process.NextStep.IdentityID,
			MethodName: process.NextStep.MethodName,
		}
		if process.NextStep.RawJSONMetadata != nil {
			view.NextStep.Metadata = &process.NextStep.RawJSONMetadata
		}
		return view, nil
	}

	// finally accept the login!

	// set subject to the identifier id
	logCtx.Subject = curIdentity.IdentifierID
	logCtx.OIDCContext.SetACRValue(process.CompleteAMRs.ToACR())
	logCtx.OIDCContext.SetAMRs(process.CompleteAMRs)
	logCtx.OIDCContext.SetMID(curIdentity.ID)
	logCtx.OIDCContext.SetAID(curIdentity.AccountID)

	view.Next = "redirect"
	var redirectTo string
	redirectTo, err = sso.authFlowService.BuildAndAcceptLogin(ctx, logCtx)
	view.RedirectTo = &redirectTo
	return view, err
}
