package login

import (
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
)

// Context bears internal data about current user authentication request
type Context struct {
	Challenge      string   `json:"challenge"`
	Skip           bool     `json:"skip"`
	Subject        string   `json:"subject"`
	RequestedScope []string `json:"requested_scope"`
	Client         struct { // concerned relying party
		ID      string      `json:"id"`
		Name    string      `json:"name"`
		LogoURL null.String `json:"logo_uri"`
	} `json:"client"`
	OIDCContext struct { // OIDC context of the current request
		ACRValues []string `json:"acr_values"`
		LoginHint string   `json:"login_hint"`
	} `json:"oidc_context"`
}

// Redirect information for the user's agent
type Redirect struct {
	To string `json:"redirect_to"`
}

// Acceptance contains data about the user authentication approval
type Acceptance struct {
	Subject     string        `json:"subject"`
	ACR         string        `json:"acr"`
	Remember    bool          `json:"remember"`
	RememberFor int           `json:"remember_for"`
	Context     authn.Context `json:"context"`
}

// // LogoutRequest contains the id of the user
// type LogoutRequest struct {
// 	UserID string `json:"user_id" validate:"required"`
// }
