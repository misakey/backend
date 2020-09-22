package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// TokenRequest contains parameters for exchanging a code against a token
// it is built from query parameters
type TokenRequest struct {
	Code         string
	Scopes       []string
	State        string
	CodeVerifier string
}

// TokenError represents the body error returned by the authorization server following https://tools.ietf.org/html/rfc6749#section-5.2
type TokenError struct {
	Code  string `json:"error"`
	Desc  string `json:"error_description"`
	Debug string `json:"error_debug"`
}

// TokenResponse is the received structure after a successful token exchange
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// ExchangeToken using an authorization code then redirect the user agent with information related to operation's success or failure
func (acf *AuthorizationCodeFlow) ExchangeToken(w http.ResponseWriter, r *http.Request) {
	// if an error has been transmitted, consider it
	authErr := r.URL.Query().Get("error")
	if authErr != "" {
		errCode := merror.Code(authErr)
		errDesc := r.URL.Query().Get("error_debug")
		if errDesc == "" {
			errDesc = r.URL.Query().Get("error_hint")
		}
		acf.redirectErr(w, errCode.String(), errDesc)
		return
	}

	// check code parameter - it shall not be empty: https://tools.ietf.org/html/rfc6749#section-4.1.2
	code := r.URL.Query().Get("code")
	if code == "" {
		acf.redirectErr(w, merror.MissingParameter.String(), "code")
		return
	}

	// check state parameter - it shall not be empty: https://tools.ietf.org/html/rfc6749#section-4.1.2
	state := r.URL.Query().Get("state")
	if len(state) == 0 {
		acf.redirectErr(w, merror.MissingParameter.String(), "state")
		return
	}

	// check code verifier (cf https://tools.ietf.org/html/rfc7636#section-4.6) parameter - it can be empty
	// TODO: we target to make mandatory PKCE with the use of the autorization code flow with confidential client
	codeVerifier := r.URL.Query().Get("code_verifier")

	// get scopes
	scopes := fromSpacedSep(r.URL.Query().Get("scope"))

	// ensure `openid` scope is part of the authorization process
	if !containsString(scopes, "openid") {
		scopes = append(scopes, "openid")
	}
	// ensure `user` scope is part of the authorization process
	if !containsString(scopes, "user") {
		scopes = append(scopes, "user")
	}

	// get access token URL to redirect to
	params := TokenRequest{
		Code:         code,
		CodeVerifier: codeVerifier,
		State:        state,
		Scopes:       scopes,
	}
	redirectURL, err := acf.getURLWithAccessToken(r.Context(), params)
	if err != nil {
		tokErr := TokenError{Code: string(merror.UnauthorizedCode), Desc: err.Error()}
		acf.redirectErr(w, tokErr.Code, fmt.Sprintf("%s (%s)", tokErr.Desc, tokErr.Debug))
		return
	}

	// redirect user's agent to the final url
	w.Header().Set("Location", redirectURL.String())
	w.WriteHeader(http.StatusFound)
}

// getURLWithAccessToken by performing an authenticated http request using the autorization code on the token endpoint
// then build the relying party redirection URL
func (acf *AuthorizationCodeFlow) getURLWithAccessToken(ctx context.Context, tokenRequest TokenRequest) (*url.URL, error) {
	params := url.Values{}
	// prepare parameter for exchange the code against the token
	params.Add("grant_type", "authorization_code")
	params.Add("code", tokenRequest.Code)
	params.Add("redirect_uri", acf.redirectCodeURL)
	params.Add("scope", strings.Join(tokenRequest.Scopes, " "))
	// code verifier is optional
	if tokenRequest.CodeVerifier != "" {
		params.Add("code_verifier", tokenRequest.CodeVerifier)
	}

	tokenResp := TokenResponse{}
	if err := acf.tokenRester.Post(ctx, "/oauth2/token", nil, params, &tokenResp); err != nil {
		return nil, err
	}

	// add token data as fragment to final URL then return it
	// fragment parameters tends toward compliancy with https://tools.ietf.org/html/rfc6749#section-5.1
	urlParams := url.Values{}
	urlParams.Add("access_token", tokenResp.AccessToken)
	urlParams.Add("token_type", tokenResp.TokenType)
	urlParams.Add("id_token", tokenResp.IDToken)
	urlParams.Add("expires_in", strconv.Itoa(tokenResp.ExpiresIn))
	// compute expiry time
	expiry := time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn)).Format(time.RFC3339)
	urlParams.Add("expiry", expiry)
	urlParams.Add("scope", tokenResp.Scope)
	urlParams.Add("state", tokenRequest.State)
	return url.Parse(fmt.Sprintf("%s#%s", acf.redirectTokenURL, urlParams.Encode()))
}
