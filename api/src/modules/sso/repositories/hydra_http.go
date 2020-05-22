package repositories

import (
	"context"
	"net/url"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
	"gitlab.misakey.dev/misakey/msk-sdk-go/rester"
)

// HTTP implements Hydra repository interface using HTTP REST
type HydraHTTP struct {
	publicRester         rester.Client
	authPublicFormRester rester.Client
	adminRester          rester.Client
	formAdminHydraRester rester.Client
}

// NewHydraHTTP is HTTP hydra structure constructor
func NewHydraHTTP(
	publicRester rester.Client,
	authPublicFormRester rester.Client,
	adminRester rester.Client,
	formAdminHydraRester rester.Client,
) *HydraHTTP {
	return &HydraHTTP{
		publicRester:         publicRester,
		authPublicFormRester: authPublicFormRester,
		adminRester:          adminRester,
		formAdminHydraRester: formAdminHydraRester,
	}
}

// GetLoginContext from hydra using challenge
func (h *HydraHTTP) GetLoginContext(ctx context.Context, challenge string) (login.Context, error) {
	logCtx := login.Context{}
	params := url.Values{}
	params.Add("login_challenge", challenge)
	err := h.adminRester.Get(ctx, "/oauth2/auth/requests/login", params, &logCtx)
	if err != nil {
		if merror.HasCode(err, merror.NotFoundCode) {
			err = merror.Transform(err).Detail("challenge", merror.DVNotFound)
		}
		return logCtx, err
	}
	return logCtx, nil
}

// Login user to Hydra
func (h *HydraHTTP) Login(ctx context.Context, challenge string, acceptance login.Acceptance) (login.Redirect, error) {
	redirect := login.Redirect{}
	params := url.Values{}
	params.Add("login_challenge", challenge)
	err := h.adminRester.Put(ctx, "/oauth2/auth/requests/login/accept", params, acceptance, &redirect)
	if err != nil {
		if merror.HasCode(err, merror.NotFoundCode) {
			err = merror.Transform(err).Detail("challenge", merror.DVNotFound)
		}
		return redirect, err
	}
	return redirect, nil
}

// // Logout : invalidates a subject's authentication session
// func (h *HydraHTTP) Logout(ctx context.Context, id string) error {
// 	route := fmt.Sprintf("/oauth2/auth/sessions/login?subject=%s", id)
//
// 	return h.adminRester.Delete(ctx, route, nil)
// }

// // RevokeToken : invalidate access & refresh tokens
// func (h *HydraHTTP) RevokeToken(ctx context.Context, revocation model.TokenRevocation) error {
// 	params := url.Values{}
// 	params.Add("token", revocation.Token)
// 	params.Add("client_id", revocation.ClientID)
// 	params.Add("client_secret", revocation.ClientSecret)
// 	return h.authPublicFormRester.Post(ctx, "/oauth2/revoke", nil, params, nil)
// }
//
// func (h *HydraHTTP) Introspect(ctx context.Context, opaqueTok string) (*model.IntrospectedToken, error) {
// 	introTok := model.IntrospectedToken{}
// 	route := fmt.Sprintf("/oauth2/introspect")
//
// 	params := url.Values{}
// 	params.Add("token", opaqueTok)
//
// 	err := h.formAdminHydraRester.Post(ctx, route, nil, params, &introTok)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &introTok, nil
// }
//
// // CreateClient: create a new Hydra Client in hydra
// func (h *HydraHTTP) CreateClient(ctx context.Context, hydraClient *model.HydraClient) error {
// 	route := fmt.Sprintf("/clients")
//
// 	return h.adminRester.Post(ctx, route, nil, hydraClient, nil)
// }
//
// // GetClient: retrieve a Hydra Client from hydra using a client id.
// func (h *HydraHTTP) GetClient(ctx context.Context, id string) (*model.HydraClient, error) {
// 	cli := model.HydraClient{}
// 	route := fmt.Sprintf("/clients/%s", id)
//
// 	err := h.adminRester.Get(ctx, route, nil, &cli)
// 	if err != nil {
// 		return nil, merror.Transform(err).If(merror.NotFound()).Detail("id", merror.DVNotFound).End()
// 	}
// 	return &cli, nil
// }
//
// // UpdateClient: update Hydra Client in hydra
// func (h *HydraHTTP) UpdateClient(ctx context.Context, hydraClient *model.HydraClient) error {
// 	route := fmt.Sprintf("/clients/%s", hydraClient.ID)
//
// 	return h.adminRester.Put(ctx, route, nil, hydraClient, hydraClient)
// }
