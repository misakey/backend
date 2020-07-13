package login

import (
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
)

// Context bears internal data about current user authentication request
type Context struct {
	Challenge      string
	Subject        string
	RequestedScope []string

	Client struct { // involved relying party
		ID        string
		Name      string
		LogoURL   null.String
		TosURL    null.String
		PolicyURL null.String
	}

	RequestURL string

	OIDCContext struct { // OIDC context of the current request
		ACRValues authn.ClassRefs
		AMRs      authn.MethodRefs
		LoginHint string
	}

	// login session
	Skip      bool
	SessionID string
}

// Redirect information for the user's agent
type Redirect struct {
	To string `json:"redirect_to"`
}

// Acceptance contains data about the user authentication approval
type Acceptance struct {
	Subject     string         `json:"subject"`
	ACR         authn.ClassRef `json:"acr"`
	Remember    bool           `json:"remember"`
	RememberFor int            `json:"remember_for"`
	Context     authn.Context  `json:"context"`
}

// // LogoutRequest contains the id of the user
// type LogoutRequest struct {
// 	UserID string `json:"user_id" validate:"required"`
// }
