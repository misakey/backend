package consent

import (
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
)

type Session struct {
	GrantScope     []string `json:"grant_scope"`
	ConsentRequest struct {
		Client struct {
			ID string `json:"client_id"`
		} `json:"client"`
	} `json:"consent_request"`
}

// Context bears internal data about current user consent request
type Context struct {
	Subject        string        `json:"subject"`
	Challenge      string        `json:"challenge"`
	Skip           bool          `json:"skip"`
	ACR            string        `json:"acr"`
	RequestedScope []string      `json:"requested_scope"`
	AuthnContext   authn.Context `json:"context"`
	Client         struct {
		ID      string      `json:"client_id"`
		Name    string      `json:"name"`
		LogoURL null.String `json:"logo_uri"`
	} `json:"client"`
}

// Acceptance contains data about the user consent approval
type Acceptance struct {
	GrantScope  []string `json:"grant_scope"`
	Remember    bool     `json:"remember"`
	RememberFor int      `json:"remember_for"`
	Session     struct {
		IDTokenClaims struct {
			// extra self-contained claims on ID Token
			Scope string   `json:"sco"`
			Email string   `json:"email"`
			AMR   []string `json:"amr"`
		} `json:"id_token"`
		AccessTokenClaims struct {
			// extra instropection claims on Access Token
			ACR string `json:"acr"`
		} `json:"access_token"`
	} `json:"session"`
}

// Redirect information for the user's agent
type Redirect struct {
	To string `json:"redirect_to"`
}
