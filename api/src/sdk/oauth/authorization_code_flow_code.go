package oauth

import (
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// AuthorizationCodeInput contains parameters for obtaining a code
// it is built from query parameters
type AuthorizationCodeInput struct {
	Scopes []string
	State  string
	Prompt string // Can only be "none" or empty
}

// RequestCode by redirecting the user's agent to the authorization server with a well built URL to request a code
// Required to be keeped for some unofficial/in-house clients.
func (acf *AuthorizationCodeFlow) RequestCode(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	// check state parameter - state should not be empty
	if state == "" {
		acf.redirectErr(w, merr.MissingParameter.String(), "state")
		return
	}

	// prompt parameter
	prompt := r.URL.Query().Get("prompt")

	// scope is allowed to be empty
	scope := r.URL.Query().Get("scope")

	// Get URL authorization code
	params := AuthorizationCodeInput{
		Scopes: fromSpacedSep(scope),
		State:  state,
		Prompt: prompt,
	}
	// add openid scope by default to retrieve an id_token
	if !containsString(params.Scopes, "openid") {
		params.Scopes = append(params.Scopes, "openid")
	}
	// add user scope by default since it will always be users using this flow
	if !containsString(params.Scopes, "user") {
		params.Scopes = append(params.Scopes, "user")
	}

	// init config and start the flow
	config := oauth2.Config{
		ClientID:    acf.clientID,
		RedirectURL: acf.redirectCodeURL,
		Endpoint: oauth2.Endpoint{
			AuthURL: acf.codeURL,
		},
		Scopes: params.Scopes,
	}

	redirectURL, _ := url.Parse(config.AuthCodeURL(params.State))

	// handle prompt params ourselves - not handled by oauth2 package AuthCodeOption
	if params.Prompt == "none" {
		redirectURL.Query().Add("prompt", params.Prompt)
	}

	// Redirect request
	w.Header().Set("Location", redirectURL.String())
	w.WriteHeader(http.StatusFound)
}
