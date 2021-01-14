package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// AuthenticationStepCmd orders:
// - the retry of an authentication step init for the identity
type AuthenticationStepCmd struct {
	LoginChallenge string     `json:"login_challenge"`
	Step           authn.Step `json:"authn_step"`
}

// BindAndValidate ...
func (cmd *AuthenticationStepCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	if err := v.ValidateStruct(&cmd.Step,
		v.Field(&cmd.Step.IdentityID, v.Required, is.UUIDv4.Error("identity_id must be an UUIDv4")),
		v.Field(&cmd.Step.MethodName, v.Required),
	); err != nil {
		return err
	}

	return v.ValidateStruct(cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
	)
}

// InitAuthnStep is used to try to init an authentication step
func (sso *SSOService) InitAuthnStep(ctx context.Context, genReq request.Request) (interface{}, error) {
	cmd := genReq.(*AuthenticationStepCmd)

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// 0. check if the identity exists
	curIdentity, err := identity.Get(ctx, tr, cmd.Step.IdentityID)
	if err != nil {
		return nil, err
	}

	// 1. check login challenge
	_, err = sso.authFlowService.GetLoginContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return nil, merr.NotFound().Desc("finding login challenge").Add("login_challenge", merr.DVNotFound)
	}

	// 2. we try to init the authentication step
	err = sso.AuthenticationService.InitStep(ctx, tr, curIdentity, cmd.Step.MethodName)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}
