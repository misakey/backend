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
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
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
		return merr.BadRequest().Ori(merr.OriQuery)
	}
	if cmd.Challenge == "" {
		return merr.BadRequest().Ori(merr.OriQuery).Add("login_challenge", merr.DVRequired)
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
		return sso.authFlowService.LoginRedirectErr(merr.From(err).Desc("initing authn process")), nil
	}

	// return the login page url
	return sso.authFlowService.BuildLoginURL(req.Challenge), nil
}

// RequireIdentityCmd orders:
// - the assurance of an identifier matching the received value
// - a new account/identity if nothing linked to the identifier value is found
// - a init of confirmation code authencation method for the identity
type RequireIdentityCmd struct {
	LoginChallenge  string `json:"login_challenge"`
	IdentifierValue string `json:"identifier_value"`
	PasswordReset   bool   `json:"password_reset"`
}

// BindAndValidate the RequireIdentityCmd
func (cmd *RequireIdentityCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	// lowcase the email
	cmd.IdentifierValue = strings.ToLower(cmd.IdentifierValue)

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
		v.Field(&cmd.IdentifierValue, v.Required, is.EmailFormat),
	); err != nil {
		return err
	}
	return nil
}

// RequireIdentityAView ...
type RequireIdentityView struct {
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

// RequireIdentity for an auth flow.
// This method is used to retrieve information about the identity attached to an identifier value.
// The identifier value is set by the end-user on the interface and we receive it here.
// The function returns information about the Account & Identity that corresponds to the identifier.
// It creates if required the pair account/identity.
// If an identity is created during this process, an confirmation code auth method is started
// This method will exceptionnaly both proof the identity and confirm the login flow within the auth flow.
func (sso *SSOService) RequireIdentity(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*RequireIdentityCmd)

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

	// 1. check if an identity exist for the identifier
	// NOTE: to_change_on_more_identifier_kind
	curIdentity, err := identity.GetByIdentifierValue(ctx, tr, cmd.IdentifierValue)
	if err != nil && !merr.IsANotFound(err) {
		return nil, err
	}

	// 2. create an identity if nothing was found
	if merr.IsANotFound(err) {
		// a. create the Identity without account
		curIdentity = identity.Identity{
			DisplayName:     strings.Title(strings.Replace(strings.Split(cmd.IdentifierValue, "@")[0], ".", " ", -1)),
			IdentifierValue: cmd.IdentifierValue,
			IdentifierKind:  identity.EmailIdentifier,
			MFAMethod:       "disabled",
		}
		err = identity.Create(ctx, tr, sso.redConn, &curIdentity)
		if err != nil {
			return nil, err
		}
	}

	// 3. compute the expected ACR
	// if no ACR is expected, set it according to the identity state
	expectedACR := logCtx.OIDCContext.ACRValues().Get()
	if expectedACR == oidc.ACR0 {
		expectedACR = oidc.ACR1
		if curIdentity.AccountID.Valid {
			expectedACR = oidc.ACR2
		}
	}
	// in all cases, if any MFA method is setup, the expected ACR is enforce according to it
	if curIdentity.MFAMethod != "disabled" {
		expectedACR = oidc.GetMethodACR(curIdentity.MFAMethod)
	}

	// 4. get the appropriate authn step - this is the start of the login flow so the current ACR is 0
	currentACR := oidc.ACR0
	step, err := sso.AuthenticationService.PrepareNextStep(
		ctx, tr, sso.redConn,
		curIdentity, currentACR, expectedACR,
		cmd.PasswordReset,
	)
	if err != nil {
		return nil, merr.From(err).Descf("preparing step").Add("identity_id", curIdentity.ID).Add("expected_acr", expectedACR.String())
	}
	if step == nil {
		return nil, merr.Internal().Descf("step is nil").Add("identity_id", curIdentity.ID).Add("expected_acr", expectedACR.String())
	}

	// 5. update process with the right expected ACR
	// and with password reset argument
	if err := sso.AuthenticationService.UpdateProcess(ctx, sso.redConn, logCtx.Challenge, expectedACR, cmd.PasswordReset); err != nil {
		return nil, merr.From(err).Desc("updating process")
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merr.From(cErr).Desc("committing transaction")
	}

	// 5. bind identity information on view
	view := RequireIdentityView{}
	view.Identity.DisplayName = curIdentity.DisplayName
	view.Identity.AvatarURL = curIdentity.AvatarURL
	view.AuthnStep.IdentityID = curIdentity.ID
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
		return merr.BadRequest().Ori(merr.OriQuery)
	}
	if cmd.Challenge == "" {
		return merr.BadRequest().Ori(merr.OriQuery).Add("login_challenge", merr.DVRequired)
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
		return view, merr.From(err).Desc("could not get context")
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
	LoginChallenge string     `json:"login_challenge"`
	Step           authn.Step `json:"authn_step"`
}

// BindAndValidate ...
func (cmd *LoginAuthnStepCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
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

	// ensure the login challenge is correct
	logCtx, err := sso.authFlowService.GetLoginContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return view, err
	}
	curIdentity, err := identity.Get(ctx, tr, cmd.Step.IdentityID)
	if err != nil {
		return view, err
	}

	// try to assert the authentication step
	err = sso.AuthenticationService.AssertStep(ctx, tr, sso.redConn, logCtx.Challenge, &curIdentity, cmd.Step)
	if err != nil {
		return view, err
	}

	// upgrade the authentication process
	process, err := sso.AuthenticationService.UpgradeProcess(ctx, tr, sso.redConn, logCtx.Challenge, curIdentity, cmd.Step.MethodName)
	if err != nil {
		return view, merr.From(err).Desc("upgrading authn process")
	}
	if cErr := tr.Commit(); cErr != nil {
		return nil, merr.From(cErr).Desc("committing transaction")
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

	// set subject to the account id if there otherwise use the identity id
	if curIdentity.AccountID.Valid {
		logCtx.Subject = curIdentity.AccountID.String
	} else {
		logCtx.Subject = curIdentity.ID
	}
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
