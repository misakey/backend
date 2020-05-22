package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// GetInfo using a login challenge, fill a login info structure using login context
func (afs AuthFlowService) LoginGetInfo(ctx context.Context, loginChallenge string) (login.FlowInfo, error) {
	// get info about current login flow
	logCtx, err := afs.authFlow.GetLoginContext(ctx, loginChallenge)
	if err != nil {
		return login.FlowInfo{}, err
	}
	return login.FlowInfo{
		ClientID:       logCtx.Client.ID,
		ACRValues:      logCtx.OIDCContext.ACRValues,
		LoginHint:      logCtx.OIDCContext.LoginHint,
		RequestedScope: logCtx.RequestedScope,
	}, nil
}
