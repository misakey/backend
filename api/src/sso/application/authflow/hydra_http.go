package authflow

import (
	"context"
	"fmt"
	"net/url"

	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/login"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/userinfo"
)

// HydraHTTP implements Hydra repository interface using HTTP REST
type HydraHTTP struct {
	publicJSONRester rester.Client // public port and no authorization required
	adminJSONRester  rester.Client // admin port but no authorization required
	adminFormRester  rester.Client // admin port but no authorization required

	protectedPublicFormRester rester.Client // public port but authorization protected endpoints
}

// NewHydraHTTP is HTTP hydra structure constructor
func NewHydraHTTP(
	publicJSONRester rester.Client,
	adminJSONRester rester.Client,
	adminFormRester rester.Client,

	protectedPublicFormRester rester.Client,
) *HydraHTTP {
	return &HydraHTTP{
		publicJSONRester: publicJSONRester,
		adminJSONRester:  adminJSONRester,
		adminFormRester:  adminFormRester,

		protectedPublicFormRester: protectedPublicFormRester,
	}
}

// GetLoginContext from hydra
func (h HydraHTTP) GetLoginContext(ctx context.Context, loginChallenge string) (login.Context, error) {
	// 1. prepare the request
	// expected hydra DTO format
	hydraLogReq := struct {
		Challenge      string   `json:"challenge"`
		Skip           bool     `json:"skip"`
		SessionID      string   `json:"session_id"`
		Subject        string   `json:"subject"`
		RequestedScope []string `json:"requested_scope"`
		RequestURL     string   `json:"request_url"`
		Client         struct { // concerned relying party
			ID        string `json:"client_id"`
			Name      string `json:"client_name"`
			LogoURI   string `json:"logo_uri"`
			TosURI    string `json:"tos_uri"`
			PolicyURI string `json:"policy_uri"`
		} `json:"client"`
		OIDCContext struct { // OIDC context of the current request
			ACRValues oidc.ClassRefs `json:"acr_values"`
			LoginHint string         `json:"login_hint"`
		} `json:"oidc_context"`
	}{}
	// query parameters
	params := url.Values{}
	params.Add("login_challenge", loginChallenge)

	// 2. perform the request
	logCtx := login.Context{}
	err := h.adminJSONRester.Get(ctx, "/oauth2/auth/requests/login", params, &hydraLogReq)
	if err != nil {
		if merr.IsANotFound(err) {
			return logCtx, merr.From(err).Add("challenge", merr.DVNotFound)
		}
		return logCtx, err
	}

	// 3. fill domain model using the DTO
	logCtx.Challenge = hydraLogReq.Challenge
	logCtx.Skip = hydraLogReq.Skip
	logCtx.SessionID = hydraLogReq.SessionID
	logCtx.Subject = hydraLogReq.Subject
	logCtx.RequestedScope = hydraLogReq.RequestedScope
	logCtx.Client.ID = hydraLogReq.Client.ID
	logCtx.Client.Name = hydraLogReq.Client.Name
	logCtx.RequestURL = hydraLogReq.RequestURL
	if hydraLogReq.Client.LogoURI != "" {
		logCtx.Client.LogoURL = null.StringFrom(hydraLogReq.Client.LogoURI)
	}
	if hydraLogReq.Client.TosURI != "" {
		logCtx.Client.TosURL = null.StringFrom(hydraLogReq.Client.TosURI)
	}
	if hydraLogReq.Client.PolicyURI != "" {
		logCtx.Client.PolicyURL = null.StringFrom(hydraLogReq.Client.PolicyURI)
	}
	// we must init ourselves the context which is a map
	// in most of other cases it is automatically initiated by the json unmarshaler
	logCtx.OIDCContext = oidc.NewContext()
	logCtx.OIDCContext.SetACRValues(hydraLogReq.OIDCContext.ACRValues)
	logCtx.OIDCContext.SetLoginHint(hydraLogReq.OIDCContext.LoginHint)
	return logCtx, nil
}

// Login user to hydra
func (h HydraHTTP) Login(ctx context.Context, loginChallenge string, acceptance login.Acceptance) (string, error) {
	redirect := struct {
		To string `json:"redirect_to"`
	}{}
	params := url.Values{}
	params.Add("login_challenge", loginChallenge)
	err := h.adminJSONRester.Put(ctx, "/oauth2/auth/requests/login/accept", params, acceptance, &redirect)
	if err != nil {
		if merr.IsANotFound(err) {
			return "", merr.From(err).Add("challenge", merr.DVNotFound)
		}
		return "", err
	}
	return redirect.To, nil
}

// GetConsentContext from hydra
func (h *HydraHTTP) GetConsentContext(ctx context.Context, consentChallenge string) (consent.Context, error) {
	consentCtx := consent.Context{}
	params := url.Values{}
	params.Add("consent_challenge", consentChallenge)
	err := h.adminJSONRester.Get(ctx, "/oauth2/auth/requests/consent", params, &consentCtx)
	if err != nil {
		return consentCtx, err
	}
	return consentCtx, nil
}

// Consent user's scope to hydra
func (h *HydraHTTP) Consent(ctx context.Context, consentChallenge string, acceptance consent.Acceptance) (consent.Redirect, error) {
	redirect := consent.Redirect{}
	params := url.Values{}
	params.Add("consent_challenge", consentChallenge)
	err := h.adminJSONRester.Put(ctx, "/oauth2/auth/requests/consent/accept", params, acceptance, &redirect)
	if err != nil {
		return redirect, err
	}
	return redirect, nil
}

// DeleteSession authentication for a subject
func (h *HydraHTTP) DeleteSession(ctx context.Context, subject string) error {
	route := fmt.Sprintf(
		"/oauth2/auth/sessions/login?subject=%s",
		url.PathEscape(subject),
	)
	return h.adminFormRester.Delete(ctx, route, nil)
}

// RevokeToken ...
func (h *HydraHTTP) RevokeToken(ctx context.Context, accessToken string) error {
	params := url.Values{}
	params.Add("token", accessToken)
	return h.protectedPublicFormRester.Post(ctx, "/oauth2/revoke", nil, params, nil)
}

// GetConsentSessions for a given Identity
func (h *HydraHTTP) GetConsentSessions(ctx context.Context, identityID string) ([]consent.Session, error) {
	consents := []consent.Session{}
	params := url.Values{}
	params.Add("subject", identityID)
	if err := h.adminJSONRester.Get(ctx, "/oauth2/auth/sessions/consent", params, &consents); err != nil {
		return nil, err
	}
	return consents, nil
}

// UserInfo ...
func (h *HydraHTTP) GetUserInfo(ctx context.Context, token string) (*userinfo.UserInfo, error) {
	userInfo := userinfo.UserInfo{}
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token
	ctxWithAuth := context.WithValue(ctx, rester.HeadersContextKey, headers)
	if err := h.publicJSONRester.Get(ctxWithAuth, "/userinfo", nil, &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

// CreateClient on hydra service
func (h *HydraHTTP) CreateClient(ctx context.Context, cli *Client) error {
	route := "/clients"
	return h.adminJSONRester.Post(ctx, route, nil, cli, nil)
}

// GetClient from hydra service using its id
func (h *HydraHTTP) GetClient(ctx context.Context, id string) (Client, error) {
	route := fmt.Sprintf("/clients/%s", id)
	cli := Client{}
	err := h.adminJSONRester.Get(ctx, route, nil, &cli)
	// Unauthorized are NotFound for this endpoint... https://www.ory.sh/hydra/docs/reference/api/#responses-10
	if merr.IsUnauthorized(err) {
		return cli, merr.NotFound().Desc(err.Error())
	}
	return cli, err
}

// UpdateClient on hydra service
func (h *HydraHTTP) UpdateClient(ctx context.Context, cli *Client) error {
	route := fmt.Sprintf("/clients/%s", cli.ID)
	return h.adminJSONRester.Put(ctx, route, nil, cli, cli)
}
