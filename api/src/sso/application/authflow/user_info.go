package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/userinfo"
)

func (afs Service) GetUserInfo(ctx context.Context, token string) (*userinfo.UserInfo, error) {
	return afs.authFlow.GetUserInfo(ctx, token)
}
