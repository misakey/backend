package application

import (
	"context"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// GetUserInfoCmd ...
type GetUserInfoCmd struct {
}

// BindAndValidate ...
func (cmd *GetUserInfoCmd) BindAndValidate(eCtx echo.Context) error {
	return nil
}

// GetUserInfo from hydra
// Basically returns the ID Token information
func (sso *SSOService) GetUserInfo(ctx context.Context, gen request.Request) (interface{}, error) {
	// retrieve access token
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.NotFound().Desc("access token not found")
	}

	// get info about current user
	userInfo, err := sso.authFlowService.GetUserInfo(ctx, acc.Token)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
