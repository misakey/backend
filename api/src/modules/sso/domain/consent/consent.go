package consent

import (
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/oidc"
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
	// oidc
	RequestURL string `json:"request_url"`
	Subject    string `json:"subject"`

	// consent
	Challenge      string   `json:"challenge"`
	Skip           bool     `json:"skip"`
	RequestedScope []string `json:"requested_scope"`

	// authentication during the login flow
	ACR            oidc.ClassRef `json:"acr"`
	OIDCContext    oidc.Context  `json:"context"`
	LoginSessionID string        `json:"login_session_id"`

	// involved client
	Client struct {
		ID        string      `json:"client_id"`
		Name      string      `json:"name"`
		LogoURL   null.String `json:"logo_uri"`
		TosURL    null.String `json:"tos_uri"`
		PolicyURL null.String `json:"policy_uri"`
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
			Scope string          `json:"sco"`
			Email string          `json:"email"`
			AMR   oidc.MethodRefs `json:"amr"`
			// optional ID token claims field only for Misakey SSO Client
			MID null.String `json:"mid,omitempty"`
			AID null.String `json:"aid,omitempty"`
		} `json:"id_token"`
		AccessTokenClaims oidc.Context `json:"access_token"`
	} `json:"session"`
}

// Redirect information for the user's agent
type Redirect struct {
	To string `json:"redirect_to"`
}
